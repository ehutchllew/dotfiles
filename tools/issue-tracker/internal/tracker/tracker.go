// Package tracker implements feature reads, validation, document edits, and
// ticket state operations on top of a storage backend.
package tracker

import (
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/apperr"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/doc"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/store"
)

// Store is the storage a backend provides.
type Store interface {
	ListFeatures() ([]string, error)
	ListTicketFiles(feature string) ([]store.TicketFile, []doc.Problem, error)
	Exists(rel string) (bool, error)
	Read(rel string) ([]byte, error)
	Write(rel string, data []byte, create bool) error
	Lock(feature string, exclusive bool) (func(), error)
	ResolveCommit(ref string) (string, error)
	IsIgnored(rel string) (ignored, ok bool)
}

// Tracker runs commands against one repository's storage.
type Tracker struct {
	Store Store
}

// SpecDoc is a spec with its storage path and hash.
type SpecDoc struct {
	*doc.Spec
	Path string
	Hash string
	Raw  []byte
}

// TicketDoc is a ticket with its storage path and hash.
type TicketDoc struct {
	*doc.Ticket
	Path       string
	Hash       string
	Raw        []byte
	FileNumber int
}

// Feature is a consistent snapshot of one spec and its tickets.
type Feature struct {
	ID           string
	Spec         *SpecDoc
	Tickets      []*TicketDoc
	FileProblems []doc.Problem
}

func (tr *Tracker) lock(id string, exclusive bool) (func(), error) {
	if err := store.CheckFeatureID(id); err != nil {
		return nil, err
	}
	return tr.Store.Lock(id, exclusive)
}

// load reads a feature. The caller must hold the feature lock.
func (tr *Tracker) load(id string) (*Feature, error) {
	f := &Feature{ID: id}
	b, err := tr.Store.Read(store.SpecPath(id))
	if e, ok := err.(*apperr.Error); ok && e.Code == "not_found" {
		return nil, apperr.New("feature_not_found", "feature %q has no spec at %s", id, store.SpecPath(id))
	}
	if err != nil {
		return nil, err
	}
	f.Spec = &SpecDoc{Spec: doc.ParseSpec(b), Path: store.SpecPath(id), Hash: doc.Hash(b), Raw: b}
	files, problems, err := tr.Store.ListTicketFiles(id)
	if err != nil {
		return nil, err
	}
	f.FileProblems = problems
	for _, file := range files {
		b, err := tr.Store.Read(file.Path)
		if err != nil {
			return nil, err
		}
		t := &TicketDoc{Ticket: doc.ParseTicket(b), Path: file.Path, Hash: doc.Hash(b), Raw: b, FileNumber: file.Number}
		f.Tickets = append(f.Tickets, t)
	}
	return f, nil
}

// Ticket finds one ticket by number.
func (f *Feature) Ticket(n int) (*TicketDoc, error) {
	var found []*TicketDoc
	for _, t := range f.Tickets {
		if t.FileNumber == n {
			found = append(found, t)
		}
	}
	switch len(found) {
	case 0:
		return nil, apperr.New("ticket_not_found", "feature %q has no ticket %s", f.ID, doc.TicketID(n))
	case 1:
		return found[0], nil
	default:
		paths := make([]string, len(found))
		for i, t := range found {
			paths[i] = t.Path
		}
		return nil, apperr.New("ambiguous_ticket", "feature %q has %d files for ticket %s", f.ID, len(found), doc.TicketID(n)).With("paths", paths)
	}
}

// byNumber maps ticket numbers to tickets, skipping ambiguous numbers.
func (f *Feature) byNumber() map[int]*TicketDoc {
	m := map[int]*TicketDoc{}
	dup := map[int]bool{}
	for _, t := range f.Tickets {
		if _, ok := m[t.FileNumber]; ok {
			dup[t.FileNumber] = true
		}
		m[t.FileNumber] = t
	}
	for n := range dup {
		delete(m, n)
	}
	return m
}

// CurrentAcceptance returns the highest-numbered acceptance ticket that
// isn't cancelled, or nil.
func (f *Feature) CurrentAcceptance() *TicketDoc {
	var cur *TicketDoc
	for _, t := range f.Tickets {
		if t.Kind == doc.KindAcceptance && t.Status != doc.StatusCancelled {
			if cur == nil || t.FileNumber > cur.FileNumber {
				cur = t
			}
		}
	}
	return cur
}

// Ready reports whether a ticket is open or in progress and every blocker
// is done.
func (f *Feature) Ready(t *TicketDoc) bool {
	if len(t.Problems) > 0 || (t.Status != doc.StatusOpen && t.Status != doc.StatusInProgress) {
		return false
	}
	return len(f.unfinishedBlockers(t)) == 0
}

func (f *Feature) unfinishedBlockers(t *TicketDoc) []string {
	idx := f.byNumber()
	var out []string
	for _, n := range t.BlockedBy {
		if b, ok := idx[n]; !ok || b.Status != doc.StatusDone {
			out = append(out, doc.TicketID(n))
		}
	}
	return out
}

func (f *Feature) nextNumber() int {
	max := 0
	for _, t := range f.Tickets {
		if t.FileNumber > max {
			max = t.FileNumber
		}
	}
	return max + 1
}

func specPath(id string) string { return store.SpecPath(id) }

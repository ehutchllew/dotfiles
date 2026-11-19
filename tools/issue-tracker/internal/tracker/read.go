package tracker

import (
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/apperr"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/doc"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/store"
)

// FeatureEntry is one row of `feature list`.
type FeatureEntry struct {
	Feature string `json:"feature"`
	Title   string `json:"title,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ListFeatures returns every feature with a spec.
func (tr *Tracker) ListFeatures() ([]FeatureEntry, error) {
	ids, err := tr.Store.ListFeatures()
	if err != nil {
		return nil, err
	}
	out := []FeatureEntry{}
	for _, id := range ids {
		e := FeatureEntry{Feature: id}
		unlock, err := tr.lock(id, false)
		if err != nil {
			return nil, err
		}
		b, err := tr.Store.Read(store.SpecPath(id))
		unlock()
		if err != nil {
			e.Error = err.Error()
		} else {
			e.Title = doc.ParseSpec(b).Title
		}
		out = append(out, e)
	}
	return out, nil
}

// TicketSummary is a ticket's metadata without its body.
type TicketSummary struct {
	Ticket         string              `json:"ticket"`
	Title          string              `json:"title"`
	Kind           string              `json:"kind"`
	Phase          string              `json:"phase,omitempty"`
	Status         string              `json:"status"`
	Attempt        *int                `json:"attempt,omitempty"`
	BlockedBy      []string            `json:"blocked_by,omitempty"`
	Covers         []string            `json:"covers,omitempty"`
	SpecRevision   int                 `json:"spec_revision,omitempty"`
	Base           string              `json:"base,omitempty"`
	Target         string              `json:"target,omitempty"`
	Ready          bool                `json:"ready"`
	Hash           string              `json:"hash"`
	Path           string              `json:"path,omitempty"`
	LastComment    string              `json:"last_comment,omitempty"`
	PendingHandoff *doc.PendingHandoff `json:"pending_handoff,omitempty"`
}

func (f *Feature) summary(t *TicketDoc, withPath bool) TicketSummary {
	s := TicketSummary{
		Ticket:         doc.TicketID(t.FileNumber),
		Title:          t.Title,
		Kind:           t.Kind,
		Phase:          t.Phase,
		Status:         t.Status,
		Covers:         t.Covers,
		SpecRevision:   t.SpecRevision,
		Base:           t.Base,
		Target:         t.Target,
		PendingHandoff: t.PendingHandoff,
		Ready:          f.Ready(t),
		Hash:           t.Hash,
	}
	if t.Kind == doc.KindImplementation {
		a := t.Attempt
		s.Attempt = &a
	}
	for _, n := range t.BlockedBy {
		s.BlockedBy = append(s.BlockedBy, doc.TicketID(n))
	}
	if withPath {
		s.Path = t.Path
	}
	return s
}

// ListTickets returns ticket summaries.
func (tr *Tracker) ListTickets(id string) (map[string]any, error) {
	unlock, err := tr.lock(id, false)
	if err != nil {
		return nil, err
	}
	defer unlock()
	f, err := tr.load(id)
	if err != nil {
		return nil, err
	}
	tickets := []TicketSummary{}
	for _, t := range f.Tickets {
		tickets = append(tickets, f.summary(t, true))
	}
	out := map[string]any{"feature": id, "tickets": tickets}
	if cur := f.CurrentAcceptance(); cur != nil {
		out["current_acceptance"] = doc.TicketID(cur.FileNumber)
	}
	return out, nil
}

// Finding is an unresolved not-blocking finding.
type Finding struct {
	Ticket  string `json:"ticket"`
	Finding string `json:"finding"`
}

// StatusReport is the compact orchestration state.
type StatusReport struct {
	Feature               string          `json:"feature"`
	Title                 string          `json:"title"`
	Revision              int             `json:"revision"`
	AttemptLimit          int             `json:"attempt_limit"`
	SpecHash              string          `json:"spec_hash"`
	Valid                 bool            `json:"valid"`
	Errors                []doc.Problem   `json:"errors,omitempty"`
	CurrentAcceptance     string          `json:"current_acceptance,omitempty"`
	EarlierAcceptanceBase string          `json:"earlier_acceptance_base,omitempty"`
	Tickets               []TicketSummary `json:"tickets"`
	NotBlocking           []Finding       `json:"not_blocking,omitempty"`
}

// Status returns the compact state Ralph orchestrates from.
func (tr *Tracker) Status(id string) (*StatusReport, error) {
	unlock, err := tr.lock(id, false)
	if err != nil {
		return nil, err
	}
	defer unlock()
	f, err := tr.load(id)
	if err != nil {
		return nil, err
	}
	r := &StatusReport{
		Feature:      id,
		Title:        f.Spec.Title,
		Revision:     f.Spec.Revision,
		AttemptLimit: f.Spec.AttemptLimit,
		SpecHash:     f.Spec.Hash,
		Errors:       f.Validate(true),
		Tickets:      []TicketSummary{},
	}
	r.Valid = len(r.Errors) == 0
	cur := f.CurrentAcceptance()
	if cur != nil {
		r.CurrentAcceptance = doc.TicketID(cur.FileNumber)
		var earlier *TicketDoc
		for _, t := range f.Tickets {
			if t.Kind == doc.KindAcceptance && t.FileNumber < cur.FileNumber && (earlier == nil || t.FileNumber > earlier.FileNumber) {
				earlier = t
			}
		}
		if earlier != nil {
			r.EarlierAcceptanceBase = earlier.Base
		}
	}
	for _, t := range f.Tickets {
		s := f.summary(t, false)
		if t.Status == doc.StatusStopped && len(t.Comments) > 0 {
			s.LastComment = t.CommentText(len(t.Comments) - 1)
		}
		r.Tickets = append(r.Tickets, s)
		for _, nb := range t.UnresolvedNotBlocking() {
			r.NotBlocking = append(r.NotBlocking, Finding{Ticket: doc.TicketID(t.FileNumber), Finding: nb})
		}
	}
	return r, nil
}

// Validate runs validation and returns an error when it fails.
func (tr *Tracker) Validate(id string, full bool) (map[string]any, error) {
	unlock, err := tr.lock(id, false)
	if err != nil {
		return nil, err
	}
	defer unlock()
	f, err := tr.load(id)
	if err != nil {
		return nil, err
	}
	problems := f.Validate(full)
	mode := "full"
	if !full {
		mode = "draft"
	}
	if len(problems) > 0 {
		return nil, apperr.New("validation_failed", "feature %q has %d validation errors", id, len(problems)).With("errors", problems).With("mode", mode)
	}
	return map[string]any{"feature": id, "valid": true, "mode": mode}, nil
}

// DocResult is the response of `spec get` and `ticket get`.
type DocResult struct {
	Feature         string          `json:"feature"`
	Ticket          string          `json:"ticket,omitempty"`
	Hash            string          `json:"hash"`
	Unchanged       bool            `json:"unchanged,omitempty"`
	Path            string          `json:"path,omitempty"`
	Revision        int             `json:"revision,omitempty"`
	AttemptLimit    int             `json:"attempt_limit,omitempty"`
	History         string          `json:"history,omitempty"`
	OmittedComments int             `json:"omitted_comments,omitempty"`
	Problems        []doc.Problem   `json:"problems,omitempty"`
	Criteria        []doc.Criterion `json:"criteria,omitempty"`
	Body            *string         `json:"body,omitempty"`
}

// SpecGet returns the spec, or only its hash when ifChanged matches.
func (tr *Tracker) SpecGet(id, ifChanged string) (*DocResult, error) {
	unlock, err := tr.lock(id, false)
	if err != nil {
		return nil, err
	}
	defer unlock()
	b, err := tr.Store.Read(store.SpecPath(id))
	if e, ok := err.(*apperr.Error); ok && e.Code == "not_found" {
		return nil, apperr.New("feature_not_found", "feature %q has no spec at %s", id, store.SpecPath(id))
	}
	if err != nil {
		return nil, err
	}
	out := &DocResult{Feature: id, Hash: doc.Hash(b)}
	if ifChanged != "" && ifChanged == out.Hash {
		out.Unchanged = true
		return out, nil
	}
	sp := doc.ParseSpec(b)
	body := string(b)
	out.Path = store.SpecPath(id)
	out.Revision = sp.Revision
	out.AttemptLimit = sp.AttemptLimit
	out.Problems = sp.Problems
	out.Body = &body
	return out, nil
}

// TicketGet returns a ticket, or only its hash when ifChanged matches. With
// full unset, comments before the last `revise` are left out.
func (tr *Tracker) TicketGet(id string, n int, ifChanged string, full bool) (*DocResult, error) {
	unlock, err := tr.lock(id, false)
	if err != nil {
		return nil, err
	}
	defer unlock()
	f, err := tr.load(id)
	if err != nil {
		return nil, err
	}
	t, err := f.Ticket(n)
	if err != nil {
		return nil, err
	}
	out := &DocResult{Feature: id, Ticket: doc.TicketID(n), Hash: t.Hash}
	if ifChanged != "" && ifChanged == t.Hash {
		out.Unchanged = true
		return out, nil
	}
	out.Path = t.Path
	body := string(t.Raw)
	if full {
		out.History = "full"
	} else if trimmed, omitted := t.CurrentHistory(); omitted > 0 {
		body = string(trimmed)
		out.History = "current"
		out.OmittedComments = omitted
	}
	out.Problems = t.Problems
	out.Criteria = t.Criteria
	out.Body = &body
	return out, nil
}

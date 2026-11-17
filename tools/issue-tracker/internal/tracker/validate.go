package tracker

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/apperr"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/doc"
)

// Validate runs the per-document checks and the checks that hold while a
// feature is being drafted. With full set, it also runs the whole-feature
// checks that need every ticket to exist.
func (f *Feature) Validate(full bool) []doc.Problem {
	var out []doc.Problem
	add := func(code, ticket, format string, args ...any) {
		out = append(out, doc.Problem{Code: code, Ticket: ticket, Message: fmt.Sprintf(format, args...)})
	}

	out = append(out, f.Spec.Problems...)
	out = append(out, f.FileProblems...)

	counts := map[int]int{}
	for _, t := range f.Tickets {
		counts[t.FileNumber]++
	}
	for _, t := range f.Tickets {
		id := doc.TicketID(t.FileNumber)
		if counts[t.FileNumber] > 1 {
			add("ambiguous_ticket", id, "%s shares ticket number %s with another file", t.Path, id)
		}
		for _, p := range t.Problems {
			p.Ticket = id
			out = append(out, p)
		}
		if t.Number != 0 && t.Number != t.FileNumber {
			add("number_mismatch", id, "the title says ticket %s but the file name says %s", doc.TicketID(t.Number), id)
		}
	}

	idx := f.byNumber()
	for _, t := range f.Tickets {
		id := doc.TicketID(t.FileNumber)
		for _, n := range t.BlockedBy {
			b, ok := idx[n]
			switch {
			case n == t.FileNumber:
				add("invalid_blocker", id, "the ticket is blocked by itself")
			case !ok:
				add("invalid_blocker", id, "`Blocked by:` names ticket %s, which doesn't exist", doc.TicketID(n))
			case b.Status == doc.StatusCancelled:
				add("invalid_blocker", id, "`Blocked by:` names ticket %s, which is cancelled", doc.TicketID(n))
			}
		}
	}
	for _, cycle := range f.cycles(nil) {
		add("blocker_cycle", cycle[0], "blockers form a cycle: %s", strings.Join(cycle, " -> "))
	}
	for _, t := range f.Tickets {
		if t.Kind == doc.KindAcceptance && t.Status == doc.StatusInProgress {
			if open := f.unfinishedBlockers(t); len(open) > 0 {
				add("blockers_not_done", doc.TicketID(t.FileNumber), "the acceptance ticket is in progress but blockers %s aren't done", strings.Join(open, ", "))
			}
		}
	}
	for _, t := range f.Tickets {
		out = append(out, f.coverProblems(t, t.Covers)...)
	}

	if !full {
		return out
	}

	active := f.Spec.ActiveStories()
	covered := map[string]bool{}
	for _, t := range f.Tickets {
		if t.Kind == doc.KindImplementation && t.Status != doc.StatusCancelled {
			for _, s := range t.Covers {
				covered[s] = true
			}
		}
	}
	for _, s := range active {
		if !covered[s] {
			add("story_not_covered", "", "%s isn't covered by any implementation ticket that isn't cancelled", s)
		}
	}

	cur := f.CurrentAcceptance()
	if cur == nil {
		add("no_acceptance", "", "the feature has no acceptance ticket that isn't cancelled")
		return out
	}
	cid := doc.TicketID(cur.FileNumber)
	have := map[string]bool{}
	for _, s := range cur.Covers {
		have[s] = true
	}
	var missing []string
	for _, s := range active {
		if !have[s] {
			missing = append(missing, s)
		}
	}
	if len(missing) > 0 {
		add("acceptance_coverage", cid, "the current acceptance ticket doesn't cover %s", strings.Join(missing, ", "))
	}
	if cur.SpecRevision != f.Spec.Revision {
		add("stale_acceptance", cid, "the current acceptance ticket checks spec revision %d but the spec is at revision %d", cur.SpecRevision, f.Spec.Revision)
	}
	blockers := map[int]bool{}
	for _, n := range cur.BlockedBy {
		blockers[n] = true
	}
	var unblocked []string
	for _, t := range f.Tickets {
		if (t.Kind == doc.KindImplementation || t.Kind == doc.KindManual) && t.Status != doc.StatusCancelled && !blockers[t.FileNumber] {
			unblocked = append(unblocked, doc.TicketID(t.FileNumber))
		}
	}
	if len(unblocked) > 0 {
		add("acceptance_blockers", cid, "the current acceptance ticket isn't blocked by %s", strings.Join(unblocked, ", "))
	}
	return out
}

// coverProblems checks that every covered story exists. Unfinished tickets
// also may not cover withdrawn stories. Done and cancelled tickets keep the
// stories they had, because they can't be edited.
func (f *Feature) coverProblems(t *TicketDoc, covers []string) []doc.Problem {
	var out []doc.Problem
	id := doc.TicketID(t.FileNumber)
	for _, s := range covers {
		st, ok := f.Spec.Story(s)
		switch {
		case !ok:
			out = append(out, doc.Problem{Code: "unknown_story", Ticket: id, Message: fmt.Sprintf("`Covers:` names %s, which isn't in the spec", s)})
		case st.Withdrawn && !t.IsFinished():
			out = append(out, doc.Problem{Code: "withdrawn_story", Ticket: id, Message: fmt.Sprintf("`Covers:` names %s, which was withdrawn in revision %d", s, st.WithdrawnIn)})
		}
	}
	return out
}

// withTicket returns a copy of the feature with the ticket at path replaced
// by data, or added when no ticket has that path.
func (f *Feature) withTicket(path string, number int, data []byte) *Feature {
	c := *f
	c.Tickets = make([]*TicketDoc, 0, len(f.Tickets)+1)
	next := &TicketDoc{Ticket: doc.ParseTicket(data), Path: path, Hash: doc.Hash(data), Raw: data, FileNumber: number}
	replaced := false
	for _, t := range f.Tickets {
		if t.Path == path {
			c.Tickets = append(c.Tickets, next)
			replaced = true
		} else {
			c.Tickets = append(c.Tickets, t)
		}
	}
	if !replaced {
		c.Tickets = append(c.Tickets, next)
	}
	return &c
}

// checkNoRegressions rejects a write that gives a ticket a draft validation
// problem it didn't already have. Problems that existed before the write
// don't block it, so a broken feature can still be repaired step by step.
func checkNoRegressions(before, after *Feature) error {
	had := map[string]bool{}
	for _, p := range before.Validate(false) {
		had[p.Code+"|"+p.Ticket] = true
	}
	var added []doc.Problem
	for _, p := range after.Validate(false) {
		if !had[p.Code+"|"+p.Ticket] {
			added = append(added, p)
		}
	}
	if len(added) == 0 {
		return nil
	}
	msg := added[0].Message
	if added[0].Code == "blockers_not_done" {
		msg += "; revise the acceptance ticket before adding blockers"
	}
	return apperr.New("would_invalidate", "the change would break validation: %s", msg).With("problems", added)
}

// cycles returns each blocker cycle once. override replaces one ticket's
// blockers, so edits can be checked before they are written.
func (f *Feature) cycles(override map[int][]int) [][]string {
	edges := map[int][]int{}
	for _, t := range f.Tickets {
		edges[t.FileNumber] = t.BlockedBy
	}
	for n, b := range override {
		edges[n] = b
	}
	nodes := make([]int, 0, len(edges))
	for n := range edges {
		nodes = append(nodes, n)
	}
	sort.Ints(nodes)

	const (
		white = iota
		grey
		black
	)
	color := map[int]int{}
	var stack []int
	var found [][]string
	seen := map[string]bool{}
	var visit func(n int)
	visit = func(n int) {
		color[n] = grey
		stack = append(stack, n)
		for _, m := range edges[n] {
			if m == n {
				continue
			}
			switch color[m] {
			case white:
				if _, ok := edges[m]; ok {
					visit(m)
				}
			case grey:
				var cyc []int
				for i := len(stack) - 1; i >= 0; i-- {
					cyc = append([]int{stack[i]}, cyc...)
					if stack[i] == m {
						break
					}
				}
				key := cycleKey(cyc)
				if !seen[key] {
					seen[key] = true
					ids := make([]string, 0, len(cyc)+1)
					for _, c := range cyc {
						ids = append(ids, doc.TicketID(c))
					}
					found = append(found, append(ids, doc.TicketID(cyc[0])))
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[n] = black
	}
	for _, n := range nodes {
		if color[n] == white {
			visit(n)
		}
	}
	return found
}

func cycleKey(c []int) string {
	s := append([]int(nil), c...)
	sort.Ints(s)
	return fmt.Sprint(s)
}

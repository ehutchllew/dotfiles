package cli_test

import (
	"strings"
	"testing"
)

const co = "checkout"

type transition struct {
	name  string
	n     string
	done  []string // tickets set to done by hand first
	setup []string // metadata set by hand on ticket n
	op    string
	input string
	err   string
	check func(t *testing.T, r *repo, tk ticketSummary, w writeReply)
}

func wantState(phase, status string, attempts int) func(*testing.T, *repo, ticketSummary, writeReply) {
	return func(t *testing.T, r *repo, tk ticketSummary, w writeReply) {
		t.Helper()
		if tk.Phase != phase || tk.Status != status || attempt(tk) != attempts {
			t.Fatalf("got %s/%s attempt %d, want %s/%s attempt %d", tk.Phase, tk.Status, attempt(tk), phase, status, attempts)
		}
	}
}

func wantComment(prefix string) func(*testing.T, *repo, ticketSummary, writeReply) {
	return func(t *testing.T, r *repo, tk ticketSummary, w writeReply) {
		t.Helper()
		if c := r.lastComment(co, tk.Ticket); !strings.HasPrefix(c, prefix) {
			t.Fatalf("last comment %q doesn't start with %q", c, prefix)
		}
	}
}

func both(checks ...func(*testing.T, *repo, ticketSummary, writeReply)) func(*testing.T, *repo, ticketSummary, writeReply) {
	return func(t *testing.T, r *repo, tk ticketSummary, w writeReply) {
		t.Helper()
		for _, c := range checks {
			c(t, r, tk, w)
		}
	}
}

func commit(sha string) string { return `{"commit":"` + sha + `"}` }

func TestTransitions(t *testing.T) {
	inProgress := []string{"Status", "in-progress", "Attempt", "1", "Base commit", shaA}
	reviewOpen := []string{"Phase", "review", "Attempt", "1", "Base commit", shaA, "Target commit", shaB}
	reviewStopped := append(append([]string{}, reviewOpen...), "Status", "stopped")
	accInProgress := []string{"Status", "in-progress", "Base commit", shaA, "Target commit", shaB}
	verified := `"verification":["ok"]`

	cases := []transition{
		{name: "claim", n: "01", op: "claim", input: commit(shaA), check: func(t *testing.T, r *repo, tk ticketSummary, w writeReply) {
			if tk.Status != "in-progress" || attempt(tk) != 1 || tk.Base != shaA || w.CommitSubject != "tickets: claim 01" {
				t.Fatalf("%+v %+v", tk, w)
			}
		}},
		{name: "claim needs a commit", n: "01", op: "claim", input: `{}`, err: "invalid_input"},
		{name: "claim keeps an existing base", n: "01", setup: []string{"Attempt", "1", "Base commit", shaA}, op: "claim", input: commit(shaB), check: func(t *testing.T, r *repo, tk ticketSummary, w writeReply) {
			if tk.Base != shaA || attempt(tk) != 2 || w.Note == "" {
				t.Fatalf("%+v %+v", tk, w)
			}
		}},
		{name: "claim continues in-progress", n: "01", setup: inProgress, op: "claim", input: commit(shaB), check: func(t *testing.T, r *repo, tk ticketSummary, w writeReply) {
			if !w.Unchanged || attempt(tk) != 1 {
				t.Fatalf("%+v", w)
			}
		}},
		{name: "claim in review", n: "01", setup: reviewOpen, op: "claim", input: commit(shaA), err: "invalid_transition"},
		{name: "claim at the attempt limit", n: "01", setup: []string{"Attempt", "3", "Base commit", shaA}, op: "claim", input: commit(shaA), err: "attempt_limit_reached"},
		{name: "claim blocked", n: "02", op: "claim", input: commit(shaA), err: "not_ready"},
		{name: "claim once blockers are done", n: "02", done: []string{"01"}, op: "claim", input: commit(shaA), check: wantState("implement", "in-progress", 1)},
		{name: "claim manual", n: "03", op: "claim", input: commit(shaA), err: "invalid_transition"},
		{name: "claim with a short sha", n: "01", op: "claim", input: commit("abc123"), err: "invalid_commit"},

		{name: "hand off", n: "01", setup: inProgress, op: "hand-off", input: `{"commit":"` + shaB + `","criteria_met":["An empty order shows zero."],` + verified + `}`,
			check: both(wantState("review", "open", 1), wantComment("hand off (attempt 1):\n- Verification:\n  - ok"), func(t *testing.T, r *repo, tk ticketSummary, w writeReply) {
				if tk.Target != shaB || !strings.Contains(r.read(tk.Path), "- [x] An empty order shows zero.\n- [ ] The count") {
					t.Fatalf("%+v\n%s", tk, r.read(tk.Path))
				}
			})},
		{name: "hand off quoted criterion", n: "01", setup: inProgress, op: "hand-off", input: `{"commit":"` + shaB + `","criteria_met":["\"An empty order shows zero.\""],` + verified + `}`, check: wantState("review", "open", 1)},
		{name: "hand off unknown criterion", n: "01", setup: inProgress, op: "hand-off", input: `{"commit":"` + shaB + `","criteria_met":["Nope."],` + verified + `}`, err: "unknown_criterion"},
		{name: "hand off without verification", n: "01", setup: inProgress, op: "hand-off", input: commit(shaB), err: "invalid_input"},
		{name: "hand off while open", n: "01", op: "hand-off", input: `{"commit":"` + shaB + `",` + verified + `}`, err: "invalid_transition"},

		{name: "pass review", n: "01", setup: reviewOpen, op: "pass", input: `{` + verified + `}`, check: both(wantState("review", "done", 1), wantComment("pass (attempt 1):"))},
		{name: "pass in implement", n: "01", op: "pass", input: `{}`, err: "invalid_transition"},
		{name: "pass manual", n: "03", op: "pass", input: `{}`, err: "invalid_transition"},
		{name: "pass open acceptance", n: "04", op: "pass", input: `{}`, err: "invalid_transition"},
		{name: "pass acceptance", n: "04", done: []string{"01", "02", "03"}, setup: accInProgress, op: "pass", input: `{"all_criteria_met":true}`, check: func(t *testing.T, r *repo, tk ticketSummary, w writeReply) {
			if tk.Status != "done" || strings.Contains(r.read(tk.Path), "- [ ]") {
				t.Fatalf("%+v\n%s", tk, r.read(tk.Path))
			}
		}},
		{name: "pass acceptance with an open blocker", n: "04", done: []string{"01", "02"}, setup: accInProgress, op: "pass", input: `{}`, err: "not_ready"},

		{name: "send back", n: "01", setup: reviewOpen, op: "send-back", input: `{"findings":["Blocking (wrong): x."]}`, check: func(t *testing.T, r *repo, tk ticketSummary, w writeReply) {
			if tk.Phase != "implement" || tk.Status != "open" || attempt(tk) != 1 || tk.Target != shaB {
				t.Fatalf("%+v", tk)
			}
		}},
		{name: "send back without findings", n: "01", setup: reviewOpen, op: "send-back", input: `{` + verified + `}`, err: "invalid_input"},
		{name: "send back with a not-blocking finding", n: "01", setup: reviewOpen, op: "send-back", input: `{"findings":["Not blocking: x"]}`, err: "invalid_input"},
		{name: "send back from implement", n: "01", op: "send-back", input: `{"findings":["x"]}`, err: "invalid_transition"},

		{name: "set feature base", n: "04", op: "set-feature-base", input: commit(shaA), check: func(t *testing.T, r *repo, tk ticketSummary, w writeReply) {
			if tk.Base != shaA || r.lastComment(co, "04") != "" {
				t.Fatalf("%+v", tk)
			}
		}},
		{name: "set feature base twice", n: "04", setup: []string{"Base commit", shaA}, op: "set-feature-base", input: commit(shaB), err: "base_already_set"},
		{name: "set feature base on implementation", n: "01", op: "set-feature-base", input: commit(shaA), err: "invalid_transition"},

		{name: "start review", n: "04", done: []string{"01", "02", "03"}, setup: []string{"Base commit", shaA}, op: "start-review", input: commit(shaB), check: func(t *testing.T, r *repo, tk ticketSummary, w writeReply) {
			if tk.Status != "in-progress" || tk.Target != shaB {
				t.Fatalf("%+v", tk)
			}
		}},
		{name: "start review continues", n: "04", done: []string{"01", "02", "03"}, setup: accInProgress, op: "start-review", input: commit(shaC), check: func(t *testing.T, r *repo, tk ticketSummary, w writeReply) {
			if !w.Unchanged || tk.Target != shaB {
				t.Fatalf("%+v", w)
			}
		}},
		{name: "start review blocked", n: "04", done: []string{"01", "02"}, setup: []string{"Base commit", shaA}, op: "start-review", input: commit(shaB), err: "not_ready"},
		{name: "start review without a base", n: "04", done: []string{"01", "02", "03"}, op: "start-review", input: commit(shaB), err: "missing_base"},
		{name: "start review of an old revision", n: "04", done: []string{"01", "02", "03"}, setup: []string{"Base commit", shaA, "Spec revision", "2"}, op: "start-review", input: commit(shaB), err: "stale_acceptance"},

		{name: "stop keeps the phase", n: "01", setup: reviewOpen, op: "stop", input: `{"reason":"Verification blocked."}`, check: both(wantState("review", "stopped", 1), wantComment("stop (attempt 1):\n- Reason: Verification blocked."))},
		{name: "stop a manual ticket", n: "03", op: "stop", input: `{"reason":"No phone."}`, check: both(wantState("", "stopped", -1), wantComment("stop:\n- Reason: No phone."))},
		{name: "stop needs a reason", n: "01", op: "stop", input: `{}`, err: "invalid_input"},
		{name: "stop a done ticket", n: "01", setup: []string{"Status", "done"}, op: "stop", input: `{"reason":"x"}`, err: "invalid_transition"},

		{name: "resume", n: "01", setup: []string{"Status", "stopped"}, op: "resume", input: `{"reason":"Fixed it."}`, check: both(wantState("implement", "open", 0), wantComment("resume (attempt 0):\n- Resolution: Fixed it."))},
		{name: "resume an open ticket", n: "01", op: "resume", input: `{"reason":"x"}`, err: "invalid_transition"},

		{name: "reset attempts", n: "01", setup: []string{"Status", "stopped", "Attempt", "3", "Base commit", shaA}, op: "reset-attempts", input: `{"reason":"More budget."}`,
			check: both(wantState("implement", "open", 0), wantComment("reset attempts (attempt 3):"))},
		{name: "reset attempts in review", n: "01", setup: reviewStopped, op: "reset-attempts", input: `{"reason":"x"}`, err: "invalid_transition"},

		{name: "revise implementation in review", n: "01", setup: reviewStopped, op: "revise", input: `{"reason":"Spec changed."}`,
			check: func(t *testing.T, r *repo, tk ticketSummary, w writeReply) {
				if tk.Phase != "implement" || tk.Status != "open" || attempt(tk) != 0 || tk.Base != shaA || tk.Target != shaB {
					t.Fatalf("%+v", tk)
				}
			}},
		{name: "revise acceptance", n: "04", done: []string{"01", "02", "03"}, setup: accInProgress, op: "revise", input: `{"reason":"New blockers."}`, check: both(wantState("", "open", -1), wantComment("revise:"))},
		{name: "revise manual", n: "03", setup: []string{"Status", "stopped"}, op: "revise", input: `{"reason":"New steps."}`, check: both(wantState("", "open", -1), wantComment("revise:\n- Reason: New steps."))},
		{name: "revise a done ticket", n: "01", setup: []string{"Status", "done"}, op: "revise", input: `{"reason":"x"}`, err: "invalid_transition"},

		{name: "complete", n: "03", op: "complete", input: `{"evidence":["Pixel 8."]}`, check: func(t *testing.T, r *repo, tk ticketSummary, w writeReply) {
			if tk.Status != "done" || !strings.Contains(r.read(tk.Path), "- [x] The count is readable") {
				t.Fatalf("%+v", tk)
			}
		}},
		{name: "complete without evidence", n: "03", op: "complete", input: `{}`, err: "invalid_input"},
		{name: "complete implementation", n: "01", op: "complete", input: `{"evidence":["x"]}`, err: "invalid_transition"},

		{name: "cancel in review", n: "01", setup: reviewOpen, op: "cancel", input: `{"reason":"Dropped."}`, check: wantState("review", "cancelled", 1)},
		{name: "cancel a done ticket", n: "01", setup: []string{"Status", "done"}, op: "cancel", input: "", err: "invalid_transition"},

		{name: "comment on a done ticket", n: "01", setup: []string{"Status", "done"}, op: "comment", input: `{"notes":["See ticket 05."]}`, check: wantComment("comment (attempt 0):\n- Note: See ticket 05.")},
		{name: "empty comment", n: "01", op: "comment", input: `{}`, err: "invalid_input"},
		{name: "field another operation takes", n: "01", op: "claim", input: `{"commit":"` + shaA + `","reason":"x"}`, err: "invalid_input"},
		{name: "field nobody takes", n: "01", op: "stop", input: `{"reason":"x","status":"done"}`, err: "invalid_input"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := newRepo(t, co)
			for _, n := range c.done {
				r.setMeta(co, n, "Status", "done")
			}
			if len(c.setup) > 0 {
				r.setMeta(co, c.n, c.setup...)
			}
			if c.err != "" {
				r.opFails(c.err, co, c.n, c.op, c.input)
				return
			}
			w := r.op(co, c.n, c.op, c.input)
			tk := r.ticket(co, c.n)
			if w.Hash != tk.Hash {
				t.Fatalf("reply hash %s, file hash %s", w.Hash, tk.Hash)
			}
			if s := r.status(co); !validExceptCoverage(s) {
				t.Fatalf("the operation left problems: %+v", s.Errors)
			}
			c.check(t, r, tk, w)
		})
	}
}

// validExceptCoverage ignores whole-feature problems that a hand-made setup
// can cause, such as cancelling a ticket that covers a story.
func validExceptCoverage(s statusReply) bool {
	for _, e := range s.Errors {
		if e.Code != "story_not_covered" && e.Code != "acceptance_blockers" {
			return false
		}
	}
	return true
}

func TestStaleWrites(t *testing.T) {
	r := newRepo(t, co)
	old := r.ticket(co, "01").Hash
	r.setMeta(co, "01", "Covers", "US 1, US 2")
	e := r.fails("stale", "", "ticket", "claim", co, "01", "--if-match", old, "--commit", shaA)
	if e.Details["current_hash"] != r.ticket(co, "01").Hash {
		t.Fatalf("stale error lacks the current hash: %+v", e)
	}
	r.fails("missing_if_match", "", "ticket", "claim", co, "01", "--commit", shaA)
	r.fails("stale", `{"title":"x"}`, "ticket", "edit", co, "01", "--if-match", old, "--input", "-")

	spec := r.specHash(co)
	r.ok(nil, `{"attempt_limit":4}`, "spec", "edit", co, "--if-match", spec, "--input", "-")
	r.fails("stale", `{"attempt_limit":5}`, "spec", "edit", co, "--if-match", spec, "--input", "-")
}

// TestChecksUseCurrentDependencies covers writes whose target hash is
// unchanged but whose blockers or spec changed since the caller's read.
func TestChecksUseCurrentDependencies(t *testing.T) {
	t.Run("a blocker reopened", func(t *testing.T) {
		r := newRepo(t, co)
		r.setMeta(co, "01", "Status", "done")
		h := r.ticket(co, "02").Hash
		r.setMeta(co, "01", "Status", "open")
		r.fails("not_ready", "", "ticket", "claim", co, "02", "--if-match", h, "--commit", shaA)
	})
	t.Run("the attempt limit dropped", func(t *testing.T) {
		r := newRepo(t, co)
		r.setMeta(co, "01", "Attempt", "2", "Base commit", shaA)
		r.ok(nil, `{"attempt_limit":2}`, "spec", "edit", co, "--if-match", r.specHash(co), "--input", "-")
		r.opFails("attempt_limit_reached", co, "01", "claim", commit(shaA))
	})

	accepting := func(t *testing.T) *repo {
		r := newRepo(t, co)
		for _, n := range []string{"01", "02", "03"} {
			r.setMeta(co, n, "Status", "done")
		}
		r.op(co, "04", "set-feature-base", commit(shaA))
		r.op(co, "04", "start-review", commit(shaB))
		return r
	}
	t.Run("the spec was revised during the acceptance review", func(t *testing.T) {
		r := accepting(t)
		h := r.ticket(co, "04").Hash
		r.ok(nil, `{"change":"revision","approved":true,"history":"US 1 reworded.","edit_stories":{"US 1":"As a shopper, I want x, so that y"}}`,
			"spec", "edit", co, "--if-match", r.specHash(co), "--input", "-")
		r.fails("stale_acceptance", `{"all_criteria_met":true}`, "ticket", "pass", co, "04", "--if-match", h, "--input", "-")
		r.op(co, "04", "stop", `{"reason":"The spec changed during the review."}`)
	})
	t.Run("a newer acceptance ticket exists", func(t *testing.T) {
		r := accepting(t)
		h := r.ticket(co, "04").Hash
		r.ok(nil, `{"title":"Accept again","kind":"acceptance","blocked_by":["01","02","03"],"covers":["US 1","US 2"],
			"what_to_build":"Review the complete feature against the spec.","criteria":["Every story works."],"verification":"Run `+"`npm test`"+`."}`,
			"ticket", "create", co, "--input", "-")
		r.fails("not_current_acceptance", `{}`, "ticket", "pass", co, "04", "--if-match", h, "--input", "-")
		if s := r.status(co); s.CurrentAcceptance != "05" || s.EarlierAcceptanceBase != shaA {
			t.Fatalf("status: current %s, earlier base %s", s.CurrentAcceptance, s.EarlierAcceptanceBase)
		}
	})
}

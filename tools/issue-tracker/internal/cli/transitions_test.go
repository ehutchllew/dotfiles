package cli_test

import (
	"fmt"
	"strings"
	"testing"
)

const co = "checkout"

func commit(sha string) string { return `{"commit":"` + sha + `"}` }

// Hand-made starting states for ticket metadata.
var (
	inProgress    = []string{"Status", "in-progress", "Attempt", "1", "Base commit", shaA}
	reviewOpen    = []string{"Phase", "review", "Attempt", "1", "Base commit", shaA, "Target commit", shaB}
	reviewStopped = append(append([]string{}, reviewOpen...), "Status", "stopped")
	accInProgress = []string{"Status", "in-progress", "Base commit", shaA, "Target commit", shaB}
	allDone       = []string{"01", "02", "03"}
)

// setup returns a checkout feature with the done tickets marked done and
// meta set on ticket n, all by hand.
func setup(t *testing.T, done []string, n string, meta []string) *repo {
	t.Helper()
	r := newRepo(t, co)
	for _, d := range done {
		r.setMeta(co, d, "Status", "done")
	}
	if len(meta) > 0 {
		r.setMeta(co, n, meta...)
	}
	return r
}

// state formats a ticket's state as "phase/status attempt N" for
// implementation tickets and as the status for the other kinds.
func state(phase, status string, attempt *int) string {
	if attempt == nil {
		return status
	}
	return fmt.Sprintf("%s/%s attempt %d", phase, status, *attempt)
}

func TestTransitions(t *testing.T) {
	cases := []struct {
		name    string
		n       string
		done    []string
		meta    []string
		op      string
		input   string
		want    string // the state afterwards, as state formats it
		comment string // the start of the last comment, or "" for none
	}{
		{name: "claim", n: "01", op: "claim", input: commit(shaA), want: "implement/in-progress attempt 1"},
		{name: "claim once blockers are done", n: "02", done: []string{"01"}, op: "claim", input: commit(shaA), want: "implement/in-progress attempt 1"},
		{name: "claim a retry", n: "01", meta: []string{"Attempt", "1", "Base commit", shaA}, op: "claim", input: commit(shaB), want: "implement/in-progress attempt 2"},

		{name: "hand off", n: "01", meta: inProgress, op: "hand-off", input: `{"commit":"` + shaB + `","verification":["ok"]}`,
			want: "review/open attempt 1", comment: "implement:hand-off (attempt 1):\n- Verification: ok"},

		{name: "pass review", n: "01", meta: reviewOpen, op: "pass", input: `{"verification":["ok"]}`, want: "review/done attempt 1", comment: "review:pass (attempt 1):"},
		{name: "pass acceptance", n: "04", done: allDone, meta: accInProgress, op: "pass", input: `{"verification":["ok"]}`, want: "done", comment: "acceptance:pass:"},

		{name: "send back", n: "01", meta: reviewOpen, op: "send-back", input: `{"findings":["Blocking (wrong): x."]}`,
			want: "implement/open attempt 1", comment: "review:send-back (attempt 1):\n- Blocking (wrong): x."},

		{name: "set feature base", n: "04", op: "set-feature-base", input: commit(shaA), want: "open"},
		{name: "start review", n: "04", done: allDone, meta: []string{"Base commit", shaA}, op: "start-review", input: commit(shaB), want: "in-progress"},

		{name: "stop keeps the phase", n: "01", meta: reviewOpen, op: "stop", input: `{"reason":"Verification blocked."}`,
			want: "review/stopped attempt 1", comment: "review:stop (attempt 1):\n- Reason: Verification blocked."},
		{name: "stop a manual ticket", n: "03", op: "stop", input: `{"reason":"No phone."}`, want: "stopped", comment: "manual:stop:\n- Reason: No phone."},

		{name: "resume", n: "01", meta: []string{"Status", "stopped"}, op: "resume", input: `{"reason":"Fixed it."}`,
			want: "implement/open attempt 0", comment: "implement:resume (attempt 0):\n- Resolution: Fixed it."},

		{name: "reset attempts", n: "01", meta: []string{"Status", "stopped", "Attempt", "3", "Base commit", shaA}, op: "reset-attempts", input: `{"reason":"More budget."}`,
			want: "implement/open attempt 0", comment: "implement:reset-attempts (attempt 3):\n- Reason: More budget."},

		{name: "revise implementation in review", n: "01", meta: reviewStopped, op: "revise", input: `{"reason":"Spec changed."}`,
			want: "implement/open attempt 0", comment: "review:revise (attempt 1):\n- Reason: Spec changed."},
		{name: "revise acceptance", n: "04", done: allDone, meta: accInProgress, op: "revise", input: `{"reason":"New blockers."}`, want: "open", comment: "acceptance:revise:"},
		{name: "revise manual", n: "03", meta: []string{"Status", "stopped"}, op: "revise", input: `{"reason":"New steps."}`, want: "open", comment: "manual:revise:\n- Reason: New steps."},

		{name: "complete", n: "03", op: "complete", input: `{"evidence":["Pixel 8."]}`, want: "done", comment: "manual:complete:\n- Evidence:\n  - Pixel 8."},

		{name: "cancel in review", n: "01", meta: reviewOpen, op: "cancel", input: `{"reason":"Dropped."}`,
			want: "review/cancelled attempt 1", comment: "review:cancel (attempt 1):\n- Reason: Dropped."},

		{name: "comment on a done ticket", n: "01", meta: []string{"Status", "done"}, op: "comment", input: `{"notes":["See ticket 05."]}`,
			want: "implement/done attempt 0", comment: "implement:comment (attempt 0):\n- Note: See ticket 05."},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := setup(t, c.done, c.n, c.meta)
			w := r.op(co, c.n, c.op, c.input)
			tk := r.ticket(co, c.n)
			if w.Hash != tk.Hash {
				t.Fatalf("reply hash %s, file hash %s", w.Hash, tk.Hash)
			}
			if got := state(tk.Phase, tk.Status, tk.Attempt); got != c.want {
				t.Errorf("ticket is %s, want %s", got, c.want)
			}
			if got := state(w.Phase, w.Status, w.Attempt); got != c.want {
				t.Errorf("reply says %s, want %s", got, c.want)
			}
			last := r.lastComment(co, c.n)
			if c.comment == "" && last != "" || !strings.HasPrefix(last, c.comment) {
				t.Errorf("last comment %q, want it to start with %q", last, c.comment)
			}
			if s := r.status(co); !validExceptCoverage(s) {
				t.Errorf("the operation left problems: %+v", s.Errors)
			}
		})
	}
}

func TestTransitionsRejected(t *testing.T) {
	cases := []struct {
		name  string
		n     string
		done  []string
		meta  []string
		op    string
		input string
		exit  int
		err   string
	}{
		{name: "claim needs a commit", n: "01", op: "claim", input: `{}`, exit: 2, err: "invalid_input"},
		{name: "claim in review", n: "01", meta: reviewOpen, op: "claim", input: commit(shaA), exit: 1, err: "invalid_transition"},
		{name: "claim at the attempt limit", n: "01", meta: []string{"Attempt", "3", "Base commit", shaA}, op: "claim", input: commit(shaA), exit: 1, err: "attempt_limit_reached"},
		{name: "claim blocked", n: "02", op: "claim", input: commit(shaA), exit: 1, err: "not_ready"},
		{name: "claim manual", n: "03", op: "claim", input: commit(shaA), exit: 1, err: "invalid_transition"},
		{name: "claim with a short sha", n: "01", op: "claim", input: commit("abc123"), exit: 1, err: "invalid_commit"},

		{name: "hand off unknown criterion", n: "01", meta: inProgress, op: "hand-off", input: `{"commit":"` + shaB + `","criteria_met":["Nope."],"verification":["ok"]}`, exit: 2, err: "unknown_criterion"},
		{name: "hand off duplicate criterion", n: "01", meta: inProgress, op: "hand-off", input: `{"commit":"` + shaB + `","criteria_met":["C01","C01"],"verification":["ok"]}`, exit: 2, err: "duplicate_criterion"},
		{name: "send back duplicate criterion", n: "01", meta: reviewOpen, op: "send-back", input: `{"criteria_unmet":["C01","C01"]}`, exit: 2, err: "duplicate_criterion"},
		{name: "hand off without verification", n: "01", meta: inProgress, op: "hand-off", input: commit(shaB), exit: 2, err: "invalid_input"},
		{name: "hand off while open", n: "01", op: "hand-off", input: `{"commit":"` + shaB + `","verification":["ok"]}`, exit: 1, err: "invalid_transition"},
		{name: "hand off from an ordinary stop", n: "01", meta: append(append([]string{}, inProgress...), "Status", "stopped"), op: "hand-off", input: `{"verification":["ok"]}`, exit: 1, err: "invalid_transition"},

		{name: "pass in implement", n: "01", op: "pass", input: `{}`, exit: 1, err: "invalid_transition"},
		{name: "pass manual", n: "03", op: "pass", input: `{}`, exit: 1, err: "invalid_transition"},
		{name: "pass open acceptance", n: "04", op: "pass", input: `{}`, exit: 1, err: "invalid_transition"},
		{name: "pass acceptance with an open blocker", n: "04", done: []string{"01", "02"}, meta: accInProgress, op: "pass", input: `{}`, exit: 1, err: "not_ready"},

		{name: "send back without findings", n: "01", meta: reviewOpen, op: "send-back", input: `{"verification":["ok"]}`, exit: 2, err: "invalid_input"},
		{name: "send back with a not-blocking finding", n: "01", meta: reviewOpen, op: "send-back", input: `{"findings":["Not blocking: x"]}`, exit: 2, err: "invalid_input"},
		{name: "send back from implement", n: "01", op: "send-back", input: `{"findings":["x"]}`, exit: 1, err: "invalid_transition"},

		{name: "set feature base twice", n: "04", meta: []string{"Base commit", shaA}, op: "set-feature-base", input: commit(shaB), exit: 1, err: "base_already_set"},
		{name: "set feature base on implementation", n: "01", op: "set-feature-base", input: commit(shaA), exit: 1, err: "invalid_transition"},

		{name: "start review blocked", n: "04", done: []string{"01", "02"}, meta: []string{"Base commit", shaA}, op: "start-review", input: commit(shaB), exit: 1, err: "not_ready"},
		{name: "start review without a base", n: "04", done: allDone, op: "start-review", input: commit(shaB), exit: 1, err: "missing_base"},
		{name: "start review of an old revision", n: "04", done: allDone, meta: []string{"Base commit", shaA, "Spec revision", "2"}, op: "start-review", input: commit(shaB), exit: 1, err: "stale_acceptance"},

		{name: "stop needs a reason", n: "01", op: "stop", input: `{}`, exit: 2, err: "invalid_input"},
		{name: "stop a done ticket", n: "01", meta: []string{"Status", "done"}, op: "stop", input: `{"reason":"x"}`, exit: 1, err: "invalid_transition"},
		{name: "resume an open ticket", n: "01", op: "resume", input: `{"reason":"x"}`, exit: 1, err: "invalid_transition"},
		{name: "reset attempts in review", n: "01", meta: reviewStopped, op: "reset-attempts", input: `{"reason":"x"}`, exit: 1, err: "invalid_transition"},
		{name: "revise a done ticket", n: "01", meta: []string{"Status", "done"}, op: "revise", input: `{"reason":"x"}`, exit: 1, err: "invalid_transition"},
		{name: "complete without evidence", n: "03", op: "complete", input: `{}`, exit: 2, err: "invalid_input"},
		{name: "complete implementation", n: "01", op: "complete", input: `{"evidence":["x"]}`, exit: 1, err: "invalid_transition"},
		{name: "cancel a done ticket", n: "01", meta: []string{"Status", "done"}, op: "cancel", input: "", exit: 1, err: "invalid_transition"},
		{name: "empty comment", n: "01", op: "comment", input: `{}`, exit: 2, err: "invalid_input"},

		{name: "field another operation takes", n: "01", op: "claim", input: `{"commit":"` + shaA + `","reason":"x"}`, exit: 2, err: "invalid_input"},
		{name: "field nobody takes", n: "01", op: "stop", input: `{"reason":"x","status":"done"}`, exit: 2, err: "invalid_input"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := setup(t, c.done, c.n, c.meta)
			r.opFails(c.exit, c.err, co, c.n, c.op, c.input)
		})
	}
}

// TestCommitFields checks which operations set the base and target
// commits, and that later operations keep them.
func TestCommitFields(t *testing.T) {
	cases := []struct {
		name         string
		n            string
		done         []string
		meta         []string
		op           string
		input        string
		base, target string
	}{
		{name: "claim sets the base", n: "01", op: "claim", input: commit(shaA), base: shaA},
		{name: "claim keeps an existing base", n: "01", meta: []string{"Attempt", "1", "Base commit", shaA}, op: "claim", input: commit(shaB), base: shaA},
		{name: "hand off sets the target", n: "01", meta: inProgress, op: "hand-off", input: `{"commit":"` + shaB + `","verification":["ok"]}`, base: shaA, target: shaB},
		{name: "send back keeps both", n: "01", meta: reviewOpen, op: "send-back", input: `{"findings":["Blocking (wrong): x."]}`, base: shaA, target: shaB},
		{name: "revise keeps both", n: "01", meta: reviewStopped, op: "revise", input: `{"reason":"Spec changed."}`, base: shaA, target: shaB},
		{name: "set feature base", n: "04", op: "set-feature-base", input: commit(shaA), base: shaA},
		{name: "start review sets the target", n: "04", done: allDone, meta: []string{"Base commit", shaA}, op: "start-review", input: commit(shaB), base: shaA, target: shaB},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := setup(t, c.done, c.n, c.meta)
			r.op(co, c.n, c.op, c.input)
			if tk := r.ticket(co, c.n); tk.Base != c.base || tk.Target != c.target {
				t.Fatalf("base %q target %q, want base %q target %q", tk.Base, tk.Target, c.base, c.target)
			}
		})
	}
}

// TestRepeatedOperations checks operations that find their work already
// done. Each reply carries a note. Only the claim on a ticket that already
// has a base commit writes anything: it starts a new attempt and keeps the
// old base.
func TestRepeatedOperations(t *testing.T) {
	cases := []struct {
		name      string
		n         string
		done      []string
		meta      []string
		op        string
		input     string
		unchanged bool
	}{
		{name: "claim with a base already set", n: "01", meta: []string{"Attempt", "1", "Base commit", shaA}, op: "claim", input: commit(shaB)},
		{name: "claim while in progress", n: "01", meta: inProgress, op: "claim", input: commit(shaB), unchanged: true},
		{name: "start review while in progress", n: "04", done: allDone, meta: accInProgress, op: "start-review", input: commit(shaC), unchanged: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := setup(t, c.done, c.n, c.meta)
			before := r.read(r.ticket(co, c.n).Path)
			w := r.op(co, c.n, c.op, c.input)
			if w.Note == "" || w.Unchanged != c.unchanged || (r.read(r.ticket(co, c.n).Path) == before) != c.unchanged {
				t.Fatalf("%+v", w)
			}
		})
	}
}

// criteria returns the lines of a ticket's Acceptance criteria section.
func criteria(r *repo, n string) string {
	r.t.Helper()
	body := r.read(r.ticket(co, n).Path)
	_, rest, _ := strings.Cut(body, "## Acceptance criteria\n\n")
	list, _, _ := strings.Cut(rest, "\n\n## ")
	return list
}

func TestCriteriaChanges(t *testing.T) {
	const (
		zero = "An empty order shows zero."
		sum  = "The count equals the sum of quantities."
	)
	cases := []struct {
		name    string
		n       string
		done    []string
		meta    []string
		tickAll bool // tick every criterion by hand first
		op      string
		input   string
		want    string
	}{
		{name: "hand off ticks met criteria", n: "01", meta: inProgress, op: "hand-off",
			input: `{"commit":"` + shaB + `","criteria_met":["C01"],"verification":["ok"]}`, want: "- [x] " + zero + "\n- [ ] " + sum},
		{name: "hand off accepts a quoted criterion", n: "01", meta: inProgress, op: "hand-off",
			input: `{"commit":"` + shaB + `","criteria_met":["\"` + zero + `\""],"verification":["ok"]}`, want: "- [x] " + zero + "\n- [ ] " + sum},
		{name: "send back unticks unmet criteria", n: "01", meta: reviewOpen, tickAll: true, op: "send-back",
			input: `{"criteria_unmet":["C02"]}`, want: "- [x] " + zero + "\n- [ ] " + sum},
		{name: "revise unticks every criterion", n: "01", meta: reviewStopped, tickAll: true, op: "revise",
			input: `{"reason":"Spec changed."}`, want: "- [ ] " + zero + "\n- [ ] " + sum},
		{name: "acceptance pass can tick every criterion", n: "04", done: allDone, meta: accInProgress, op: "pass",
			input: `{"all_criteria_met":true}`, want: "- [x] Every story in the spec works.\n- [x] Feature verification passes."},
		{name: "complete ticks every criterion", n: "03", op: "complete",
			input: `{"evidence":["Pixel 8."]}`, want: "- [x] The count is readable without zooming."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := setup(t, c.done, c.n, c.meta)
			if c.tickAll {
				path := r.ticket(co, c.n).Path
				r.write(path, strings.ReplaceAll(r.read(path), "- [ ] ", "- [x] "))
			}
			r.op(co, c.n, c.op, c.input)
			if got := criteria(r, c.n); got != c.want {
				t.Fatalf("criteria:\n%s\nwant:\n%s", got, c.want)
			}
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

func TestCancelUpdatesOtherTickets(t *testing.T) {
	r := newRepo(t, co)
	w := r.op(co, "02", "cancel", `{"reason":"Not needed."}`)
	if len(w.Changed) != 2 || w.Hashes["04"] != r.ticket(co, "04").Hash || w.CommitSubject != "tickets: cancel 02" {
		t.Fatalf("cancel: %+v", w)
	}
	if b := r.ticket(co, "04").BlockedBy; strings.Join(b, ",") != "01,03" {
		t.Fatalf("blockers %v", b)
	}
	// Losing ticket 02 leaves US 2 without an implementation ticket.
	s := r.status(co)
	if s.Valid || s.Errors[0].Code != "story_not_covered" {
		t.Fatalf("status: %+v", s.Errors)
	}
	r.op(co, "03", "cancel", "")
	r.op(co, "01", "cancel", "")
	if !strings.Contains(r.read(r.ticket(co, "04").Path), "Kind: acceptance\nStatus: open\nCovers:") {
		t.Fatal("an empty Blocked by line must be removed")
	}
	// Cancelling again repairs leftovers and otherwise changes nothing.
	if w := r.op(co, "01", "cancel", ""); !w.Unchanged {
		t.Fatalf("%+v", w)
	}
}

func TestStaleWrites(t *testing.T) {
	r := newRepo(t, co)
	old := r.ticket(co, "01").Hash
	r.setMeta(co, "01", "Covers", "US 1, US 2")
	e := r.fails(1, "stale", "", "ticket", "claim", co, "01", "--if-match", old, "--commit", shaA)
	if e.Details["current_hash"] != r.ticket(co, "01").Hash {
		t.Fatalf("stale error lacks the current hash: %+v", e)
	}
	r.fails(2, "missing_if_match", "", "ticket", "claim", co, "01", "--commit", shaA)
	r.fails(1, "stale", `{"title":"x"}`, "ticket", "edit", co, "01", "--if-match", old, "--input", "-")

	spec := r.specHash(co)
	r.ok(`{"attempt_limit":4}`, "spec", "edit", co, "--if-match", spec, "--input", "-")
	r.fails(1, "stale", `{"attempt_limit":5}`, "spec", "edit", co, "--if-match", spec, "--input", "-")
}

// TestChecksUseCurrentDependencies covers writes whose target hash is
// unchanged but whose blockers or spec changed since the caller's read.
func TestChecksUseCurrentDependencies(t *testing.T) {
	t.Run("a blocker reopened", func(t *testing.T) {
		r := newRepo(t, co)
		r.setMeta(co, "01", "Status", "done")
		h := r.ticket(co, "02").Hash
		r.setMeta(co, "01", "Status", "open")
		r.fails(1, "not_ready", "", "ticket", "claim", co, "02", "--if-match", h, "--commit", shaA)
	})
	t.Run("the attempt limit dropped", func(t *testing.T) {
		r := newRepo(t, co)
		r.setMeta(co, "01", "Attempt", "2", "Base commit", shaA)
		r.ok(`{"attempt_limit":2}`, "spec", "edit", co, "--if-match", r.specHash(co), "--input", "-")
		r.opFails(1, "attempt_limit_reached", co, "01", "claim", commit(shaA))
	})

	accepting := func(t *testing.T) *repo {
		r := setup(t, allDone, "", nil)
		r.op(co, "04", "set-feature-base", commit(shaA))
		r.op(co, "04", "start-review", commit(shaB))
		return r
	}
	t.Run("the spec was revised during the acceptance review", func(t *testing.T) {
		r := accepting(t)
		h := r.ticket(co, "04").Hash
		r.ok(`{"change":"revision","approved":true,"history":"US 1 reworded.","edit_stories":{"US 1":"As a shopper, I want x, so that y"}}`,
			"spec", "edit", co, "--if-match", r.specHash(co), "--input", "-")
		r.fails(1, "stale_acceptance", `{"all_criteria_met":true}`, "ticket", "pass", co, "04", "--if-match", h, "--input", "-")
		r.op(co, "04", "stop", `{"reason":"The spec changed during the review."}`)
	})
	t.Run("a newer acceptance ticket exists", func(t *testing.T) {
		r := accepting(t)
		h := r.ticket(co, "04").Hash
		r.ok(`{"title":"Accept again","kind":"acceptance","blocked_by":["01","02","03"],"covers":["US 1","US 2"],
			"what_to_build":"Review the complete feature against the spec.","criteria":["Every story works."],"verification":"Run `+"`npm test`"+`."}`,
			"ticket", "create", co, "--input", "-")
		r.fails(1, "not_current_acceptance", `{}`, "ticket", "pass", co, "04", "--if-match", h, "--input", "-")
		if s := r.status(co); s.CurrentAcceptance != "05" || s.EarlierAcceptanceBase != shaA {
			t.Fatalf("status: current %s, earlier base %s", s.CurrentAcceptance, s.EarlierAcceptanceBase)
		}
	})
}

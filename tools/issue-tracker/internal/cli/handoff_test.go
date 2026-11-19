package cli_test

import (
	"fmt"
	"strings"
	"testing"
)

func TestPendingHandoffKeepsTheLastAttempt(t *testing.T) {
	r, git := newGitRepo(t)
	r.ok(`{"attempt_limit":1}`, "spec", "edit", co, "--if-match", r.specHash(co), "--input", "-")
	git("add", ".scratch")
	git("commit", "-qm", "spec and tickets")
	r.op(co, "01", "claim", commit("HEAD"))
	git("add", ".scratch")
	git("commit", "-qm", "tickets: claim 01")
	r.write("app.txt", "the implementation\n")
	git("add", "app.txt")
	git("commit", "-qm", "01: implement")
	candidate := git("rev-parse", "HEAD")

	r.opFails(2, "unknown_criterion", co, "01", "hand-off", `{"commit":"HEAD","criteria_met":["C99"],"verification":["npm test passed"]}`)
	e := r.opFails(2, "invalid_input", co, "01", "hand-off", `{"commit":"HEAD","notes":"not an array","verification":["npm test passed"]}`)
	if e.Message != "notes must be an array of strings" {
		t.Fatalf("wrong type: %+v", e)
	}
	stopForHandoff(r)
	git("add", ".scratch")
	git("commit", "-qm", "tickets: stop 01")

	// Start a separate CLI process, which only knows the state on disk.
	args := r.opArgs(co, "01", "hand-off", correctedHandoff)
	res := runProcess(t, r.root, correctedHandoff, args...)
	if res.code != 0 {
		t.Fatalf("hand-off: %s %s", res.raw, res.stderr)
	}
	tk := r.ticket(co, "01")
	if state(tk.Phase, tk.Status, tk.Attempt) != "review/open attempt 1" || tk.Target != candidate || tk.PendingHandoff != nil {
		t.Fatalf("hand-off state: %+v", tk)
	}
	if got := r.lastComment(co, "01"); !strings.HasPrefix(got, "implement:hand-off (attempt 1):\n- Verification: Required checks passed.") {
		t.Fatalf("hand-off comment: %q", got)
	}
	r.op(co, "01", "pass", `{"verification":["npm test passed in review"]}`)
	if tk := r.ticket(co, "01"); state(tk.Phase, tk.Status, tk.Attempt) != "review/done attempt 1" {
		t.Fatalf("review state: %+v", tk)
	}
}

const correctedHandoff = `{"criteria_met":["C01","C02"],"verification":["Required checks passed."]}`

func stopForHandoff(r *repo) {
	r.t.Helper()
	spec := r.specHash(co)
	r.op(co, "01", "stop", fmt.Sprintf(`{"stop_kind":"handoff-pending","reason":"Unexpected criterion ID C99.","commit":"HEAD","spec_hash":%q,"verification":["npm test passed"]}`, spec))
	s := r.status(co)
	tk := s.Tickets[0]
	if !s.Valid || tk.Status != "stopped" || tk.Ready || tk.PendingHandoff == nil || tk.PendingHandoff.SpecHash != spec || !strings.Contains(tk.LastComment, "npm test passed") {
		r.t.Fatalf("pending hand-off: %+v, status: %+v", tk, s)
	}
}

func TestPendingHandoffSnapshot(t *testing.T) {
	for _, name := range []string{"spec changed", "ticket changed", "source changed", "other scratch files changed", "dirty checkout", "wrong commit", "resume"} {
		t.Run(name, func(t *testing.T) {
			r, git := newGitRepo(t)
			git("add", ".scratch")
			git("commit", "-qm", "spec and tickets")
			r.op(co, "01", "claim", commit("HEAD"))
			git("add", ".scratch")
			git("commit", "-qm", "tickets: claim 01")
			stopForHandoff(r)
			code, input := "stale_handoff", correctedHandoff
			switch name {
			case "spec changed":
				r.ok(`{"change":"editorial","solution":"Adjusted solution."}`, "spec", "edit", co, "--if-match", r.specHash(co), "--input", "-")
			case "ticket changed":
				r.ok(`{"what_to_build":"Different work."}`, ticketEditArgs(r, "01")...)
			case "source changed", "dirty checkout":
				r.write("app.txt", "different work\n")
				if name == "source changed" {
					git("add", "app.txt")
				} else {
					code = "handoff_checkout"
				}
			case "other scratch files changed":
				r.write(".scratch/tool-config.txt", "different configuration\n")
			case "wrong commit":
				input = `{"commit":"HEAD","verification":["Required checks passed."]}`
			case "resume":
				r.op(co, "01", "resume", `{"reason":"Start new work."}`)
				if tk := r.ticket(co, "01"); tk.PendingHandoff != nil || state(tk.Phase, tk.Status, tk.Attempt) != "implement/open attempt 1" {
					t.Fatalf("after resume: %+v", tk)
				}
				code = "invalid_transition"
			}
			git("add", ".scratch")
			git("commit", "-qm", "changes after the stop")
			r.opFails(1, code, co, "01", "hand-off", input)
		})
	}
}

package cli_test

import (
	"strings"
	"testing"
)

func TestRejectedReportKeepsTheAttempt(t *testing.T) {
	r := newRepo(t, co)
	r.ok(`{"attempt_limit":1}`, "spec", "edit", co, "--if-match", r.specHash(co), "--input", "-")
	r.op(co, "01", "claim", commit(shaA))

	r.opFails(2, "unknown_criterion", co, "01", "hand-off", `{"commit":"`+shaB+`","criteria_met":["C99"],"verification":["npm test passed"]}`)
	e := r.opFails(2, "invalid_input", co, "01", "hand-off", `{"commit":"`+shaB+`","notes":"not an array","verification":["npm test passed"]}`)
	if e.Message != "notes must be an array of strings" {
		t.Fatalf("wrong type: %+v", e)
	}

	// A fresh implementer claims again and continues the same attempt.
	if w := r.op(co, "01", "claim", commit(shaB)); !strings.Contains(w.Note, "current attempt continues") {
		t.Fatalf("second claim: %+v", w)
	}
	r.op(co, "01", "hand-off", `{"commit":"`+shaB+`","criteria_met":["C01","C02"],"verification":["Required checks passed."]}`)
	tk := r.ticket(co, "01")
	if state(tk.Phase, tk.Status, tk.Attempt) != "review/open attempt 1" || tk.Target != shaB {
		t.Fatalf("hand-off state: %+v", tk)
	}
	if got := r.lastComment(co, "01"); !strings.HasPrefix(got, "implement:hand-off (attempt 1):\n- Verification: Required checks passed.") {
		t.Fatalf("hand-off comment: %q", got)
	}
}

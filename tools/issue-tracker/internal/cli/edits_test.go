package cli_test

import (
	"strings"
	"testing"
)

const newSpec = `{"title":"Order summary","problem_statement":"Shoppers can't see counts.","solution":"Show counts.",
"user_stories":["As a shopper, I want counts, so that I know what I ordered","As a shopper, I want counts on my phone, so that I can check on the go"],
"implementation_decisions":"- Add a count to the ` + "`summary`" + ` module.","testing_decisions":"- Run ` + "`npm test`" + `.\n- A manual check on a phone.",
"out_of_scope":"- Totals.","further_notes":"Totals come next."}`

func ticketJSON(title, kind, covers, blockedBy string) string {
	return `{"title":"` + title + `","kind":"` + kind + `","covers":` + covers + `,"blocked_by":` + blockedBy + `,
		"what_to_build":"Build ` + title + `.","criteria":["` + title + ` works."],"verification":"Run ` + "`npm test`" + `."}`
}

func TestDraftingAFeature(t *testing.T) {
	r := newRepo(t)
	var w writeReply
	r.ok(&w, newSpec, "spec", "create", "orders", "--input", "-")
	if w.Revision != 1 || w.CommitSubject != "spec: create orders" || w.Changed[0] != ".scratch/orders/spec.md" {
		t.Fatalf("spec create: %+v", w)
	}
	golden(t, "drafting/spec.md", r.read(".scratch/orders/spec.md"))
	r.fails("feature_exists", newSpec, "spec", "create", "orders", "--input", "-")

	r.ok(&w, ticketJSON("Show counts", "implementation", `["US 1"]`, `[]`), "ticket", "create", "orders", "--input", "-")
	if w.Ticket != "01" || w.CommitSubject != "tickets: create 01" || w.Changed[0] != ".scratch/orders/issues/01-show-counts.md" {
		t.Fatalf("ticket create: %+v", w)
	}
	golden(t, "drafting/01-show-counts.md", r.read(w.Changed[0]))

	// A partial breakdown passes draft validation but not full validation.
	r.ok(nil, "", "validate", "orders", "--draft")
	r.fails("validation_failed", "", "validate", "orders")

	r.fails("invalid_reference", ticketJSON("Bad blocker", "implementation", `["US 1"]`, `["07"]`), "ticket", "create", "orders", "--input", "-")
	r.fails("invalid_reference", ticketJSON("Bad story", "implementation", `["US 9"]`, `[]`), "ticket", "create", "orders", "--input", "-")
	r.fails("invalid_input", `{"title":"No covers","kind":"manual","what_to_build":"x","criteria":["x"],"verification":"x"}`, "ticket", "create", "orders", "--input", "-")
	r.fails("invalid_input", `{"title":"x","kind":"manual","covers":[],"what_to_build":"x","criteria":["x"],"verification":"x","status":"done"}`, "ticket", "create", "orders", "--input", "-")
	r.fails("invalid_input", `{"title":"x","kind":"manual","covers":[],"what_to_build":"x","criteria":["x","x"],"verification":"x"}`, "ticket", "create", "orders", "--input", "-")

	r.ok(nil, ticketJSON("Phone layout", "implementation", `["US 2"]`, `["01"]`), "ticket", "create", "orders", "--input", "-")
	r.ok(nil, ticketJSON("Check on a phone", "manual", `["US 2"]`, `["01"]`), "ticket", "create", "orders", "--input", "-")
	r.ok(&w, ticketJSON("Accept", "acceptance", `["US 1","US 2"]`, `["01","02","03"]`), "ticket", "create", "orders", "--input", "-")
	golden(t, "drafting/04-accept.md", r.read(w.Changed[0]))
	r.ok(nil, "", "validate", "orders")
}

func TestSpecEdits(t *testing.T) {
	r := newRepo(t, co)
	edit := func(input string) writeReply {
		t.Helper()
		var w writeReply
		r.ok(&w, input, "spec", "edit", co, "--if-match", r.specHash(co), "--input", "-")
		return w
	}
	editFails := func(code, input string) {
		t.Helper()
		r.fails(code, input, "spec", "edit", co, "--if-match", r.specHash(co), "--input", "-")
	}

	editFails("invalid_input", `{"solution":"New."}`)
	editFails("invalid_input", `{"change":"revision","solution":"New.","history":"x"}`)
	editFails("invalid_input", `{"change":"revision","approved":true,"solution":"New."}`)
	editFails("invalid_input", `{"change":"editorial","add_stories":["As a x, I want y, so that z"]}`)
	editFails("invalid_input", `{"change":"editorial","attempt_limit":4}`)
	editFails("invalid_input", `{"change":"editorial","solution":"## Sneaky"}`)
	editFails("invalid_input", `{"change":"editorial","edit_stories":{"US 1":"US 1: doubled prefix"}}`)
	editFails("unknown_story", `{"change":"editorial","edit_stories":{"US 9":"x"}}`)

	// An attempt-limit change touches only that line and keeps the revision.
	before := r.read(".scratch/checkout/spec.md")
	if w := edit(`{"attempt_limit":5}`); w.Revision != 1 {
		t.Fatalf("attempt limit: %+v", w)
	}
	if r.read(".scratch/checkout/spec.md") != strings.Replace(before, "Attempt limit: 3", "Attempt limit: 5", 1) {
		t.Fatal("an attempt-limit edit changed other lines")
	}

	if w := edit(`{"change":"editorial","solution":"Show item counts in the summary."}`); w.Revision != 1 || w.CommitSubject != "spec: revise checkout" {
		t.Fatalf("editorial: %+v", w)
	}
	if w := edit(`{"change":"revision","approved":true,"history":"US 3 added. US 1 reworded.",
		"add_stories":["As a shopper, I want totals, so that I can pay"],
		"edit_stories":{"US 1":"As a shopper, I want the item count, so that I know what I ordered"},
		"assumptions":"- Prices exist."}`); w.Revision != 2 {
		t.Fatalf("revision: %+v", w)
	}
	if !strings.Contains(r.read(".scratch/checkout/spec.md"), "## Out of scope\n\n- Order totals.\n\n## Assumptions\n\n- Prices exist.\n\n## Change history\n") {
		t.Fatalf("a new section must go in template order:\n%s", r.read(".scratch/checkout/spec.md"))
	}
	edit(`{"change":"revision","approved":true,"history":"US 2 withdrawn.","withdraw_stories":["US 2"]}`)
	editFails("invalid_input", `{"change":"editorial","edit_stories":{"US 2":"Back again"}}`)
	edit(`{"change":"editorial","assumptions":""}`)
	golden(t, "spec-edits/spec.md", r.read(".scratch/checkout/spec.md"))
}

func TestTicketEdits(t *testing.T) {
	r := newRepo(t, co)
	editArgs := func(n string) []string {
		return []string{"ticket", "edit", co, n, "--if-match", r.ticket(co, n).Hash, "--input", "-"}
	}

	r.op(co, "01", "claim", commit(shaA))
	r.op(co, "01", "hand-off", `{"commit":"`+shaB+`","criteria_met":["An empty order shows zero."],"verification":["ok"]}`)
	var w writeReply
	r.ok(&w, `{"criteria":["An empty order shows zero.","A new criterion."],"out_of_scope":""}`, editArgs("01")...)
	if w.Status != "open" || w.Phase != "review" || w.CommitSubject != "tickets: edit 01" {
		t.Fatalf("edit reply: %+v", w)
	}
	golden(t, "ticket-edits/01-show-counts.md", r.read(r.ticket(co, "01").Path))

	r.fails("invalid_reference", `{"blocked_by":["02"]}`, editArgs("01")...)
	r.fails("invalid_input", `{"spec_revision":1}`, editArgs("01")...)
	r.fails("invalid_input", `{"status":"done"}`, editArgs("01")...)
	r.fails("invalid_input", `{"titel":"Typo"}`, editArgs("01")...)

	r.ok(nil, `{"covers":[],"title":"Prefactor"}`, editArgs("02")...)
	tk := r.ticket(co, "02")
	if tk.Title != "Prefactor" || len(tk.Covers) != 0 || !strings.Contains(r.read(tk.Path), "Covers: none\n") || !strings.HasSuffix(tk.Path, "02-phone-layout.md") {
		t.Fatalf("%+v", tk)
	}
	// An edit that changes nothing doesn't rewrite the file.
	h := tk.Hash
	if r.ok(&w, `{"title":"Prefactor"}`, editArgs("02")...); !w.Unchanged || w.Hash != h {
		t.Fatalf("no-op edit: %+v", w)
	}

	r.setMeta(co, "03", "Status", "done")
	r.fails("ticket_finished", `{"title":"x"}`, editArgs("03")...)
}

// TestEditsCantBreakValidation covers the to-tickets order: an in-progress
// acceptance ticket must be revised before it gets a new blocker.
func TestEditsCantBreakValidation(t *testing.T) {
	r := newRepo(t, co)
	for _, n := range []string{"01", "02", "03"} {
		r.setMeta(co, n, "Status", "done")
	}
	r.op(co, "04", "set-feature-base", commit(shaA))
	r.op(co, "04", "start-review", commit(shaB))
	r.ok(nil, ticketJSON("Repair", "implementation", `["US 1"]`, `[]`), "ticket", "create", co, "--input", "-")

	blockers := `{"blocked_by":["01","02","03","05"]}`
	before := r.read(r.ticket(co, "04").Path)
	e := r.fails("would_invalidate", blockers, "ticket", "edit", co, "04", "--if-match", r.ticket(co, "04").Hash, "--input", "-")
	if !strings.Contains(e.Message, "revise the acceptance ticket") || r.read(r.ticket(co, "04").Path) != before {
		t.Fatalf("%+v", e)
	}
	r.op(co, "04", "revise", `{"reason":"Ticket 05 repairs totals."}`)
	r.ok(nil, blockers, "ticket", "edit", co, "04", "--if-match", r.ticket(co, "04").Hash, "--input", "-")
	r.ok(nil, "", "validate", co)
}

// TestExistingProblemsDontBlockRepairs checks that writes on an already
// invalid feature work when they add no new problem.
func TestExistingProblemsDontBlockRepairs(t *testing.T) {
	r := newRepo(t, co)
	r.setMeta(co, "04", "Status", "in-progress", "Base commit", shaA, "Target commit", shaB)
	r.op(co, "01", "claim", commit(shaA))
	r.op(co, "04", "stop", `{"reason":"Started too early."}`)
	r.ok(nil, "", "validate", co)
}

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
	w := okReply[writeReply](r, newSpec, "spec", "create", "orders", "--input", "-")
	if w.Revision != 1 || w.CommitSubject != "spec: create orders" || w.Changed[0] != ".scratch/orders/spec.md" {
		t.Fatalf("spec create: %+v", w)
	}
	golden(t, "drafting/spec.md", r.read(".scratch/orders/spec.md"))

	w = okReply[writeReply](r, ticketJSON("Show counts", "implementation", `["US 1"]`, `[]`), "ticket", "create", "orders", "--input", "-")
	if w.Ticket != "01" || w.CommitSubject != "tickets: create 01" || w.Changed[0] != ".scratch/orders/issues/01-show-counts.md" {
		t.Fatalf("ticket create: %+v", w)
	}
	golden(t, "drafting/01-show-counts.md", r.read(w.Changed[0]))

	// A partial breakdown passes draft validation but not full validation.
	r.ok("", "validate", "orders", "--draft")
	r.fails(1, "validation_failed", "", "validate", "orders")

	r.ok(ticketJSON("Phone layout", "implementation", `["US 2"]`, `["01"]`), "ticket", "create", "orders", "--input", "-")
	r.ok(ticketJSON("Check on a phone", "manual", `["US 2"]`, `["01"]`), "ticket", "create", "orders", "--input", "-")
	w = okReply[writeReply](r, ticketJSON("Accept", "acceptance", `["US 1","US 2"]`, `["01","02","03"]`), "ticket", "create", "orders", "--input", "-")
	golden(t, "drafting/04-accept.md", r.read(w.Changed[0]))
	r.ok("", "validate", "orders")
}

func TestCreateRejects(t *testing.T) {
	ticketCreate := []string{"ticket", "create", co, "--input", "-"}
	for _, c := range []struct {
		name  string
		args  []string
		input string
		exit  int
		err   string
	}{
		{"spec for an existing feature", []string{"spec", "create", co, "--input", "-"}, newSpec, 1, "feature_exists"},
		{"unknown blocker", ticketCreate, ticketJSON("Bad blocker", "implementation", `["US 1"]`, `["07"]`), 1, "invalid_reference"},
		{"unknown story", ticketCreate, ticketJSON("Bad story", "implementation", `["US 9"]`, `[]`), 1, "invalid_reference"},
		{"no covers", ticketCreate, `{"title":"No covers","kind":"manual","what_to_build":"x","criteria":["x"],"verification":"x"}`, 2, "invalid_input"},
		{"state field", ticketCreate, `{"title":"x","kind":"manual","covers":[],"what_to_build":"x","criteria":["x"],"verification":"x","status":"done"}`, 2, "invalid_input"},
		{"duplicate criterion", ticketCreate, `{"title":"x","kind":"manual","covers":[],"what_to_build":"x","criteria":["x","x"],"verification":"x"}`, 2, "invalid_input"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := newRepo(t, co)
			before := snapshot(t, r.root)
			r.fails(c.exit, c.err, c.input, c.args...)
			if snapshot(t, r.root) != before {
				t.Fatal("a failed create changed the checkout")
			}
		})
	}
}

func TestSpecEdits(t *testing.T) {
	r := newRepo(t, co)
	edit := func(input string) writeReply {
		t.Helper()
		return okReply[writeReply](r, input, "spec", "edit", co, "--if-match", r.specHash(co), "--input", "-")
	}

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
	r.fails(2, "invalid_input", `{"change":"editorial","edit_stories":{"US 2":"Back again"}}`, "spec", "edit", co, "--if-match", r.specHash(co), "--input", "-")
	edit(`{"change":"editorial","assumptions":""}`)
	golden(t, "spec-edits/spec.md", r.read(".scratch/checkout/spec.md"))
}

func TestSpecEditRejects(t *testing.T) {
	for _, c := range []struct {
		name  string
		input string
		err   string
	}{
		{"content without a change kind", `{"solution":"New."}`, "invalid_input"},
		{"revision without approval", `{"change":"revision","solution":"New.","history":"x"}`, "invalid_input"},
		{"revision without history", `{"change":"revision","approved":true,"solution":"New."}`, "invalid_input"},
		{"editorial story addition", `{"change":"editorial","add_stories":["As a x, I want y, so that z"]}`, "invalid_input"},
		{"change kind on an attempt-limit edit", `{"change":"editorial","attempt_limit":4}`, "invalid_input"},
		{"heading in section text", `{"change":"editorial","solution":"## Sneaky"}`, "invalid_input"},
		{"story ID in story text", `{"change":"editorial","edit_stories":{"US 1":"US 1: doubled prefix"}}`, "invalid_input"},
		{"unknown story", `{"change":"editorial","edit_stories":{"US 9":"x"}}`, "unknown_story"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := newRepo(t, co)
			before := r.read(".scratch/checkout/spec.md")
			r.fails(2, c.err, c.input, "spec", "edit", co, "--if-match", r.specHash(co), "--input", "-")
			if r.read(".scratch/checkout/spec.md") != before {
				t.Fatal("a failed edit changed the spec")
			}
		})
	}
}

func ticketEditArgs(r *repo, n string) []string {
	return []string{"ticket", "edit", co, n, "--if-match", r.ticket(co, n).Hash, "--input", "-"}
}

func TestTicketEdits(t *testing.T) {
	r := newRepo(t, co)
	r.op(co, "01", "claim", commit(shaA))
	r.op(co, "01", "hand-off", `{"commit":"`+shaB+`","criteria_met":["An empty order shows zero."],"verification":["ok"]}`)
	w := okReply[writeReply](r, `{"criteria":["An empty order shows zero.","A new criterion."],"out_of_scope":""}`, ticketEditArgs(r, "01")...)
	if w.Status != "open" || w.Phase != "review" || w.CommitSubject != "tickets: edit 01" {
		t.Fatalf("edit reply: %+v", w)
	}
	golden(t, "ticket-edits/01-show-counts.md", r.read(r.ticket(co, "01").Path))

	r.ok(`{"covers":[],"title":"Prefactor"}`, ticketEditArgs(r, "02")...)
	tk := r.ticket(co, "02")
	if tk.Title != "Prefactor" || len(tk.Covers) != 0 || !strings.Contains(r.read(tk.Path), "Covers: none\n") || !strings.HasSuffix(tk.Path, "02-phone-layout.md") {
		t.Fatalf("%+v", tk)
	}
	// An edit that changes nothing doesn't rewrite the file.
	if w := okReply[writeReply](r, `{"title":"Prefactor"}`, ticketEditArgs(r, "02")...); !w.Unchanged || w.Hash != tk.Hash {
		t.Fatalf("no-op edit: %+v", w)
	}
}

func TestCriterionIDsSurviveReordering(t *testing.T) {
	r := newRepo(t, co)
	path := r.ticket(co, "01").Path
	before := r.read(path)
	legacy := okReply[docReply](r, "", "ticket", "get", co, "01").Criteria
	if len(legacy) != 2 || legacy[0].ID != "C01" || legacy[1].ID != "C02" || r.read(path) != before {
		t.Fatalf("legacy criteria: %+v", legacy)
	}
	r.op(co, "01", "claim", commit(shaA))
	r.op(co, "01", "hand-off", `{"commit":"`+shaB+`","criteria_met":["C01"],"verification":["ok"]}`)
	r.ok(`{"criteria":["The count equals the sum of quantities.","An empty order shows zero."]}`, ticketEditArgs(r, "01")...)
	cs := okReply[docReply](r, "", "ticket", "get", co, "01").Criteria
	if len(cs) != 2 || cs[0].ID != "C02" || cs[0].Checked || cs[1].ID != "C01" || !cs[1].Checked {
		t.Fatalf("reordered criteria: %+v", cs)
	}
	r.ok(`{"criteria":["An empty order shows zero."]}`, ticketEditArgs(r, "01")...)
	r.ok(`{"criteria":["An empty order shows zero.","A new criterion."]}`, ticketEditArgs(r, "01")...)
	cs = okReply[docReply](r, "", "ticket", "get", co, "01").Criteria
	if len(cs) != 2 || cs[1].ID != "C03" {
		t.Fatalf("criteria after deletion and addition: %+v", cs)
	}
}

func TestTicketEditRejects(t *testing.T) {
	for _, c := range []struct {
		name  string
		n     string
		meta  []string // metadata set by hand on ticket n first
		input string
		exit  int
		err   string
	}{
		{"blocker cycle", "01", nil, `{"blocked_by":["02"]}`, 1, "invalid_reference"},
		{"spec revision on implementation", "01", nil, `{"spec_revision":1}`, 2, "invalid_input"},
		{"state field", "01", nil, `{"status":"done"}`, 2, "invalid_input"},
		{"misspelled field", "01", nil, `{"titel":"Typo"}`, 2, "invalid_input"},
		{"finished ticket", "03", []string{"Status", "done"}, `{"title":"x"}`, 1, "ticket_finished"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := newRepo(t, co)
			if c.meta != nil {
				r.setMeta(co, c.n, c.meta...)
			}
			before := snapshot(t, r.root)
			r.fails(c.exit, c.err, c.input, ticketEditArgs(r, c.n)...)
			if snapshot(t, r.root) != before {
				t.Fatal("a failed edit changed the checkout")
			}
		})
	}
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
	r.ok(ticketJSON("Repair", "implementation", `["US 1"]`, `[]`), "ticket", "create", co, "--input", "-")

	blockers := `{"blocked_by":["01","02","03","05"]}`
	before := r.read(r.ticket(co, "04").Path)
	e := r.fails(1, "would_invalidate", blockers, ticketEditArgs(r, "04")...)
	if !strings.Contains(e.Message, "revise the acceptance ticket") || r.read(r.ticket(co, "04").Path) != before {
		t.Fatalf("%+v", e)
	}
	r.op(co, "04", "revise", `{"reason":"Ticket 05 repairs totals."}`)
	r.ok(blockers, ticketEditArgs(r, "04")...)
	r.ok("", "validate", co)
}

// TestExistingProblemsDontBlockRepairs checks that writes on an already
// invalid feature work when they add no new problem.
func TestExistingProblemsDontBlockRepairs(t *testing.T) {
	r := newRepo(t, co)
	r.setMeta(co, "04", "Status", "in-progress", "Base commit", shaA, "Target commit", shaB)
	r.op(co, "01", "claim", commit(shaA))
	r.op(co, "04", "stop", `{"reason":"Started too early."}`)
	r.ok("", "validate", co)
}

// TestExistingCyclesDontBlockOtherBlockerEdits starts from a feature where
// 02 and 03 block each other and 01 names a missing blocker. Blocker
// changes that don't close a new cycle must still work.
func TestExistingCyclesDontBlockOtherBlockerEdits(t *testing.T) {
	for _, c := range []struct {
		name  string
		n     string // the ticket to edit, or "" to create one
		input string
		err   string
	}{
		{"remove a missing blocker", "01", `{"blocked_by":[]}`, ""},
		{"add a blocker from the cycle", "01", `{"blocked_by":["03"]}`, ""},
		{"create a ticket blocked by the cycle", "", ticketJSON("Repair", "implementation", `["US 1"]`, `["02"]`), ""},
		{"close a new cycle through the old one", "02", `{"blocked_by":["03","04"]}`, "invalid_reference"},
		{"close a new cycle from outside", "01", `{"blocked_by":["04"]}`, "invalid_reference"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := newRepo(t, co)
			r.setMeta(co, "01", "Blocked by", "09")
			r.setMeta(co, "02", "Blocked by", "03")
			r.setMeta(co, "03", "Blocked by", "02")
			args := []string{"ticket", "create", co, "--input", "-"}
			if c.n != "" {
				args = ticketEditArgs(r, c.n)
			}
			if c.err != "" {
				before := snapshot(t, r.root)
				e := r.fails(1, c.err, c.input, args...)
				if !strings.Contains(e.Message, "cycle") || snapshot(t, r.root) != before {
					t.Fatalf("%+v", e)
				}
				return
			}
			r.ok(c.input, args...)
			// The old cycle is still there and is still reported.
			errs := validationErrors(t, r.run("", "validate", co, "--draft"), "draft")
			checkProblems(t, "validate --draft", errs, []string{"blocker_cycle"}, nil)
		})
	}
}

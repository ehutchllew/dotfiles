package cli_test

import (
	"strings"
	"testing"
)

const orders = "order-summary"

func TestDocumentedExamplesAreUsable(t *testing.T) {
	r := newRepo(t, orders)
	s := r.status(orders)
	if !s.Valid || s.CurrentAcceptance != "03" || s.Revision != 1 || s.AttemptLimit != 3 || len(s.Tickets) != 3 {
		t.Fatalf("status: %+v", s)
	}
	if !s.Tickets[0].Ready || s.Tickets[1].Ready || s.Tickets[2].Ready {
		t.Fatalf("readiness: %+v", s.Tickets)
	}
	features := okReply[struct {
		Features []struct{ Feature, Title string } `json:"features"`
	}](r, "", "feature", "list")
	if len(features.Features) != 1 || features.Features[0].Title != "Order summary" {
		t.Fatalf("features: %+v", features)
	}
	// Reads return the file bytes unchanged.
	spec := okReply[docReply](r, "", "spec", "get", orders)
	tk := okReply[docReply](r, "", "ticket", "get", orders, "1")
	if *spec.Body != r.read(".scratch/order-summary/spec.md") || *tk.Body != r.read(r.ticket(orders, "01").Path) {
		t.Fatal("a read changed the document bytes")
	}
}

// TestLifecycle runs a feature from setup to acceptance with simulated
// worker results: a send-back, a retry, manual evidence, and an acceptance
// review. The final documents must match the golden files byte for byte.
func TestLifecycle(t *testing.T) {
	r := newRepo(t, orders)
	var w writeReply

	r.op(orders, "03", "set-feature-base", `{"commit":"`+shaA+`"}`)

	w = r.op(orders, "01", "claim", `{"commit":"`+shaA+`"}`)
	if *w.Attempt != 1 || w.Status != "in-progress" || w.CommitSubject != "tickets: claim 01" || len(w.Changed) != 1 {
		t.Fatalf("claim: %+v", w)
	}
	if w.Hash != r.ticket(orders, "01").Hash {
		t.Fatal("the reply's hash doesn't match the file")
	}
	if w = r.op(orders, "01", "claim", `{"commit":"`+shaB+`"}`); !w.Unchanged || len(w.Changed) != 0 {
		t.Fatalf("a second claim must continue the attempt: %+v", w)
	}

	r.op(orders, "01", "hand-off", `{"commit":"`+shaB+`",
		"criteria_met":["The item count equals the sum of the order's item quantities."],
		"verification":["`+"`npm test`"+` passed."]}`)
	w = r.op(orders, "01", "send-back", `{
		"criteria_unmet":["An empty order shows an item count of zero."],
		"findings":["Blocking (missing): an empty order shows no item count line."],
		"verification":["`+"`npm test`"+` passed."],
		"not_blocking":["the local `+"`n`"+` doesn't say it holds the item count."]}`)
	if w.Phase != "implement" || w.Status != "open" || *w.Attempt != 1 || w.CommitSubject != "tickets: send back 01" {
		t.Fatalf("send-back: %+v", w)
	}
	if s := r.status(orders); len(s.NotBlocking) != 1 || s.NotBlocking[0].Ticket != "01" {
		t.Fatalf("not-blocking findings: %+v", s.NotBlocking)
	}

	if w = r.op(orders, "01", "claim", `{"commit":"`+shaC+`"}`); *w.Attempt != 2 {
		t.Fatalf("retry: %+v", w)
	}
	r.op(orders, "01", "hand-off", `{"commit":"`+shaC+`",
		"criteria_met":["An empty order shows an item count of zero."],
		"verification":["`+"`npm test`"+` passed."]}`)
	r.op(orders, "01", "pass", `{"verification":["`+"`npm test`"+` passed."],
		"resolved":[{"finding":"the local `+"`n`"+` doesn't say it holds the item count.","resolution":"fixed"}]}`)

	if !r.ticket(orders, "02").Ready {
		t.Fatal("the manual ticket must be ready once 01 is done")
	}
	r.op(orders, "02", "complete", `{"evidence":["Pixel 8, Chrome: readable without zoom. Screenshot in the PR."]}`)

	r.op(orders, "03", "start-review", `{"commit":"`+shaC+`"}`)
	r.op(orders, "03", "pass", `{"all_criteria_met":true,
		"verification":["`+"`npm test`"+` passed."],
		"evidence":["Ticket 02: Pixel 8 evidence."]}`)

	s := r.status(orders)
	if !s.Valid || len(s.NotBlocking) != 0 {
		t.Fatalf("final status: %+v", s)
	}
	for _, tk := range s.Tickets {
		if tk.Status != "done" {
			t.Fatalf("ticket %s is %s", tk.Ticket, tk.Status)
		}
		path := r.ticket(orders, tk.Ticket).Path
		golden(t, "lifecycle/"+path[strings.LastIndex(path, "/")+1:], r.read(path))
	}
}

func TestCRLFDocumentsKeepTheirLineEndings(t *testing.T) {
	r := newRepo(t, orders)
	path := r.ticket(orders, "01").Path
	crlf := strings.ReplaceAll(r.read(path), "\n", "\r\n")
	r.write(path, crlf)

	if d := okReply[docReply](r, "", "ticket", "get", orders, "01"); *d.Body != crlf {
		t.Fatal("ticket get changed CRLF bytes")
	}
	r.op(orders, "01", "claim", `{"commit":"`+shaA+`"}`)
	r.op(orders, "01", "comment", `{"notes":["first\nsecond"]}`)
	got := r.read(path)
	if strings.Count(got, "\n") != strings.Count(got, "\r\n") {
		t.Fatalf("mixed line endings:\n%q", got)
	}
	if !strings.Contains(got, "Status: in-progress\r\n") || !r.status(orders).Valid {
		t.Fatalf("CRLF ticket after writes:\n%q", got)
	}
}

func TestHandWrittenContentIsPreserved(t *testing.T) {
	r := newRepo(t, orders)
	path := r.ticket(orders, "01").Path
	sketch := "## Design sketch\n\n```text\nStatus: done\n## not a heading\n```\n\n"
	r.edit(path, "## Comments\n", sketch+"## Comments\n\nA note someone left before any operation.\n")

	if tk := r.ticket(orders, "01"); tk.Status != "open" || !r.status(orders).Valid {
		t.Fatalf("fenced text was read as metadata or a section: %+v", tk)
	}
	r.op(orders, "01", "claim", `{"commit":"`+shaA+`"}`)
	r.op(orders, "01", "comment", `{"notes":["hello"]}`)
	got := r.read(path)
	if !strings.Contains(got, sketch) || !strings.Contains(got, "A note someone left before any operation.\n\ncomment (attempt 1):\n- Note: hello\n") {
		t.Fatalf("hand-written content changed:\n%s", got)
	}
}

// TestCommentRendering covers every comment item type, multi-line values,
// and fenced blocks in findings, verification, evidence, reasons, and notes.
func TestCommentRendering(t *testing.T) {
	r := newRepo(t, "checkout")
	fence := "```text\\nok  \\tpkg\\t0.1s\\n```"
	r.setMeta("checkout", "01", "Phase", "review", "Attempt", "1", "Base commit", shaA, "Target commit", shaB)
	r.op("checkout", "01", "send-back", `{
		"criteria_unmet":["An empty order shows zero.","The count equals the sum of quantities."],
		"findings":["Blocking (wrong): output differs:\n`+fence+`"],
		"verification":["`+fence+`","`+"`go test`"+` output:\n`+fence+`"],
		"not_blocking":["Rename `+"`n`"+`.","Split the long function."],
		"notes":["`+fence+`"]}`)
	r.op("checkout", "01", "stop", `{"reason":"Blocked:\n`+fence+`","evidence":["`+fence+`"]}`)
	r.op("checkout", "01", "comment", `{"resolved":[{"finding":"Rename `+"`n`"+`","resolution":"ignored"},
		{"finding":"Split the long function.","resolution":"moved","ticket":"02"}]}`)
	r.op("checkout", "01", "resume", `{"reason":"Fixed the database."}`)

	s := r.status("checkout")
	if !s.Valid || len(s.NotBlocking) != 0 {
		t.Fatalf("status: %+v", s)
	}
	golden(t, "comments/01-show-counts.md", r.read(r.ticket("checkout", "01").Path))
}

// TestCommentInputRejected checks comment input that would break the
// ticket or resolve a finding the ticket doesn't have open.
func TestCommentInputRejected(t *testing.T) {
	for _, c := range []struct {
		name  string
		input string
		exit  int
		err   string
	}{
		// A fence in one value can't swallow the rest of the ticket.
		{"unclosed fence", `{"notes":["` + "```" + `text\nunclosed"]}`, 2, "invalid_input"},
		{"finding text that doesn't match", `{"resolved":[{"finding":"Rename n","resolution":"fixed"}]}`, 1, "unknown_finding"},
		{"finding already resolved", `{"resolved":[{"finding":"Split the long function.","resolution":"fixed"}]}`, 1, "unknown_finding"},
		{"unknown resolution", `{"resolved":[{"finding":"Rename ` + "`n`" + `.","resolution":"deleted"}]}`, 2, "invalid_input"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := newRepo(t, co)
			r.setMeta(co, "01", "Phase", "review", "Attempt", "1", "Base commit", shaA, "Target commit", shaB)
			r.op(co, "01", "send-back", `{"findings":["Blocking (wrong): x."],"not_blocking":["Rename `+"`n`"+`.","Split the long function."]}`)
			r.op(co, "01", "comment", `{"resolved":[{"finding":"Split the long function.","resolution":"fixed"}]}`)
			r.opFails(c.exit, c.err, co, "01", "comment", c.input)
		})
	}
}

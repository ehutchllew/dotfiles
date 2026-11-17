package cli_test

import (
	"strings"
	"testing"
)

func TestConditionalReads(t *testing.T) {
	r := newRepo(t, co)
	var first, again docReply
	r.ok(&first, "", "spec", "get", co)
	if first.Body == nil || first.Revision != 1 {
		t.Fatalf("cold read: %+v", first)
	}
	r.ok(&again, "", "spec", "get", co, "--if-changed", first.Hash)
	if !again.Unchanged || again.Body != nil || again.Hash != first.Hash {
		t.Fatalf("conditional read: %+v", again)
	}

	// A ticket write leaves the spec hash alone.
	r.op(co, "01", "claim", commit(shaA))
	if r.ok(&again, "", "spec", "get", co, "--if-changed", first.Hash); !again.Unchanged {
		t.Fatal("a ticket write changed the spec hash")
	}
	// An attempt-limit change keeps the revision but changes the hash.
	r.ok(nil, `{"attempt_limit":5}`, "spec", "edit", co, "--if-match", first.Hash, "--input", "-")
	if r.ok(&again, "", "spec", "get", co, "--if-changed", first.Hash); again.Unchanged || again.Body == nil || again.Revision != 1 || again.AttemptLimit != 5 {
		t.Fatalf("after an attempt-limit change: %+v", again)
	}
	// A hand edit to a ticket changes its hash.
	tk := r.ticket(co, "02")
	r.edit(tk.Path, "## Comments\n", "## Comments\n\nA hand-written note.\n")
	var d docReply
	if r.ok(&d, "", "ticket", "get", co, "02", "--if-changed", tk.Hash); d.Unchanged || d.Body == nil || d.Hash == tk.Hash {
		t.Fatalf("after a hand edit: %+v", d)
	}
	// Reads leave no "last read" state: a read without a hash always has a body.
	r.ok(&d, "", "ticket", "get", co, "02")
	if d.Body == nil {
		t.Fatal("a cold read had no body")
	}
}

func TestTicketHistory(t *testing.T) {
	r := newRepo(t, co)
	r.setMeta(co, "01", "Phase", "review", "Attempt", "1", "Base commit", shaA, "Target commit", shaB)
	r.op(co, "01", "send-back", `{"findings":["Blocking (missing): old."],"not_blocking":["Old smell."]}`)
	r.op(co, "01", "revise", `{"reason":"Spec revision 2."}`)
	r.op(co, "01", "claim", commit(shaB))

	var cur, full docReply
	r.ok(&cur, "", "ticket", "get", co, "01")
	if cur.History != "current" || cur.OmittedComments != 1 || strings.Contains(*cur.Body, "Blocking (missing): old.") || !strings.Contains(*cur.Body, "revise (attempt 1):") {
		t.Fatalf("current history: %+v\n%s", cur, *cur.Body)
	}
	r.ok(&full, "", "ticket", "get", co, "01", "--full")
	if *full.Body != r.read(cur.Path) || full.Hash != cur.Hash || full.History != "full" {
		t.Fatal("--full must return the file bytes")
	}
	// The finding from before the revise is still unresolved for Ralph.
	if s := r.status(co); len(s.NotBlocking) != 1 || s.NotBlocking[0].Finding != "Old smell." {
		t.Fatalf("findings: %+v", s.NotBlocking)
	}
}

func TestStatusReportsStoppedTickets(t *testing.T) {
	r := newRepo(t, co)
	r.op(co, "01", "claim", commit(shaA))
	r.op(co, "01", "stop", `{"reason":"Stuck on the schema.","verification":["npm test failed"]}`)
	tk := r.status(co).Tickets[0]
	if tk.Status != "stopped" || tk.LastComment != "stop (attempt 1):\n- Reason: Stuck on the schema.\n- Verification:\n  - npm test failed" {
		t.Fatalf("%+v", tk)
	}
}

func TestReadsChangeNothing(t *testing.T) {
	r := newRepo(t, co)
	r.op(co, "01", "claim", commit(shaA))
	before := snapshot(t, r.root)
	h := r.ticket(co, "01").Hash
	for _, args := range [][]string{
		{"feature", "list"},
		{"status", co},
		{"validate", co},
		{"validate", co, "--draft"},
		{"spec", "get", co},
		{"ticket", "list", co},
		{"ticket", "get", co, "01"},
		{"ticket", "get", co, "01", "--full", "--if-changed", h},
		{"help"},
	} {
		r.ok(nil, "", args...)
	}
	// Writes that change nothing write nothing.
	r.op(co, "01", "claim", commit(shaB))
	r.ok(nil, `{"title":"Show counts"}`, "ticket", "edit", co, "01", "--if-match", h, "--input", "-")
	if after := snapshot(t, r.root); after != before {
		t.Fatalf("the checkout changed:\nbefore:\n%s\nafter:\n%s", before, after)
	}
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

package cli_test

import (
	"strings"
	"testing"
)

func TestConditionalSpecRead(t *testing.T) {
	r := newRepo(t, co)
	first := okReply[docReply](r, "", "spec", "get", co)
	if first.Body == nil || first.Revision != 1 {
		t.Fatalf("cold read: %+v", first)
	}
	if again := okReply[docReply](r, "", "spec", "get", co, "--if-changed", first.Hash); !again.Unchanged || again.Body != nil || again.Hash != first.Hash {
		t.Fatalf("conditional read: %+v", again)
	}
	// Reads leave no "last read" state: a read without a hash has a body.
	if again := okReply[docReply](r, "", "spec", "get", co); again.Unchanged || again.Body == nil {
		t.Fatalf("cold read after a conditional read: %+v", again)
	}
}

func TestConditionalTicketRead(t *testing.T) {
	r := newRepo(t, co)
	tk := r.ticket(co, "02")
	if d := okReply[docReply](r, "", "ticket", "get", co, "02", "--if-changed", tk.Hash); !d.Unchanged || d.Body != nil || d.Hash != tk.Hash {
		t.Fatalf("conditional read: %+v", d)
	}
	// A hand edit changes the hash.
	r.edit(tk.Path, "## Comments\n", "## Comments\n\nA hand-written note.\n")
	if d := okReply[docReply](r, "", "ticket", "get", co, "02", "--if-changed", tk.Hash); d.Unchanged || d.Body == nil || d.Hash == tk.Hash {
		t.Fatalf("after a hand edit: %+v", d)
	}
}

// TestHashesArePerDocument checks that a write changes only the hash of the
// document it writes, and that the spec hash follows its bytes rather than
// its revision.
func TestHashesArePerDocument(t *testing.T) {
	r := newRepo(t, co)
	spec := r.specHash(co)
	other := r.ticket(co, "02").Hash

	r.op(co, "01", "claim", commit(shaA))
	if d := okReply[docReply](r, "", "spec", "get", co, "--if-changed", spec); !d.Unchanged {
		t.Fatal("a ticket write changed the spec hash")
	}
	if r.ticket(co, "02").Hash != other {
		t.Fatal("a write to ticket 01 changed the hash of ticket 02")
	}

	ticket := r.ticket(co, "01").Hash
	r.ok(`{"attempt_limit":5}`, "spec", "edit", co, "--if-match", spec, "--input", "-")
	if d := okReply[docReply](r, "", "spec", "get", co, "--if-changed", spec); d.Unchanged || d.Body == nil || d.Revision != 1 || d.AttemptLimit != 5 {
		t.Fatalf("an attempt-limit change must change the hash but not the revision: %+v", d)
	}
	if d := okReply[docReply](r, "", "ticket", "get", co, "01", "--if-changed", ticket); !d.Unchanged {
		t.Fatal("a spec write changed a ticket hash")
	}
}

func TestTicketHistory(t *testing.T) {
	r := newRepo(t, co)
	r.setMeta(co, "01", "Phase", "review", "Attempt", "1", "Base commit", shaA, "Target commit", shaB)
	r.op(co, "01", "send-back", `{"findings":["Blocking (missing): old."],"not_blocking":["Old smell."]}`)
	r.op(co, "01", "revise", `{"reason":"Spec revision 2."}`)
	r.op(co, "01", "claim", commit(shaB))

	cur := okReply[docReply](r, "", "ticket", "get", co, "01")
	if cur.History != "current" || cur.OmittedComments != 1 || strings.Contains(*cur.Body, "Blocking (missing): old.") || !strings.Contains(*cur.Body, "revise (attempt 1):") {
		t.Fatalf("current history: %+v\n%s", cur, *cur.Body)
	}
	full := okReply[docReply](r, "", "ticket", "get", co, "01", "--full")
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
		r.ok("", args...)
	}
	// Writes that change nothing write nothing.
	r.op(co, "01", "claim", commit(shaB))
	r.ok(`{"title":"Show counts"}`, "ticket", "edit", co, "01", "--if-match", h, "--input", "-")
	if after := snapshot(t, r.root); after != before {
		t.Fatalf("the checkout changed:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

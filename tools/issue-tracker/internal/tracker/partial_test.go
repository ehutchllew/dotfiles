package tracker_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/apperr"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/store"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/tracker"
)

// failingStore fails writes to paths containing fail. It stands in for a
// disk error on one file, which can't be caused from outside the process.
type failingStore struct {
	*store.Local
	fail string
}

func (s *failingStore) Write(rel string, data []byte, create bool) error {
	if s.fail != "" && strings.Contains(rel, s.fail) {
		return apperr.New("write_failed", "disk full")
	}
	return s.Local.Write(rel, data, create)
}

// TestPartialCancelCanBeFinished checks the multi-file contract: cancel
// writes the target ticket first, reports what it couldn't write, and
// finishes when run again.
func TestPartialCancelCanBeFinished(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, ".git"), 0o755)
	src := filepath.Join("..", "cli", "testdata", "features", "checkout")
	filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		rel, _ := filepath.Rel(src, p)
		dst := filepath.Join(root, ".scratch", "checkout", rel)
		if info.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		b, _ := os.ReadFile(p)
		return os.WriteFile(dst, b, 0o644)
	})
	local, err := store.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	local.LockDir = t.TempDir()
	fs := &failingStore{Local: local, fail: "/04-"}
	tr := &tracker.Tracker{Store: fs}

	hash := func(n string) string {
		list, err := tr.ListTickets("checkout")
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range list["tickets"].([]tracker.TicketSummary) {
			if s.Ticket == n {
				return s.Hash
			}
		}
		t.Fatalf("no ticket %s", n)
		return ""
	}

	_, err = tr.Apply("checkout", 1, "cancel", hash("01"), &tracker.OpInput{Reason: "Dropped."})
	var e *apperr.Error
	if !errors.As(err, &e) || e.Code != "partial_write" {
		t.Fatalf("got %v", err)
	}
	if len(e.Details["written"].([]string)) != 3 || e.Details["pending"].([]string)[0] != ".scratch/checkout/issues/04-accept-checkout-counts.md" {
		t.Fatalf("details: %+v", e.Details)
	}

	fs.fail = ""
	r, err := tr.Apply("checkout", 1, "cancel", hash("01"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Changed) != 1 || !strings.Contains(r.Changed[0], "/04-") || r.Note == "" {
		t.Fatalf("retry: %+v", r)
	}
	b, _ := os.ReadFile(filepath.Join(root, ".scratch", "checkout", "issues", "01-show-counts.md"))
	if strings.Count(string(b), "cancel (attempt 0):") != 1 {
		t.Fatalf("the retry must not add a second comment:\n%s", b)
	}
	status, err := tr.Status("checkout")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range status.Tickets {
		for _, blocker := range s.BlockedBy {
			if blocker == "01" {
				t.Fatalf("ticket %s is still blocked by 01", s.Ticket)
			}
		}
	}
}

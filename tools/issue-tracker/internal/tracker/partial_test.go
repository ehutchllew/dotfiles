package tracker_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/apperr"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/store"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/tracker"
)

// failingStore fails writes to paths containing fail. It stands in for a
// disk error on one file, which can't be caused from outside the process.
// It records the order of the writes it was asked for.
type failingStore struct {
	*store.Local
	fail   string
	writes []string
}

func (s *failingStore) Write(rel string, data []byte, create bool) error {
	s.writes = append(s.writes, rel)
	if s.fail != "" && strings.Contains(rel, s.fail) {
		return apperr.New("write_failed", "disk full")
	}
	return s.Local.Write(rel, data, create)
}

const issues = ".scratch/checkout/issues/"

var ticketPaths = map[string]string{
	"01": issues + "01-show-counts.md",
	"02": issues + "02-phone-layout.md",
	"03": issues + "03-check-on-a-phone.md",
	"04": issues + "04-accept-checkout-counts.md",
}

func paths(ids []string) []string {
	var out []string
	for _, id := range ids {
		out = append(out, ticketPaths[id])
	}
	return out
}

// TestPartialCancelCanBeFinished checks the multi-file contract: cancel
// writes the target ticket first, reports what it couldn't write, and
// finishes when run again. Cancelling 01 rewrites 01 and then 02, 03, and
// 04, which are all blocked by it.
func TestPartialCancelCanBeFinished(t *testing.T) {
	for _, c := range []struct {
		name    string
		fail    string   // the ticket whose write fails
		written []string // tickets written before the failure
	}{
		{"the target write fails", "01", nil},
		{"the first other write fails", "02", []string{"01"}},
		{"the last write fails", "04", []string{"01", "02", "03"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
				t.Fatal(err)
			}
			src := os.DirFS(filepath.Join("..", "cli", "testdata", "features", "checkout"))
			if err := os.CopyFS(filepath.Join(root, ".scratch", "checkout"), src); err != nil {
				t.Fatal(err)
			}
			local, err := store.Open(root)
			if err != nil {
				t.Fatal(err)
			}
			local.LockDir = t.TempDir()
			fs := &failingStore{Local: local, fail: "/" + c.fail + "-"}
			tr := &tracker.Tracker{Store: fs}
			read := func(id string) string {
				b, err := os.ReadFile(filepath.Join(root, ticketPaths[id]))
				if err != nil {
					t.Fatal(err)
				}
				return string(b)
			}
			hash := func(id string) string {
				list, err := tr.ListTickets("checkout")
				if err != nil {
					t.Fatal(err)
				}
				for _, s := range list["tickets"].([]tracker.TicketSummary) {
					if s.Ticket == id {
						return s.Hash
					}
				}
				t.Fatalf("no ticket %s", id)
				return ""
			}
			all := []string{"01", "02", "03", "04"}
			before := map[string]string{}
			for _, id := range all {
				before[id] = read(id)
			}
			in := &tracker.OpInput{Reason: "Dropped."}

			_, err = tr.Apply("checkout", 1, "cancel", hash("01"), in)
			var e *apperr.Error
			if !errors.As(err, &e) {
				t.Fatalf("got %v", err)
			}
			attempted := append(slices.Clone(c.written), c.fail)
			if !slices.Equal(fs.writes, paths(attempted)) {
				t.Fatalf("writes %v, want %v", fs.writes, paths(attempted))
			}
			pending := all[len(c.written):]
			if c.written == nil {
				if e.Code != "write_failed" {
					t.Fatalf("a failed target write must fail plainly, got %v", err)
				}
			} else if e.Code != "partial_write" ||
				!slices.Equal(e.Details["written"].([]string), paths(c.written)) ||
				!slices.Equal(e.Details["pending"].([]string), paths(pending)) {
				t.Fatalf("got %v with details %+v", err, e.Details)
			}
			for _, id := range c.written {
				if read(id) == before[id] {
					t.Errorf("ticket %s is reported written but didn't change", id)
				}
			}
			for _, id := range pending {
				if read(id) != before[id] {
					t.Errorf("ticket %s is pending but changed", id)
				}
			}

			// Running the same operation again finishes it.
			fs.fail, fs.writes = "", nil
			r, err := tr.Apply("checkout", 1, "cancel", hash("01"), in)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(r.Changed, paths(pending)) {
				t.Fatalf("retry changed %v, want %v", r.Changed, paths(pending))
			}
			if c.written != nil && r.Note == "" {
				t.Error("a retry after a partial write must say the ticket was already cancelled")
			}
			if n := strings.Count(read("01"), "implement:cancel (attempt 0):"); n != 1 {
				t.Fatalf("ticket 01 has %d cancel comments:\n%s", n, read("01"))
			}
			status, err := tr.Status("checkout")
			if err != nil {
				t.Fatal(err)
			}
			for _, s := range status.Tickets {
				if s.Ticket == "01" && s.Status != "cancelled" {
					t.Errorf("ticket 01 is %s", s.Status)
				}
				if slices.Contains(s.BlockedBy, "01") {
					t.Errorf("ticket %s is still blocked by 01", s.Ticket)
				}
			}
		})
	}
}

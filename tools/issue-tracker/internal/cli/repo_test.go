package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestRepositoryDiscovery(t *testing.T) {
	r := newRepo(t)
	nested := filepath.Join(r.root, "src", "deep")
	os.MkdirAll(nested, 0o755)
	if res := r.runIn(nested, newSpec, "spec", "create", "orders", "--input", "-"); res.code != 0 {
		t.Fatal(res.raw)
	}
	if _, err := os.Stat(filepath.Join(r.root, ".scratch", "orders", "spec.md")); err != nil {
		t.Fatal("a nested invocation must write at the working-tree root")
	}

	elsewhere := t.TempDir()
	rel, _ := filepath.Rel(elsewhere, r.root)
	for _, repoFlag := range []string{r.root, rel} {
		if res := r.runIn(elsewhere, "", "--repo", repoFlag, "status", "orders"); res.code != 0 {
			t.Fatalf("--repo %s: %s", repoFlag, res.raw)
		}
	}
	if res := r.runIn(elsewhere, "", "status", "orders"); res.code != 1 || res.err(t).Code != "repo_not_found" {
		t.Fatalf("outside a repo: %s", res.raw)
	}

	// A worktree has a .git file and its own .scratch, even inside another checkout.
	wt := filepath.Join(r.root, "worktrees", "feature")
	os.MkdirAll(filepath.Join(wt, "src"), 0o755)
	os.WriteFile(filepath.Join(wt, ".git"), []byte("gitdir: ../../.git/worktrees/feature\n"), 0o644)
	if res := r.runIn(filepath.Join(wt, "src"), "", "status", "orders"); res.err(t).Code != "feature_not_found" {
		t.Fatalf("the worktree used the outer checkout's storage: %s", res.raw)
	}
	if res := r.runIn(wt, newSpec, "spec", "create", "orders", "--input", "-"); res.code != 0 {
		t.Fatal(res.raw)
	}
	if _, err := os.Stat(filepath.Join(wt, ".scratch", "orders", "spec.md")); err != nil {
		t.Fatal(err)
	}
}

func TestBackendSelection(t *testing.T) {
	for body, code := range map[string]string{
		"# Issue tracker\n\nBackend: local\n":   "",
		"# Issue tracker\n\nBackend: `local`\n": "",
		"# Issue tracker\n\nNo backend line.\n": "",
		"Backend: github\n":                     "unsupported_backend",
		"Backend:\n":                            "invalid_config",
	} {
		r := newRepo(t, co)
		r.write("docs/agents/issue-tracker.md", body)
		res := r.run("", "feature", "list")
		if code == "" {
			if res.code != 0 {
				t.Errorf("%q: %s", body, res.raw)
			}
			continue
		}
		if res.code != 1 || res.err(t).Code != code {
			t.Errorf("%q: %s", body, res.raw)
		}
	}
}

func TestUsageAndHelp(t *testing.T) {
	r := newRepo(t, co)
	for _, c := range []struct {
		args []string
		code string
	}{
		{[]string{"nope"}, "usage"},
		{[]string{"ticket", "fly", co, "01"}, "usage"},
		{[]string{"status"}, "usage"},
		{[]string{"ticket", "get", co, "x1"}, "usage"},
		{[]string{"status", co, "--if-match", "x"}, "usage"},
		{[]string{"status", co, "--bogus"}, "usage"},
		{[]string{"spec", "create", "orders"}, "usage"},
		{[]string{"ticket", "send-back", co, "01", "--if-match", "x", "--commit", shaA}, "usage"},
		{[]string{"ticket", "claim", co, "01", "--commit", shaA}, "missing_if_match"},
	} {
		res := r.run("", c.args...)
		if res.code != 2 || res.err(t).Code != c.code {
			t.Errorf("%v: exit %d %s", c.args, res.code, res.raw)
		}
	}
	if res := r.run("not json", "ticket", "create", co, "--input", "-"); res.code != 2 || res.err(t).Code != "invalid_input" {
		t.Errorf("bad JSON: %s", res.raw)
	}
	for _, args := range [][]string{{}, {"help"}, {"help", "status"}, {"ticket", "hand-off", "--help"}, {"spec", "edit", "-h"}} {
		if res := r.run("", args...); res.code != 0 || !strings.HasPrefix(res.raw, "issue-tracker") && !strings.HasPrefix(res.raw, "Usage:") {
			t.Errorf("help %v: %d %s", args, res.code, res.raw)
		}
	}
	res := r.run("", "help", "ticket", "send-back")
	if !strings.Contains(res.raw, `"criteria_unmet"`) || !strings.Contains(res.raw, "--if-match") || strings.Contains(res.raw, "--commit") {
		t.Errorf("send-back help:\n%s", res.raw)
	}

	// --input reads a file relative to the working directory.
	r.write("in.json", `{"notes":["from a file"]}`)
	r.ok(nil, "", "ticket", "comment", co, "01", "--if-match", r.ticket(co, "01").Hash, "--input", "in.json")
	r.ok(nil, "", "ticket", "claim", co, "01", "--if-match", r.ticket(co, "01").Hash, "--commit", shaA)
	r.fails("invalid_input", commit(shaA), "ticket", "claim", co, "01", "--if-match", r.ticket(co, "01").Hash, "--commit", shaA, "--input", "-")
}

// TestConcurrentProcesses runs the CLI as separate processes against one
// feature, the way Ralph and a user's shell can.
func TestConcurrentProcesses(t *testing.T) {
	r := newRepo(t, co)
	runProc := func(stdin string, args ...string) (int, string) {
		cmd := exec.Command(os.Args[0], args...)
		cmd.Dir = r.root
		cmd.Env = append(os.Environ(), "ISSUE_TRACKER_TEST_EXEC=1")
		cmd.Stdin = strings.NewReader(stdin)
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		code := 0
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else if err != nil {
			t.Error(err)
		}
		return code, out.String()
	}
	parallel := func(n int, f func(i int)) {
		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func(i int) { defer wg.Done(); f(i) }(i)
		}
		wg.Wait()
	}

	h := r.ticket(co, "01").Hash
	var mu sync.Mutex
	codes := map[string]int{}
	parallel(8, func(i int) {
		code, out := runProc(fmt.Sprintf(`{"notes":["note %d"]}`, i), "ticket", "comment", co, "01", "--if-match", h, "--input", "-")
		key := "ok"
		if code != 0 {
			var e struct{ Error apiError }
			json.Unmarshal([]byte(out), &e)
			key = e.Error.Code
		}
		mu.Lock()
		codes[key]++
		mu.Unlock()
	})
	if codes["ok"] != 1 || codes["stale"] != 7 {
		t.Fatalf("with one hash, exactly one write must win: %v", codes)
	}

	numbers := map[string]bool{}
	parallel(10, func(i int) {
		code, out := runProc(ticketJSON(fmt.Sprintf("Work %d", i), "implementation", `["US 1"]`, `[]`), "ticket", "create", co, "--input", "-")
		var w writeReply
		json.Unmarshal([]byte(out), &w)
		mu.Lock()
		defer mu.Unlock()
		if code != 0 || numbers[w.Ticket] {
			t.Errorf("create %d: exit %d, ticket %q", i, code, w.Ticket)
		}
		numbers[w.Ticket] = true
	})
	for n := 5; n <= 14; n++ {
		if !numbers[fmt.Sprintf("%02d", n)] {
			t.Fatalf("ticket numbers: %v", numbers)
		}
	}
}

func TestFailedWriteChangesNothing(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	r := newRepo(t, co)
	issues := filepath.Join(r.root, ".scratch", co, "issues")
	before := snapshot(t, r.root)
	h := r.ticket(co, "01").Hash
	os.Chmod(issues, 0o555)
	defer os.Chmod(issues, 0o755)
	r.fails("write_failed", "", "ticket", "claim", co, "01", "--if-match", h, "--commit", shaA)
	os.Chmod(issues, 0o755)
	filepath.WalkDir(r.root, func(p string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(p, ".tmp") {
			t.Errorf("temp file left: %s", p)
		}
		return nil
	})
	if snapshot(t, r.root) != before {
		t.Fatal("a failed write changed the checkout")
	}
}

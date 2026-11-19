package cli_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/cli"
)

func TestRepositoryDiscovery(t *testing.T) {
	r := newRepo(t)
	nested := filepath.Join(r.root, "src", "deep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if res := r.runIn(nested, newSpec, "spec", "create", "orders", "--input", "-"); res.code != 0 {
		t.Fatal(res.raw)
	}
	if _, err := os.Stat(filepath.Join(r.root, ".scratch", "orders", "spec.md")); err != nil {
		t.Fatal("a nested invocation must write at the working-tree root")
	}

	elsewhere := t.TempDir()
	rel, err := filepath.Rel(elsewhere, r.root)
	if err != nil {
		t.Fatal(err)
	}
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
	r.write("worktrees/feature/.git", "gitdir: ../../.git/worktrees/feature\n")
	if err := os.Mkdir(filepath.Join(wt, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
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
	for _, c := range []struct {
		name string
		body string
		err  string
	}{
		{"local", "# Issue tracker\n\nBackend: local\n", ""},
		{"local in backticks", "# Issue tracker\n\nBackend: `local`\n", ""},
		{"no backend line", "# Issue tracker\n\nNo backend line.\n", ""},
		{"another backend", "Backend: github\n", "unsupported_backend"},
		{"empty backend", "Backend:\n", "invalid_config"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := newRepo(t, co)
			r.write("docs/agents/issue-tracker.md", c.body)
			if c.err == "" {
				r.ok("", "feature", "list")
			} else {
				r.fails(1, c.err, "", "feature", "list")
			}
		})
	}
}

func TestUsageErrors(t *testing.T) {
	for _, c := range []struct {
		name  string
		args  []string
		stdin string
		err   string
	}{
		{"unknown command", []string{"nope"}, "", "usage"},
		{"unknown operation", []string{"ticket", "fly", co, "01"}, "", "usage"},
		{"missing feature", []string{"status"}, "", "usage"},
		{"bad ticket number", []string{"ticket", "get", co, "x1"}, "", "usage"},
		{"flag for another command", []string{"status", co, "--if-match", "x"}, "", "usage"},
		{"unknown flag", []string{"status", co, "--bogus"}, "", "usage"},
		{"missing --input", []string{"spec", "create", "orders"}, "", "usage"},
		{"--commit for an operation without a commit", []string{"ticket", "send-back", co, "01", "--if-match", "x", "--commit", shaA}, "", "usage"},
		{"missing --if-match", []string{"ticket", "claim", co, "01", "--commit", shaA}, "", "missing_if_match"},
		{"input that isn't JSON", []string{"ticket", "create", co, "--input", "-"}, "not json", "invalid_input"},
	} {
		t.Run(c.name, func(t *testing.T) {
			newRepo(t, co).fails(2, c.err, c.stdin, c.args...)
		})
	}
}

func TestHelp(t *testing.T) {
	for _, c := range []struct {
		name   string
		args   []string
		prefix string
	}{
		{"no arguments", nil, "issue-tracker reads and writes"},
		{"help", []string{"help"}, "issue-tracker reads and writes"},
		{"help for a command", []string{"help", "status"}, "Usage: issue-tracker status FEATURE"},
		{"--help", []string{"ticket", "hand-off", "--help"}, "Usage: issue-tracker ticket hand-off"},
		{"-h", []string{"spec", "edit", "-h"}, "Usage: issue-tracker spec edit"},
	} {
		t.Run(c.name, func(t *testing.T) {
			res := newRepo(t, co).ok("", c.args...)
			if !strings.HasPrefix(res.raw, c.prefix) {
				t.Fatalf("got:\n%s", res.raw)
			}
		})
	}

	// Operation help shows the operation's own input and flags.
	res := newRepo(t, co).ok("", "help", "ticket", "send-back")
	if !strings.Contains(res.raw, `"criteria_unmet"`) || !strings.Contains(res.raw, "--if-match") || strings.Contains(res.raw, "--commit") {
		t.Errorf("send-back help:\n%s", res.raw)
	}
}

func TestInputFromFile(t *testing.T) {
	r := newRepo(t, co)
	// The path is relative to the working directory, not the repository root.
	r.write("src/in.json", `{"notes":["from a file"]}`)
	h := r.ticket(co, "01").Hash
	if res := r.runIn(filepath.Join(r.root, "src"), "", "ticket", "comment", co, "01", "--if-match", h, "--input", "in.json"); res.code != 0 {
		t.Fatal(res.raw)
	}
	if c := r.lastComment(co, "01"); c != "implement:comment (attempt 0):\n- Note: from a file" {
		t.Fatalf("comment %q", c)
	}
	r.fails(2, "invalid_input", "", "ticket", "comment", co, "01", "--if-match", r.ticket(co, "01").Hash, "--input", "missing.json")
}

func TestCommitFlag(t *testing.T) {
	r := newRepo(t, co)
	r.fails(2, "invalid_input", commit(shaA), "ticket", "claim", co, "01", "--if-match", r.ticket(co, "01").Hash, "--commit", shaA, "--input", "-")
	r.ok("", "ticket", "claim", co, "01", "--if-match", r.ticket(co, "01").Hash, "--commit", shaA)
	if tk := r.ticket(co, "01"); tk.Base != shaA {
		t.Fatalf("base %q", tk.Base)
	}
}

// TestInvalidCommitExitCodes covers the one error code that comes with
// either exit code. A ref that looks like a flag is bad input, but a ref
// git can't resolve is a failure of the command.
func TestInvalidCommitExitCodes(t *testing.T) {
	for _, c := range []struct {
		name string
		ref  string
		exit int
	}{
		{"ref that looks like a flag", "-bad", 2},
		{"ref git can't resolve", "no-such-ref", 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			newRepo(t, co).opFails(c.exit, "invalid_commit", co, "01", "claim", commit(c.ref))
		})
	}
}

// TestInputMustBeOneJSONObject uses cancel because it accepts an empty
// object, so only the shape of the input decides the outcome.
func TestInputMustBeOneJSONObject(t *testing.T) {
	for _, c := range []struct {
		name  string
		input string
		err   string
	}{
		{"object", `{"reason":"Dropped."}`, ""},
		{"empty object", `{}`, ""},
		{"null", `null`, "invalid_input"},
		{"array", `[]`, "invalid_input"},
		{"string", `"Dropped."`, "invalid_input"},
		{"number", `1`, "invalid_input"},
		{"empty input", ``, "invalid_input"},
		{"malformed", `{"reason":`, "invalid_input"},
		{"data after the object", `{} {}`, "invalid_input"},
		{"unknown field", `{"reasn":"Dropped."}`, "invalid_input"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := newRepo(t, co)
			args := []string{"ticket", "cancel", co, "03", "--if-match", r.ticket(co, "03").Hash, "--input", "-"}
			if c.err == "" {
				r.ok(c.input, args...)
				if s := r.ticket(co, "03").Status; s != "cancelled" {
					t.Fatalf("status %s", s)
				}
				return
			}
			before := snapshot(t, r.root)
			r.fails(2, c.err, c.input, args...)
			if snapshot(t, r.root) != before {
				t.Fatal("rejected input changed the checkout")
			}
		})
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

// TestOutputFailures checks that a reply that can't be written fails the
// command and is reported on stderr.
func TestOutputFailures(t *testing.T) {
	for _, c := range []struct {
		name   string
		args   []string
		exit   int
		stderr string
	}{
		{"help", []string{"help"}, 1, "writing its response to stdout failed: disk full"},
		{"JSON reply", []string{"feature", "list"}, 1, "writing its response to stdout failed: disk full"},
		{"usage error reply", []string{"nope"}, 2, `usage: unknown command "nope"`},
		{"error reply", []string{"status", "missing"}, 1, "feature_not_found:"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := newRepo(t, co)
			var errOut bytes.Buffer
			code := cli.Run(c.args, cli.Env{Dir: r.root, Stdin: strings.NewReader(""), Stdout: failingWriter{}, Stderr: &errOut})
			if code != c.exit || !strings.Contains(errOut.String(), c.stderr) || !strings.Contains(errOut.String(), "disk full") {
				t.Fatalf("exit %d, stderr %q", code, errOut.String())
			}
		})
	}
}

// TestOutputFailureAfterAWrite checks that a write whose reply can't be
// printed still fails, and that the write itself isn't undone.
func TestOutputFailureAfterAWrite(t *testing.T) {
	r := newRepo(t, co)
	args := []string{"ticket", "comment", co, "01", "--if-match", r.ticket(co, "01").Hash, "--input", "-"}
	var errOut bytes.Buffer
	code := cli.Run(args, cli.Env{Dir: r.root, Stdin: strings.NewReader(`{"notes":["kept"]}`), Stdout: failingWriter{}, Stderr: &errOut})
	if code != 1 || !strings.Contains(errOut.String(), "the command ran") {
		t.Fatalf("exit %d, stderr %q", code, errOut.String())
	}
	if c := r.lastComment(co, "01"); c != "implement:comment (attempt 0):\n- Note: kept" {
		t.Fatalf("comment %q", c)
	}
}

func newGitRepo(t *testing.T) (*repo, func(...string) string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git isn't installed")
	}
	// Keep the user's git configuration out of both this test's git calls
	// and the CLI's.
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	r := newRepo(t, co)
	if err := os.Remove(filepath.Join(r.root, ".git")); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.CommandContext(t.Context(), "git", append([]string{"-C", r.root}, args...)...)
		var errOut bytes.Buffer
		cmd.Stderr = &errOut
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, errOut.String())
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q", "--template=")
	// Local settings in the temporary repository: no hooks and no global
	// ignore file.
	git("config", "core.hooksPath", os.DevNull)
	git("config", "core.excludesFile", os.DevNull)
	git("config", "user.name", "Test")
	git("config", "user.email", "test@example.com")
	return r, git
}

func TestGitRefsAndIgnoredStorage(t *testing.T) {
	r, git := newGitRepo(t)
	git("commit", "-q", "--allow-empty", "--no-verify", "-m", "first")
	first := git("rev-parse", "HEAD")
	git("commit", "-q", "--allow-empty", "--no-verify", "-m", "second")
	second := git("rev-parse", "HEAD")

	w := okReply[writeReply](r, "", "ticket", "claim", co, "01", "--if-match", r.ticket(co, "01").Hash, "--commit", "HEAD~1")
	if w.GitIgnored == nil || *w.GitIgnored {
		t.Fatalf("storage isn't ignored yet: %+v", w)
	}
	r.op(co, "01", "hand-off", `{"commit":"HEAD","verification":["ok"]}`)
	if tk := r.ticket(co, "01"); tk.Base != first || tk.Target != second {
		t.Fatalf("base %s target %s, want %s and %s", tk.Base, tk.Target, first, second)
	}

	r.opFails(1, "invalid_commit", co, "04", "set-feature-base", commit("no-such-ref"))

	r.write(".gitignore", ".scratch/\n")
	if w := r.op(co, "04", "set-feature-base", commit("HEAD")); w.GitIgnored == nil || !*w.GitIgnored {
		t.Fatalf("storage is ignored now: %+v", w)
	}
	if tk := r.ticket(co, "04"); tk.Base != second {
		t.Fatalf("base %s, want %s", tk.Base, second)
	}
}

// runProcess runs the CLI as a separate process in dir. Goroutines can call
// it at the same time.
func runProcess(t *testing.T, dir, stdin string, args ...string) result {
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "ISSUE_TRACKER_TEST_EXEC=1")
	cmd.Stdin = strings.NewReader(stdin)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	var exit *exec.ExitError
	if err := cmd.Run(); err != nil && !errors.As(err, &exit) {
		t.Errorf("%v: %v", args, err)
	}
	return result{code: cmd.ProcessState.ExitCode(), raw: out.String(), stderr: errOut.String()}
}

// TestConcurrentProcesses runs the CLI as separate processes against one
// feature, the way Ralph and a user's shell can.
func TestConcurrentProcesses(t *testing.T) {
	r := newRepo(t, co)
	// runAll runs one process per stdin at the same time.
	runAll := func(stdins []string, args ...string) []result {
		results := make([]result, len(stdins))
		var wg sync.WaitGroup
		for i, stdin := range stdins {
			wg.Go(func() { results[i] = runProcess(t, r.root, stdin, args...) })
		}
		wg.Wait()
		return results
	}

	notes := make([]string, 8)
	for i := range notes {
		notes[i] = fmt.Sprintf(`{"notes":["note %d"]}`, i)
	}
	codes := map[string]int{}
	for _, res := range runAll(notes, "ticket", "comment", co, "01", "--if-match", r.ticket(co, "01").Hash, "--input", "-") {
		if res.code == 0 {
			codes["ok"]++
		} else {
			codes[res.err(t).Code]++
		}
	}
	if codes["ok"] != 1 || codes["stale"] != 7 {
		t.Fatalf("with one hash, exactly one write must win: %v", codes)
	}

	tickets := make([]string, 10)
	for i := range tickets {
		tickets[i] = ticketJSON(fmt.Sprintf("Work %d", i), "implementation", `["US 1"]`, `[]`)
	}
	numbers := map[string]bool{}
	for _, res := range runAll(tickets, "ticket", "create", co, "--input", "-") {
		if res.code != 0 {
			t.Fatalf("create: exit %d: %s %s", res.code, res.raw, res.stderr)
		}
		numbers[decode[writeReply](t, res).Ticket] = true
	}
	for n := 5; n <= 14; n++ {
		if !numbers[fmt.Sprintf("%02d", n)] {
			t.Fatalf("each create must get its own number: %v", numbers)
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
	if err := os.Chmod(issues, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(issues, 0o755) })
	r.fails(1, "write_failed", "", "ticket", "claim", co, "01", "--if-match", h, "--commit", shaA)
	if err := os.Chmod(issues, 0o755); err != nil {
		t.Fatal(err)
	}
	// The snapshot lists every path, so a leftover temporary file fails too.
	if snapshot(t, r.root) != before {
		t.Fatal("a failed write changed the checkout")
	}
}

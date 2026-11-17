package cli_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/cli"
)

var update = flag.Bool("update", false, "rewrite golden files")

// TestMain lets a test run the CLI as a separate process: the test binary
// runs it when ISSUE_TRACKER_TEST_EXEC is set.
func TestMain(m *testing.M) {
	if os.Getenv("ISSUE_TRACKER_TEST_EXEC") == "1" {
		dir, _ := os.Getwd()
		os.Exit(cli.Run(os.Args[1:], cli.Env{Dir: dir, Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}))
	}
	os.Exit(m.Run())
}

var (
	shaA = strings.Repeat("a", 40)
	shaB = strings.Repeat("b", 40)
	shaC = strings.Repeat("c", 40)
)

// repo is a working tree in a temp dir. Its .git directory is empty: the
// CLI only looks for the entry to find the root.
type repo struct {
	t    *testing.T
	root string
}

func newRepo(t *testing.T, features ...string) *repo {
	t.Helper()
	root, _ := filepath.EvalSymlinks(t.TempDir())
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ISSUE_TRACKER_LOCK_DIR", t.TempDir())
	r := &repo{t: t, root: root}
	for _, f := range features {
		r.copyFeature(f)
	}
	return r
}

func (r *repo) copyFeature(name string) {
	r.t.Helper()
	src := filepath.Join("testdata", "features", name)
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		dst := filepath.Join(r.root, ".scratch", name, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, b, 0o644)
	})
	if err != nil {
		r.t.Fatal(err)
	}
}

type result struct {
	code int
	raw  string
}

// decode replaces v with the reply, so fields from an earlier reply decoded
// into the same variable don't survive.
func (res result) decode(t *testing.T, v any) {
	t.Helper()
	target := reflect.ValueOf(v).Elem()
	target.Set(reflect.Zero(target.Type()))
	if err := json.Unmarshal([]byte(res.raw), v); err != nil {
		t.Fatalf("bad JSON %q: %v", res.raw, err)
	}
}

type apiError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

func (res result) err(t *testing.T) apiError {
	t.Helper()
	var out struct {
		Error *apiError `json:"error"`
	}
	res.decode(t, &out)
	if out.Error == nil {
		t.Fatalf("expected an error, got %s", res.raw)
	}
	return *out.Error
}

// runIn runs the CLI in dir with stdin.
func (r *repo) runIn(dir, stdin string, args ...string) result {
	var out, errOut bytes.Buffer
	code := cli.Run(args, cli.Env{Dir: dir, Stdin: strings.NewReader(stdin), Stdout: &out, Stderr: &errOut})
	return result{code: code, raw: out.String()}
}

func (r *repo) run(stdin string, args ...string) result {
	return r.runIn(r.root, stdin, args...)
}

// ok runs a command that must succeed and decodes its reply into v.
func (r *repo) ok(v any, stdin string, args ...string) {
	r.t.Helper()
	res := r.run(stdin, args...)
	if res.code != 0 {
		r.t.Fatalf("%v: exit %d: %s", args, res.code, res.raw)
	}
	if v != nil {
		res.decode(r.t, v)
	}
}

// fails runs a command that must fail with code and returns the error.
func (r *repo) fails(code string, stdin string, args ...string) apiError {
	r.t.Helper()
	res := r.run(stdin, args...)
	if res.code == 0 {
		r.t.Fatalf("%v: expected %s, got success: %s", args, code, res.raw)
	}
	e := res.err(r.t)
	if e.Code != code {
		r.t.Fatalf("%v: got %s (%s), want %s", args, e.Code, e.Message, code)
	}
	return e
}

type writeReply struct {
	Ticket        string            `json:"ticket"`
	Operation     string            `json:"operation"`
	Unchanged     bool              `json:"unchanged"`
	Hash          string            `json:"hash"`
	Revision      int               `json:"revision"`
	Phase         string            `json:"phase"`
	Status        string            `json:"status"`
	Attempt       *int              `json:"attempt"`
	Changed       []string          `json:"changed"`
	Hashes        map[string]string `json:"hashes"`
	CommitSubject string            `json:"commit_subject"`
	Note          string            `json:"note"`
}

type ticketSummary struct {
	Ticket       string   `json:"ticket"`
	Title        string   `json:"title"`
	Kind         string   `json:"kind"`
	Phase        string   `json:"phase"`
	Status       string   `json:"status"`
	Attempt      *int     `json:"attempt"`
	BlockedBy    []string `json:"blocked_by"`
	Covers       []string `json:"covers"`
	SpecRevision int      `json:"spec_revision"`
	Base         string   `json:"base"`
	Target       string   `json:"target"`
	Ready        bool     `json:"ready"`
	Hash         string   `json:"hash"`
	Path         string   `json:"path"`
	LastComment  string   `json:"last_comment"`
}

type problem struct {
	Code    string `json:"code"`
	Ticket  string `json:"ticket"`
	Message string `json:"message"`
}

type statusReply struct {
	Title                 string          `json:"title"`
	Revision              int             `json:"revision"`
	AttemptLimit          int             `json:"attempt_limit"`
	SpecHash              string          `json:"spec_hash"`
	Valid                 bool            `json:"valid"`
	Errors                []problem       `json:"errors"`
	CurrentAcceptance     string          `json:"current_acceptance"`
	EarlierAcceptanceBase string          `json:"earlier_acceptance_base"`
	Tickets               []ticketSummary `json:"tickets"`
	NotBlocking           []struct {
		Ticket  string `json:"ticket"`
		Finding string `json:"finding"`
	} `json:"not_blocking"`
}

type docReply struct {
	Hash            string  `json:"hash"`
	Unchanged       bool    `json:"unchanged"`
	Path            string  `json:"path"`
	Revision        int     `json:"revision"`
	AttemptLimit    int     `json:"attempt_limit"`
	History         string  `json:"history"`
	OmittedComments int     `json:"omitted_comments"`
	Body            *string `json:"body"`
}

func (r *repo) status(feature string) statusReply {
	r.t.Helper()
	var s statusReply
	r.ok(&s, "", "status", feature)
	return s
}

func (r *repo) ticket(feature, n string) ticketSummary {
	r.t.Helper()
	var list struct {
		Tickets []ticketSummary `json:"tickets"`
	}
	r.ok(&list, "", "ticket", "list", feature)
	for _, t := range list.Tickets {
		if t.Ticket == n {
			return t
		}
	}
	r.t.Fatalf("no ticket %s", n)
	return ticketSummary{}
}

func (r *repo) specHash(feature string) string {
	r.t.Helper()
	var d docReply
	r.ok(&d, "", "spec", "get", feature)
	return d.Hash
}

// op runs a state operation with the ticket's current hash.
func (r *repo) op(feature, n, name, input string) writeReply {
	r.t.Helper()
	var w writeReply
	r.ok(&w, input, r.opArgs(feature, n, name, input)...)
	return w
}

// opFails runs a state operation that must fail, and checks that the
// ticket file didn't change.
func (r *repo) opFails(code, feature, n, name, input string) apiError {
	r.t.Helper()
	path := r.ticket(feature, n).Path
	before := r.read(path)
	e := r.fails(code, input, r.opArgs(feature, n, name, input)...)
	if r.read(path) != before {
		r.t.Fatalf("a failed %s changed the ticket", name)
	}
	return e
}

func (r *repo) opArgs(feature, n, name, input string) []string {
	args := []string{"ticket", name, feature, n, "--if-match", r.ticket(feature, n).Hash}
	if input != "" {
		args = append(args, "--input", "-")
	}
	return args
}

func (r *repo) read(rel string) string {
	r.t.Helper()
	b, err := os.ReadFile(filepath.Join(r.root, rel))
	if err != nil {
		r.t.Fatal(err)
	}
	return string(b)
}

func (r *repo) write(rel, content string) {
	r.t.Helper()
	p := filepath.Join(r.root, rel)
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

// edit replaces text in a file, as a person editing it by hand would.
func (r *repo) edit(rel, old, new string) {
	r.t.Helper()
	s := r.read(rel)
	if !strings.Contains(s, old) {
		r.t.Fatalf("%s doesn't contain %q", rel, old)
	}
	r.write(rel, strings.Replace(s, old, new, 1))
}

// setMeta sets metadata lines on a ticket file by hand. An empty value
// removes the line.
func (r *repo) setMeta(feature, n string, kv ...string) {
	r.t.Helper()
	path := r.ticket(feature, n).Path
	s := r.read(path)
	for i := 0; i < len(kv); i += 2 {
		line := ""
		if kv[i+1] != "" {
			line = kv[i] + ": " + kv[i+1] + "\n"
		}
		re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(kv[i]) + `:.*\n`)
		if loc := re.FindStringIndex(s); loc != nil {
			s = s[:loc[0]] + line + s[loc[1]:]
		} else if line != "" {
			end := strings.Index(s, "\n\n## ")
			s = s[:end+1] + line + s[end+1:]
		}
	}
	r.write(path, s)
}

// lastComment returns the last comment in a ticket's Comments section.
func (r *repo) lastComment(feature, n string) string {
	r.t.Helper()
	body := r.read(r.ticket(feature, n).Path)
	comments := body[strings.Index(body, "## Comments\n")+len("## Comments\n"):]
	head := regexp.MustCompile(`(?m)^[a-z ]+( \(attempt \d+\))?:$`)
	locs := head.FindAllStringIndex(comments, -1)
	if len(locs) == 0 {
		return ""
	}
	return strings.TrimRight(comments[locs[len(locs)-1][0]:], "\n")
}

// snapshot describes every path under root with its mode, size, mtime, and
// content, so a test can prove nothing changed.
func snapshot(t *testing.T, root string) string {
	t.Helper()
	var b strings.Builder
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Fatal(err)
		}
		info, _ := d.Info()
		rel, _ := filepath.Rel(root, p)
		fmt.Fprintf(&b, "%s %v %d %d", rel, info.Mode(), info.Size(), info.ModTime().UnixNano())
		if d.Type().IsRegular() {
			data, _ := os.ReadFile(p)
			fmt.Fprintf(&b, " %q", data)
		}
		b.WriteString("\n")
		return nil
	})
	return b.String()
}

// golden compares got with testdata/golden/name, or rewrites it with -update.
func golden(t *testing.T, name, got string) {
	t.Helper()
	p := filepath.Join("testdata", "golden", name)
	if *update {
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("%v; run go test with -update to create it", err)
	}
	if got != string(want) {
		t.Fatalf("%s differs from the golden file.\ngot:\n%s\nwant:\n%s", name, got, want)
	}
}

func intp(n int) *int { return &n }

func attempt(t ticketSummary) int {
	if t.Attempt == nil {
		return -1
	}
	return *t.Attempt
}

// Package cli parses arguments, runs commands, and prints JSON.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/apperr"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/doc"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/store"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/tracker"
)

// Env is what a command run can see. Tests replace each part.
type Env struct {
	Dir    string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type args struct {
	pos   []string
	flags map[string]string
}

// flag definitions: true when the flag takes a value.
var globalFlags = map[string]bool{"repo": true, "pretty": false, "help": false}

func parse(raw []string, defs map[string]bool) (*args, error) {
	a := &args{flags: map[string]string{}}
	for i := 0; i < len(raw); i++ {
		s := raw[i]
		if s == "--" {
			a.pos = append(a.pos, raw[i+1:]...)
			break
		}
		if s == "-h" {
			s = "--help"
		}
		if !strings.HasPrefix(s, "--") || s == "-" {
			a.pos = append(a.pos, s)
			continue
		}
		name, value, hasValue := strings.Cut(s[2:], "=")
		takes, ok := defs[name]
		if !ok {
			takes, ok = globalFlags[name]
		}
		if !ok {
			return nil, apperr.Usage("usage", "unknown flag --%s", name)
		}
		if _, dup := a.flags[name]; dup {
			return nil, apperr.Usage("usage", "--%s is given twice", name)
		}
		switch {
		case takes && hasValue:
		case takes:
			if i+1 >= len(raw) {
				return nil, apperr.Usage("usage", "--%s needs a value", name)
			}
			i++
			value = raw[i]
		case hasValue:
			return nil, apperr.Usage("usage", "--%s takes no value", name)
		default:
			value = "true"
		}
		a.flags[name] = value
	}
	return a, nil
}

// Run executes one command and returns the exit code.
func Run(argv []string, env Env) int {
	pretty := false
	for _, a := range argv {
		if a == "--" {
			break
		}
		if a == "--pretty" {
			pretty = true
		}
	}
	out, err := run(argv, env)
	if err != nil {
		var e *apperr.Error
		if !errors.As(err, &e) {
			e = apperr.New("internal_error", "%v", err)
		}
		code := 1
		if e.Usage {
			code = 2
		}
		if werr := write(env.Stdout, map[string]any{"error": e}, pretty); werr != nil {
			fmt.Fprintf(env.Stderr, "issue-tracker: %v; writing this error to stdout also failed: %v\n", e, werr)
		}
		return code
	}
	if werr := write(env.Stdout, out, pretty); werr != nil {
		// A write command has already changed its files at this point.
		fmt.Fprintf(env.Stderr, "issue-tracker: the command ran, but writing its response to stdout failed: %v\n", werr)
		return 1
	}
	return 0
}

type helpText string

func write(w io.Writer, v any, pretty bool) error {
	if s, ok := v.(helpText); ok {
		_, err := io.WriteString(w, string(s))
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(v)
}

// commandFlags lists each command's own flags.
var commandFlags = map[string]map[string]bool{
	"feature list":  {},
	"status":        {},
	"validate":      {"draft": false},
	"spec get":      {"if-changed": true},
	"spec create":   {"input": true},
	"spec edit":     {"input": true, "if-match": true},
	"ticket list":   {},
	"ticket get":    {"if-changed": true, "full": false},
	"ticket create": {"input": true},
	"ticket edit":   {"input": true, "if-match": true},
}

var opFlags = map[string]bool{"input": true, "if-match": true, "commit": true}

// commandOf finds the command words at the start of the positional args.
func commandOf(pos []string) (string, int) {
	if len(pos) == 0 {
		return "", 0
	}
	switch pos[0] {
	case "status", "validate", "help":
		return pos[0], 1
	case "feature", "spec", "ticket":
		if len(pos) < 2 {
			return pos[0], 1
		}
		return pos[0] + " " + pos[1], 2
	}
	return pos[0], 1
}

func run(argv []string, env Env) (any, error) {
	all := map[string]bool{}
	for _, defs := range commandFlags {
		maps.Copy(all, defs)
	}
	maps.Copy(all, opFlags)
	a, err := parse(argv, all)
	if err != nil {
		return nil, err
	}
	cmd, n := commandOf(a.pos)
	if cmd == "" || cmd == "help" {
		h, ok := helpFor(a.pos[min(n, len(a.pos)):])
		if !ok {
			return nil, apperr.Usage("usage", "no help for %q; commands: %s", strings.Join(a.pos[n:], " "), strings.Join(knownCommands(), "; "))
		}
		return helpText(h), nil
	}
	words := strings.Fields(cmd)
	if _, ok := a.flags["help"]; ok {
		h, ok := helpFor(words)
		if !ok {
			return nil, apperr.Usage("usage", "unknown command %q; run issue-tracker help", cmd)
		}
		return helpText(h), nil
	}

	defs, ok := commandFlags[cmd]
	isOp := len(words) == 2 && words[0] == "ticket" && tracker.IsOperation(words[1])
	if isOp {
		defs, ok = opFlags, true
	}
	if !ok {
		return nil, apperr.Usage("usage", "unknown command %q; run issue-tracker help", cmd)
	}
	for k := range a.flags {
		if _, ok := defs[k]; !ok {
			if _, global := globalFlags[k]; !global {
				return nil, apperr.Usage("usage", "%s doesn't take --%s; run issue-tracker help %s", cmd, k, cmd)
			}
		}
	}
	rest := a.pos[n:]

	tr, err := open(a.flags["repo"], env.Dir)
	if err != nil {
		return nil, err
	}

	need := func(names ...string) error {
		if len(rest) != len(names) {
			return apperr.Usage("usage", "usage: issue-tracker %s %s", cmd, strings.ToUpper(strings.Join(names, " ")))
		}
		return nil
	}
	ticketNum := func(s string) (int, error) {
		n, err := doc.ParseTicketNumber(s)
		if err != nil {
			return 0, apperr.Usage("usage", "%v", err)
		}
		return n, nil
	}
	feature := func() (string, error) {
		if err := store.CheckFeatureID(rest[0]); err != nil {
			return "", err
		}
		return rest[0], nil
	}

	switch {
	case cmd == "feature list":
		if err := need(); err != nil {
			return nil, err
		}
		list, err := tr.ListFeatures()
		if err != nil {
			return nil, err
		}
		return map[string]any{"features": list}, nil

	case cmd == "status", cmd == "validate", cmd == "spec get", cmd == "ticket list", cmd == "spec create", cmd == "ticket create":
		if err := need("feature"); err != nil {
			return nil, err
		}
		id, err := feature()
		if err != nil {
			return nil, err
		}
		switch cmd {
		case "status":
			return tr.Status(id)
		case "validate":
			_, draft := a.flags["draft"]
			return tr.Validate(id, !draft)
		case "spec get":
			return tr.SpecGet(id, a.flags["if-changed"])
		case "ticket list":
			return tr.ListTickets(id)
		case "spec create":
			data, err := readInput(a.flags, env, true)
			if err != nil {
				return nil, err
			}
			in, err := tracker.DecodeSpecCreate(data)
			if err != nil {
				return nil, err
			}
			return tr.SpecCreate(id, in)
		default:
			data, err := readInput(a.flags, env, true)
			if err != nil {
				return nil, err
			}
			in, err := tracker.DecodeTicketCreate(data)
			if err != nil {
				return nil, err
			}
			return tr.TicketCreate(id, in)
		}

	case cmd == "spec edit":
		if err := need("feature"); err != nil {
			return nil, err
		}
		id, err := feature()
		if err != nil {
			return nil, err
		}
		data, err := readInput(a.flags, env, true)
		if err != nil {
			return nil, err
		}
		in, err := tracker.DecodeSpecEdit(data)
		if err != nil {
			return nil, err
		}
		return tr.SpecEdit(id, a.flags["if-match"], in)

	case cmd == "ticket get", cmd == "ticket edit", isOp:
		if err := need("feature", "nn"); err != nil {
			return nil, err
		}
		id, err := feature()
		if err != nil {
			return nil, err
		}
		num, err := ticketNum(rest[1])
		if err != nil {
			return nil, err
		}
		switch {
		case cmd == "ticket get":
			_, full := a.flags["full"]
			return tr.TicketGet(id, num, a.flags["if-changed"], full)
		case cmd == "ticket edit":
			data, err := readInput(a.flags, env, true)
			if err != nil {
				return nil, err
			}
			in, err := tracker.DecodeTicketEdit(data)
			if err != nil {
				return nil, err
			}
			return tr.TicketEdit(id, num, a.flags["if-match"], in)
		default:
			data, err := readInput(a.flags, env, false)
			if err != nil {
				return nil, err
			}
			in, err := tracker.DecodeOp(words[1], data)
			if err != nil {
				return nil, err
			}
			if c, ok := a.flags["commit"]; ok {
				if !slices.Contains(tracker.OperationFields(words[1]), "commit") {
					return nil, apperr.Usage("usage", "%s doesn't take a commit", words[1])
				}
				if in.Commit != "" {
					return nil, apperr.Usage("invalid_input", "give the commit in --commit or in the input, not both")
				}
				in.Commit = c
			}
			return tr.Apply(id, num, words[1], a.flags["if-match"], in)
		}
	}
	return nil, apperr.Usage("usage", "unknown command %q; run issue-tracker help", cmd)
}

func open(repo, dir string) (*tracker.Tracker, error) {
	root := repo
	if root == "" {
		var err error
		root, err = store.FindRepo(dir)
		if err != nil {
			return nil, err
		}
	} else if !filepath.IsAbs(root) {
		root = filepath.Join(dir, root)
	}
	l, err := store.Open(root)
	if err != nil {
		return nil, err
	}
	return &tracker.Tracker{Store: l}, nil
}

func readInput(flags map[string]string, env Env, required bool) ([]byte, error) {
	path, ok := flags["input"]
	if !ok {
		if required {
			return nil, apperr.Usage("usage", "this command needs --input FILE, or --input - for stdin")
		}
		return nil, nil
	}
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(env.Stdin)
	} else {
		if !filepath.IsAbs(path) {
			path = filepath.Join(env.Dir, path)
		}
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, apperr.Usage("invalid_input", "can't read input: %v", err)
	}
	return data, nil
}

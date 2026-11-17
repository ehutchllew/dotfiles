package doc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DefaultAttemptLimit applies when a spec has no `Attempt limit:` line.
const DefaultAttemptLimit = 3

// SpecMetaOrder is the canonical order of spec metadata lines.
var SpecMetaOrder = []string{"Revision", "Attempt limit"}

// SpecSectionOrder is the canonical order of spec sections.
var SpecSectionOrder = []string{"Problem statement", "Solution", "User stories", "Implementation decisions", "Testing decisions", "Out of scope", "Assumptions", "Further notes", "Change history"}

// SpecRequiredSections must exist in every spec.
var SpecRequiredSections = []string{"Problem statement", "Solution", "User stories", "Implementation decisions", "Testing decisions", "Out of scope", "Change history"}

// SpecOptionalSections may be omitted when empty.
var SpecOptionalSections = []string{"Assumptions", "Further notes"}

// Story is one user story line.
type Story struct {
	N           int
	Text        string
	Withdrawn   bool
	WithdrawnIn int
	Line        int
}

// ID returns the story ID, such as "US 1".
func (s Story) ID() string { return fmt.Sprintf("US %d", s.N) }

// Spec is a parsed spec document.
type Spec struct {
	Doc          *Document
	S            *Structure
	Title        string
	Revision     int
	AttemptLimit int
	Stories      []Story
	History      []int
	Problems     []Problem
}

var (
	storyRe     = regexp.MustCompile(`^US (\d+): (.*)$`)
	withdrawnRe = regexp.MustCompile(`^Withdrawn in revision (\d+)\.$`)
	historyRe   = regexp.MustCompile(`^- Revision (\d+):`)
)

// ParseSpec parses spec bytes and records per-document problems.
func ParseSpec(b []byte) *Spec {
	sp := &Spec{Doc: Split(b)}
	sp.parse()
	return sp
}

// Reparse refreshes parsed fields after the document lines change.
func (sp *Spec) Reparse() { sp.parse() }

func (sp *Spec) problem(code, format string, args ...any) {
	sp.Problems = append(sp.Problems, Problem{Code: code, Ticket: "spec", Message: fmt.Sprintf(format, args...)})
}

func (sp *Spec) parse() {
	*sp = Spec{Doc: sp.Doc, AttemptLimit: DefaultAttemptLimit}
	s := ParseStructure(sp.Doc)
	sp.S = s
	sp.Title = s.Title
	for _, p := range s.Problems {
		sp.problem("malformed", "%s", p)
	}
	seen := map[string]bool{}
	for _, m := range s.Meta {
		if seen[m.Key] {
			sp.problem("duplicate_field", "`%s:` appears more than once", m.Key)
		}
		seen[m.Key] = true
		if !oneOf(m.Key, SpecMetaOrder...) {
			sp.problem("unknown_field", "`%s:` is not a spec field", m.Key)
		}
	}
	if v, ok := s.MetaValue("Revision"); !ok {
		sp.problem("missing_field", "`Revision:` is missing")
	} else if n, err := strconv.Atoi(v); err != nil || n < 1 {
		sp.problem("invalid_field", "`Revision:` %q is not a positive integer", v)
	} else {
		sp.Revision = n
	}
	if v, ok := s.MetaValue("Attempt limit"); ok {
		if n, err := strconv.Atoi(v); err != nil || n < 1 {
			sp.problem("invalid_field", "`Attempt limit:` %q is not a positive integer", v)
		} else {
			sp.AttemptLimit = n
		}
	}

	secSeen := map[string]bool{}
	for _, sec := range s.Sections {
		if secSeen[sec.Name] {
			sp.problem("duplicate_section", "`## %s` appears more than once", sec.Name)
		}
		secSeen[sec.Name] = true
	}
	for _, name := range SpecRequiredSections {
		if !secSeen[name] {
			sp.problem("missing_section", "`## %s` is missing", name)
		}
	}

	if sec, ok := s.Section("User stories"); ok {
		for i := sec.Heading + 1; i < sec.End; i++ {
			if s.Fenced[i] {
				continue
			}
			m := storyRe.FindStringSubmatch(sp.Doc.Lines[i])
			if m == nil {
				continue
			}
			n, _ := strconv.Atoi(m[1])
			st := Story{N: n, Text: strings.TrimSpace(m[2]), Line: i}
			if w := withdrawnRe.FindStringSubmatch(st.Text); w != nil {
				st.Withdrawn = true
				st.WithdrawnIn, _ = strconv.Atoi(w[1])
			}
			sp.Stories = append(sp.Stories, st)
		}
		if len(sp.Stories) == 0 {
			sp.problem("missing_stories", "`## User stories` has no `US <n>:` lines")
		}
		for i, st := range sp.Stories {
			if st.N != i+1 {
				sp.problem("story_ids", "story IDs must run from US 1 in order without gaps or repeats; found %s at position %d", st.ID(), i+1)
				break
			}
		}
	}

	if sec, ok := s.Section("Change history"); ok {
		for i := sec.Heading + 1; i < sec.End; i++ {
			if m := historyRe.FindStringSubmatch(sp.Doc.Lines[i]); m != nil && !s.Fenced[i] {
				n, _ := strconv.Atoi(m[1])
				sp.History = append(sp.History, n)
			}
		}
		if sp.Revision > 0 {
			found := false
			for _, n := range sp.History {
				found = found || n == sp.Revision
			}
			if !found {
				sp.problem("missing_history", "`## Change history` has no entry for revision %d", sp.Revision)
			}
		}
	}
}

// Story returns the story with ID "US n".
func (sp *Spec) Story(id string) (Story, bool) {
	for _, st := range sp.Stories {
		if st.ID() == id {
			return st, true
		}
	}
	return Story{}, false
}

// ActiveStories returns the IDs of stories that aren't withdrawn.
func (sp *Spec) ActiveStories() []string {
	var out []string
	for _, st := range sp.Stories {
		if !st.Withdrawn {
			out = append(out, st.ID())
		}
	}
	return out
}

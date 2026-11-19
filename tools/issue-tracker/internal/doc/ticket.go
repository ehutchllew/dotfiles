package doc

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Ticket kinds, phases, and statuses.
const (
	KindImplementation = "implementation"
	KindAcceptance     = "acceptance"
	KindManual         = "manual"

	PhaseImplement = "implement"
	PhaseReview    = "review"

	StatusOpen       = "open"
	StatusInProgress = "in-progress"
	StatusStopped    = "stopped"
	StatusDone       = "done"
	StatusCancelled  = "cancelled"
)

// TicketMetaOrder is the canonical order of ticket metadata lines.
var TicketMetaOrder = []string{"Kind", "Phase", "Status", "Attempt", "Blocked by", "Covers", "Spec revision", "Base commit", "Target commit", "Next criterion"}

// TicketSectionOrder is the canonical order of ticket sections.
var TicketSectionOrder = []string{"What to build", "Acceptance criteria", "Verification", "Out of scope", "Comments"}

var fieldKinds = map[string][]string{
	"Kind":           {KindImplementation, KindAcceptance, KindManual},
	"Phase":          {KindImplementation},
	"Status":         {KindImplementation, KindAcceptance, KindManual},
	"Attempt":        {KindImplementation},
	"Blocked by":     {KindImplementation, KindAcceptance, KindManual},
	"Covers":         {KindImplementation, KindAcceptance, KindManual},
	"Spec revision":  {KindAcceptance},
	"Base commit":    {KindImplementation, KindAcceptance},
	"Target commit":  {KindImplementation, KindAcceptance},
	"Next criterion": {KindImplementation, KindAcceptance, KindManual},
}

// Problem is one validation failure.
type Problem struct {
	Code    string `json:"code"`
	Ticket  string `json:"ticket,omitempty"`
	Message string `json:"message"`
}

// Criterion is one acceptance criterion checkbox.
type Criterion struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
	Line    int    `json:"-"`
}

// Item is one top-level `- ` list item in a comment.
type Item struct {
	Label string
	Text  string
	Line  int
}

// Comment is one entry in a ticket's Comments section.
type Comment struct {
	Op         string
	Attempt    int
	HasAttempt bool
	Header     int
	End        int
	Items      []Item
}

// Ticket is a parsed ticket document.
type Ticket struct {
	Doc           *Document
	S             *Structure
	Number        int
	Title         string
	Kind          string
	Phase         string
	Status        string
	Attempt       int
	BlockedBy     []int
	Covers        []string
	SpecRevision  int
	Base          string
	Target        string
	Criteria      []Criterion
	NextCriterion int
	Comments      []Comment
	Problems      []Problem
}

var (
	ticketTitleRe = regexp.MustCompile(`^(\d+): (\S.*)$`)
	storyIDRe     = regexp.MustCompile(`^US (\d+)$`)
	criterionRe   = regexp.MustCompile(`^- \[([ xX])\] (.*)$`)
	criterionIDRe = regexp.MustCompile(`^(C[0-9]+): (.*)$`)
	itemRe        = regexp.MustCompile(`^- ([^:]+):(?: (.*))?$`)
	commentHeadRe = regexp.MustCompile(`^(?:(?:implement|review|manual|acceptance):)?(claim|hand[- ]off|pass|send[- ]back|set[- ]feature[- ]base|start[- ]review|stop|resume|reset[- ]attempts|revise|complete|cancel|comment)(?: \(attempt (\d+)\))?:\s*$`)
	shaRe         = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)
)

// IsFullSHA reports whether s is a full SHA-1 or SHA-256 commit ID.
func IsFullSHA(s string) bool { return shaRe.MatchString(s) }

// TicketID formats a ticket number.
func TicketID(n int) string { return fmt.Sprintf("%02d", n) }

// CriterionID formats a criterion number within one ticket.
func CriterionID(n int) string { return fmt.Sprintf("C%02d", n) }

// ParseTicketNumber parses "1", "01", or "001".
func ParseTicketNumber(s string) (int, error) {
	s = strings.TrimSpace(s)
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 || strings.HasPrefix(s, "+") {
		return 0, fmt.Errorf("%q is not a ticket number", s)
	}
	return n, nil
}

// ParseTicket parses ticket bytes and records per-document problems.
func ParseTicket(b []byte) *Ticket {
	d := Split(b)
	t := &Ticket{Doc: d}
	t.parse()
	return t
}

// Reparse refreshes parsed fields after the document lines change.
func (t *Ticket) Reparse() { t.parse() }

func (t *Ticket) problem(code, format string, args ...any) {
	t.Problems = append(t.Problems, Problem{Code: code, Message: fmt.Sprintf(format, args...)})
}

func (t *Ticket) parse() {
	*t = Ticket{Doc: t.Doc}
	s := ParseStructure(t.Doc)
	t.S = s
	for _, p := range s.Problems {
		t.problem("malformed", "%s", p)
	}
	if s.TitleLine >= 0 {
		m := ticketTitleRe.FindStringSubmatch(s.Title)
		if m == nil {
			t.problem("malformed", "the title must look like `# <NN>: <title>`")
		} else {
			t.Number, _ = strconv.Atoi(m[1])
			t.Title = strings.TrimSpace(m[2])
		}
	}

	seen := map[string]bool{}
	for _, m := range s.Meta {
		if seen[m.Key] {
			t.problem("duplicate_field", "`%s:` appears more than once", m.Key)
		}
		seen[m.Key] = true
		if _, ok := fieldKinds[m.Key]; !ok {
			t.problem("unknown_field", "`%s:` is not a ticket field", m.Key)
		}
	}
	get := func(k string) (string, bool) { return s.MetaValue(k) }

	t.Kind, _ = get("Kind")
	validKind := oneOf(t.Kind, KindImplementation, KindAcceptance, KindManual)
	if !validKind {
		t.problem("invalid_field", "`Kind:` must be implementation, acceptance, or manual")
	}
	for _, m := range s.Meta {
		if kinds, ok := fieldKinds[m.Key]; ok && validKind && !oneOf(t.Kind, kinds...) {
			t.problem("field_not_allowed", "`%s:` doesn't apply to %s tickets", m.Key, t.Kind)
		}
	}

	var ok bool
	t.Status, ok = get("Status")
	if !ok {
		t.problem("missing_field", "`Status:` is missing")
	} else if !oneOf(t.Status, StatusOpen, StatusInProgress, StatusStopped, StatusDone, StatusCancelled) {
		t.problem("invalid_field", "`Status:` %q is not allowed", t.Status)
	}

	if t.Kind == KindImplementation {
		t.Phase, ok = get("Phase")
		if !ok {
			t.problem("missing_field", "`Phase:` is missing")
		} else if !oneOf(t.Phase, PhaseImplement, PhaseReview) {
			t.problem("invalid_field", "`Phase:` %q is not allowed", t.Phase)
		}
		v, ok := get("Attempt")
		if !ok {
			t.problem("missing_field", "`Attempt:` is missing")
		} else if n, err := strconv.Atoi(v); err != nil || n < 0 {
			t.problem("invalid_field", "`Attempt:` %q is not a non-negative integer", v)
		} else {
			t.Attempt = n
		}
	}

	if v, ok := get("Blocked by"); ok {
		t.BlockedBy = nil
		for _, part := range splitList(v) {
			n, err := ParseTicketNumber(part)
			if err != nil {
				t.problem("invalid_field", "`Blocked by:` has %q, which is not a ticket number", part)
				continue
			}
			t.BlockedBy = append(t.BlockedBy, n)
		}
	}

	if v, ok := get("Covers"); !ok {
		t.problem("missing_field", "`Covers:` is missing")
	} else if covers, err := ParseCovers(v); err != nil {
		t.problem("invalid_field", "`Covers:` %v", err)
	} else {
		t.Covers = covers
	}

	if t.Kind == KindAcceptance {
		v, ok := get("Spec revision")
		if !ok {
			t.problem("missing_field", "`Spec revision:` is missing")
		} else if n, err := strconv.Atoi(v); err != nil || n < 1 {
			t.problem("invalid_field", "`Spec revision:` %q is not a positive integer", v)
		} else {
			t.SpecRevision = n
		}
	}

	for _, f := range []struct {
		key string
		dst *string
	}{{"Base commit", &t.Base}, {"Target commit", &t.Target}} {
		if v, ok := get(f.key); ok {
			if !IsFullSHA(v) {
				t.problem("invalid_field", "`%s:` %q is not a full commit SHA", f.key, v)
			}
			*f.dst = v
		}
	}

	if t.Kind == KindImplementation {
		if t.Phase == PhaseReview && !oneOf(t.Status, StatusOpen, StatusStopped, StatusDone, StatusCancelled) {
			t.problem("invalid_state", "an implementation ticket in the review phase must be open, stopped, done, or cancelled, not %s", t.Status)
		}
		if t.Attempt > 0 && t.Base == "" {
			t.problem("invalid_state", "`Attempt:` is above 0 but `Base commit:` is missing")
		}
		if t.Phase == PhaseReview && t.Target == "" {
			t.problem("invalid_state", "the ticket is in the review phase but `Target commit:` is missing")
		}
	}

	secSeen := map[string]bool{}
	for _, sec := range s.Sections {
		if secSeen[sec.Name] {
			t.problem("duplicate_section", "`## %s` appears more than once", sec.Name)
		}
		secSeen[sec.Name] = true
	}
	for _, name := range []string{"What to build", "Acceptance criteria", "Verification", "Comments"} {
		if !secSeen[name] {
			t.problem("missing_section", "`## %s` is missing", name)
		}
	}

	if sec, ok := s.Section("Acceptance criteria"); ok {
		// fromPosition records whether an ID was written or came from the
		// criterion's place in the list.
		fromPosition := map[string]bool{}
		var clashes []string
		next := 1
		for i := sec.Heading + 1; i < sec.End; i++ {
			if s.Fenced[i] {
				continue
			}
			if m := criterionRe.FindStringSubmatch(t.Doc.Lines[i]); m != nil {
				id, text, positional := CriterionID(len(t.Criteria)+1), strings.TrimSpace(m[2]), true
				if tagged := criterionIDRe.FindStringSubmatch(text); tagged != nil {
					id, text, positional = tagged[1], strings.TrimSpace(tagged[2]), false
				}
				n, err := strconv.Atoi(id[1:])
				if err != nil || n < 1 || n == math.MaxInt || id != CriterionID(n) {
					t.problem("invalid_criterion", "criterion ID %q is invalid", id)
				} else {
					next = max(next, n+1)
				}
				if earlier, ok := fromPosition[id]; ok {
					if earlier || positional {
						clashes = append(clashes, id)
					} else {
						t.problem("duplicate_criterion", "criterion ID %s appears more than once", id)
					}
				}
				fromPosition[id] = fromPosition[id] || positional
				t.Criteria = append(t.Criteria, Criterion{ID: id, Text: text, Checked: m[1] != " ", Line: i})
			}
		}
		if len(t.Criteria) == 0 {
			t.problem("missing_criteria", "`## Acceptance criteria` has no `- [ ]` criteria")
		}
		t.NextCriterion = next
		if v, ok := get("Next criterion"); ok {
			if n, err := strconv.Atoi(v); err != nil || n < next {
				t.problem("invalid_field", "`Next criterion:` must be at least %d", next)
			} else {
				t.NextCriterion = n
			}
		}
		for _, id := range clashes {
			t.problem("duplicate_criterion", "two criteria use ID %s. One has no ID written on it, so it got that ID from its place in the list. Write an unused ID, such as %s, on each criterion that has none.", id, CriterionID(t.NextCriterion))
		}
	}

	t.parseComments()
}

func (t *Ticket) parseComments() {
	sec, ok := t.S.Section("Comments")
	if !ok {
		return
	}
	lines := t.Doc.Lines
	var cur *Comment
	for i := sec.Heading + 1; i < sec.End; i++ {
		line := lines[i]
		if !t.S.Fenced[i] {
			if m := commentHeadRe.FindStringSubmatch(line); m != nil {
				if cur != nil {
					cur.End = i
					t.Comments = append(t.Comments, *cur)
				}
				cur = &Comment{Op: strings.ReplaceAll(m[1], " ", "-"), Header: i}
				if m[2] != "" {
					cur.Attempt, _ = strconv.Atoi(m[2])
					cur.HasAttempt = true
				}
				continue
			}
		}
		if cur == nil || t.S.Fenced[i] {
			continue
		}
		if m := itemRe.FindStringSubmatch(line); m != nil {
			cur.Items = append(cur.Items, Item{Label: strings.TrimSpace(m[1]), Text: strings.TrimSpace(m[2]), Line: i})
		}
	}
	if cur != nil {
		cur.End = sec.End
		t.Comments = append(t.Comments, *cur)
	}
}

// ParseCovers parses a `Covers:` value. "none" returns an empty list.
func ParseCovers(v string) ([]string, error) {
	v = strings.TrimSpace(v)
	if v == "none" {
		return []string{}, nil
	}
	if v == "" {
		return nil, fmt.Errorf("is empty; use `none` when the ticket covers no story")
	}
	var out []string
	for _, part := range splitList(v) {
		m := storyIDRe.FindStringSubmatch(part)
		n, _ := strconv.Atoi(safeIndex(m, 1))
		if m == nil || n < 1 {
			return nil, fmt.Errorf("has %q, which is not a story ID such as `US 1`", part)
		}
		out = append(out, fmt.Sprintf("US %d", n))
	}
	return out, nil
}

// FormatCovers renders a `Covers:` value.
func FormatCovers(ids []string) string {
	if len(ids) == 0 {
		return "none"
	}
	return strings.Join(ids, ", ")
}

// FormatBlockedBy renders a `Blocked by:` value.
func FormatBlockedBy(ns []int) string {
	parts := make([]string, len(ns))
	for i, n := range ns {
		parts[i] = TicketID(n)
	}
	return strings.Join(parts, ", ")
}

// ID returns the ticket's two-digit number.
func (t *Ticket) ID() string { return TicketID(t.Number) }

// FindingKey normalizes a not-blocking finding for matching.
func FindingKey(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSuffix(s, ".")
}

var movedRe = regexp.MustCompile(`^(.*?)\s*Moved to ticket (\d+)\.$`)

// ParseResolution splits a `Not blocking resolved:` value into the finding
// and how it was resolved.
func ParseResolution(text string) (finding, how string, ok bool) {
	text = strings.TrimSpace(text)
	if m := movedRe.FindStringSubmatch(text); m != nil {
		return strings.TrimSpace(m[1]), "moved", true
	}
	for _, suffix := range []string{"Fixed.", "Ignored."} {
		if before, found := strings.CutSuffix(text, suffix); found {
			return strings.TrimSpace(before), strings.ToLower(strings.TrimSuffix(suffix, ".")), true
		}
	}
	return "", "", false
}

// UnresolvedNotBlocking returns the not-blocking findings with no later
// resolution, oldest first.
func (t *Ticket) UnresolvedNotBlocking() []string {
	var order []string
	open := map[string]string{}
	for _, c := range t.Comments {
		for _, it := range c.Items {
			switch it.Label {
			case "Not blocking":
				if it.Text == "" {
					continue
				}
				k := FindingKey(it.Text)
				if _, ok := open[k]; !ok {
					open[k] = it.Text
					order = append(order, k)
				}
			case "Not blocking resolved":
				if f, _, ok := ParseResolution(it.Text); ok {
					delete(open, FindingKey(f))
				}
			}
		}
	}
	var out []string
	for _, k := range order {
		if v, ok := open[k]; ok {
			out = append(out, v)
			delete(open, k)
		}
	}
	return out
}

// LastReviseComment returns the index of the last `revise` comment, or -1.
func (t *Ticket) LastReviseComment() int {
	for i, c := range slices.Backward(t.Comments) {
		if c.Op == "revise" {
			return i
		}
	}
	return -1
}

// CommentText returns the lines of comment i.
func (t *Ticket) CommentText(i int) string {
	c := t.Comments[i]
	lines := t.Doc.Lines[c.Header:c.End]
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// CurrentHistory returns the document with comments before the last
// `revise` removed, and how many comments it removed.
func (t *Ticket) CurrentHistory() ([]byte, int) {
	i := t.LastReviseComment()
	if i <= 0 {
		return t.Doc.Bytes(), 0
	}
	sec, _ := t.S.Section("Comments")
	d := t.Doc.Clone()
	d.Lines = replaceRange(d.Lines, sec.Heading+1, t.Comments[i].Header, []string{""})
	return d.Bytes(), i
}

// IsFinished reports whether the ticket is done or cancelled.
func (t *Ticket) IsFinished() bool {
	return t.Status == StatusDone || t.Status == StatusCancelled
}

func safeIndex(m []string, i int) string {
	if i < len(m) {
		return m[i]
	}
	return ""
}

func splitList(v string) []string {
	var out []string
	for part := range strings.SplitSeq(v, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func oneOf(v string, options ...string) bool {
	return slices.Contains(options, v)
}

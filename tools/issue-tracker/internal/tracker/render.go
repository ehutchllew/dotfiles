package tracker

import (
	"strings"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/apperr"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/doc"
)

// commentItem is one top-level item of a rendered comment.
type commentItem struct {
	label  string
	values []string
	list   bool
	raw    bool
}

// renderComment builds comment lines. A single one-line value goes on the
// item's line. A multi-line value goes on the lines below its label, so a
// fence in it starts its own line. Several values, or a list item, become a
// nested list. Continuation lines are indented so they can't start a new
// item or comment.
func renderComment(header string, items []commentItem) []string {
	lines := []string{header + ":"}
	for _, it := range items {
		switch {
		case it.raw:
			for _, v := range it.values {
				lines = append(lines, indentRest("- "+v, "  ")...)
			}
		case len(it.values) == 1 && !it.list && !strings.Contains(it.values[0], "\n"):
			lines = append(lines, "- "+it.label+": "+it.values[0])
		case len(it.values) == 1 && !it.list:
			lines = append(lines, "- "+it.label+":")
			lines = append(lines, indentRest("\n"+it.values[0], "  ")[1:]...)
		default:
			lines = append(lines, "- "+it.label+":")
			for _, v := range it.values {
				lines = append(lines, indentRest("  - "+v, "    ")...)
			}
		}
	}
	return lines
}

func indentRest(text, indent string) []string {
	parts := strings.Split(text, "\n")
	for i := 1; i < len(parts); i++ {
		if strings.TrimSpace(parts[i]) != "" {
			parts[i] = indent + parts[i]
		} else {
			parts[i] = ""
		}
	}
	return parts
}

// checkValues trims values and rejects empty ones, unclosed fences, and
// line breaks where a single line is required.
func checkValues(field string, values []string, singleLine bool) ([]string, error) {
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(strings.ReplaceAll(v, "\r\n", "\n"))
		if v == "" {
			return nil, apperr.Usage("invalid_input", "%s has an empty value", field)
		}
		if singleLine && strings.Contains(v, "\n") {
			return nil, apperr.Usage("invalid_input", "each %s value must be one line", field)
		}
		if strings.ContainsRune(v, '\r') {
			return nil, apperr.Usage("invalid_input", "%s contains a carriage return", field)
		}
		if _, unclosed := doc.FenceMask(strings.Split(v, "\n")); unclosed {
			return nil, apperr.Usage("invalid_input", "%s has a fenced code block that is never closed", field)
		}
		out = append(out, v)
	}
	return out, nil
}

// findCriterion matches an ID, or exact text for older callers.
func findCriterion(t *doc.Ticket, text string) (int, error) {
	s := strings.TrimSpace(text)
	for i, c := range t.Criteria {
		if c.ID == s {
			return i, nil
		}
	}
	try := func(s string) []int {
		var hits []int
		for i, c := range t.Criteria {
			if c.Text == s {
				hits = append(hits, i)
			}
		}
		return hits
	}
	hits := try(s)
	if len(hits) == 0 && len(s) >= 2 && strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		hits = try(strings.TrimSpace(s[1 : len(s)-1]))
	}
	switch len(hits) {
	case 1:
		return hits[0], nil
	case 0:
		return 0, apperr.Usage("unknown_criterion", "ticket %s has no criterion %q; use a criterion ID from ticket get", t.ID(), text).With("criteria", t.Criteria)
	default:
		return 0, apperr.New("ambiguous_criterion", "ticket %s has %d criteria with the text %q", t.ID(), len(hits), text)
	}
}

// setCriterion ticks or unticks one criterion line in place.
func setCriterion(t *doc.Ticket, i int, checked bool) {
	line := t.Doc.Lines[t.Criteria[i].Line]
	mark := " "
	if checked {
		mark = "x"
	}
	t.Doc.Lines[t.Criteria[i].Line] = line[:3] + mark + line[4:]
	t.Criteria[i].Checked = checked
}

func setAllCriteria(t *doc.Ticket, checked bool) {
	for i := range t.Criteria {
		setCriterion(t, i, checked)
	}
}

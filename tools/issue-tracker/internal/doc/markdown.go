// Package doc parses and edits spec and ticket Markdown while preserving
// every line it does not change.
package doc

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// Document holds a file as lines without terminators, so edits can replace
// individual lines and leave the rest byte-for-byte intact.
type Document struct {
	Lines        []string
	EOL          string
	FinalNewline bool
}

// Split turns file bytes into a Document. It keeps the file's line ending
// style for lines it later adds.
func Split(b []byte) *Document {
	s := string(b)
	d := &Document{EOL: "\n"}
	if strings.Contains(s, "\r\n") {
		d.EOL = "\r\n"
	}
	if s == "" {
		return d
	}
	d.FinalNewline = strings.HasSuffix(s, "\n")
	s = strings.TrimSuffix(s, "\n")
	for line := range strings.SplitSeq(s, "\n") {
		d.Lines = append(d.Lines, strings.TrimSuffix(line, "\r"))
	}
	return d
}

// Bytes renders the document.
func (d *Document) Bytes() []byte {
	var b strings.Builder
	for i, line := range d.Lines {
		b.WriteString(line)
		if i < len(d.Lines)-1 || d.FinalNewline {
			b.WriteString(d.EOL)
		}
	}
	return []byte(b.String())
}

// Clone returns an independent copy.
func (d *Document) Clone() *Document {
	c := *d
	c.Lines = append([]string(nil), d.Lines...)
	return &c
}

// Hash returns the opaque content hash for document bytes: the first 64 bits
// of the SHA-256 digest, which is enough to detect changes and keeps
// responses short.
func Hash(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:8])
}

// FenceMask reports, for each line, whether it belongs to a fenced code
// block, including the fence lines themselves.
func FenceMask(lines []string) (mask []bool, unclosed bool) {
	mask = make([]bool, len(lines))
	var open bool
	var ch byte
	var n int
	for i, line := range lines {
		c, count, rest, ok := fenceMarker(line)
		switch {
		case !open && ok:
			if c == '`' && strings.Contains(rest, "`") {
				continue
			}
			open, ch, n = true, c, count
			mask[i] = true
		case open:
			mask[i] = true
			if ok && c == ch && count >= n && strings.TrimSpace(rest) == "" {
				open = false
			}
		}
	}
	return mask, open
}

func fenceMarker(line string) (byte, int, string, bool) {
	t := stripListMarker(strings.TrimLeft(line, " \t"))
	if len(t) < 3 || (t[0] != '`' && t[0] != '~') {
		return 0, 0, "", false
	}
	c := t[0]
	i := 0
	for i < len(t) && t[i] == c {
		i++
	}
	if i < 3 {
		return 0, 0, "", false
	}
	return c, i, t[i:], true
}

// stripListMarker removes one leading list marker, such as "- " or "1. ",
// because a fence can open a list item.
func stripListMarker(t string) string {
	if len(t) >= 2 && (t[0] == '-' || t[0] == '*' || t[0] == '+') && t[1] == ' ' {
		return strings.TrimLeft(t[2:], " ")
	}
	i := 0
	for i < len(t) && i < 9 && t[i] >= '0' && t[i] <= '9' {
		i++
	}
	if i > 0 && i+1 < len(t) && (t[i] == '.' || t[i] == ')') && t[i+1] == ' ' {
		return strings.TrimLeft(t[i+2:], " ")
	}
	return t
}

// Meta is one metadata line under the title.
type Meta struct {
	Key   string
	Value string
	Line  int
}

// Section is a level-2 heading and the lines up to the next one.
// Its body is Lines[Heading+1 : End].
type Section struct {
	Name    string
	Heading int
	End     int
}

// Structure is the parsed shape shared by specs and tickets.
type Structure struct {
	TitleLine int
	Title     string
	Meta      []Meta
	Sections  []Section
	Fenced    []bool
	Problems  []string
}

var metaRe = regexp.MustCompile(`^([A-Z][a-z]+(?: [a-z]+)*):(?: (.*))?$`)

// ParseStructure finds the title, metadata block, and level-2 sections.
// Headings and metadata inside fenced code blocks are ignored.
func ParseStructure(d *Document) *Structure {
	s := &Structure{TitleLine: -1}
	mask, unclosed := FenceMask(d.Lines)
	s.Fenced = mask
	if unclosed {
		s.Problems = append(s.Problems, "a fenced code block is never closed")
	}
	i := 0
	for i < len(d.Lines) && strings.TrimSpace(d.Lines[i]) == "" {
		i++
	}
	if i >= len(d.Lines) || mask[i] || !strings.HasPrefix(d.Lines[i], "# ") {
		s.Problems = append(s.Problems, "the first line must be a `# ` title")
	} else {
		s.TitleLine = i
		s.Title = strings.TrimSpace(strings.TrimPrefix(d.Lines[i], "# "))
		i++
		for i < len(d.Lines) && strings.TrimSpace(d.Lines[i]) == "" {
			i++
		}
		for i < len(d.Lines) && !mask[i] {
			m := metaRe.FindStringSubmatch(d.Lines[i])
			if m == nil {
				break
			}
			s.Meta = append(s.Meta, Meta{Key: m[1], Value: strings.TrimSpace(m[2]), Line: i})
			i++
		}
	}
	for j, line := range d.Lines {
		if mask[j] {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			if n := len(s.Sections); n > 0 {
				s.Sections[n-1].End = j
			}
			s.Sections = append(s.Sections, Section{Name: strings.TrimSpace(line[3:]), Heading: j, End: len(d.Lines)})
		} else if strings.HasPrefix(line, "# ") && j != s.TitleLine {
			s.Problems = append(s.Problems, fmt.Sprintf("line %d has a second `# ` title", j+1))
		}
	}
	return s
}

// MetaValue returns the value of the first metadata line with key.
func (s *Structure) MetaValue(key string) (string, bool) {
	for _, m := range s.Meta {
		if m.Key == key {
			return m.Value, true
		}
	}
	return "", false
}

// Section returns the first section with the given name.
func (s *Structure) Section(name string) (Section, bool) {
	for _, sec := range s.Sections {
		if sec.Name == name {
			return sec, true
		}
	}
	return Section{}, false
}

// SetMeta replaces the value of key, or inserts the line in canonical order.
func SetMeta(d *Document, key, value string, order []string) error {
	s := ParseStructure(d)
	if s.TitleLine < 0 {
		return fmt.Errorf("document has no title")
	}
	line := key + ": " + value
	for _, m := range s.Meta {
		if m.Key == key {
			d.Lines[m.Line] = line
			return nil
		}
	}
	rank := indexOf(order, key)
	insertAt := -1
	for _, m := range s.Meta {
		if r := indexOf(order, m.Key); r >= 0 && r < rank {
			insertAt = m.Line + 1
		}
	}
	if insertAt < 0 {
		if len(s.Meta) > 0 {
			insertAt = s.Meta[0].Line
		} else {
			d.Lines = insertLines(d.Lines, s.TitleLine+1, "", line)
			return nil
		}
	}
	d.Lines = insertLines(d.Lines, insertAt, line)
	return nil
}

// RemoveMeta deletes every metadata line with key.
func RemoveMeta(d *Document, key string) {
	s := ParseStructure(d)
	for _, m := range slices.Backward(s.Meta) {
		if m.Key == key {
			d.Lines = append(d.Lines[:m.Line], d.Lines[m.Line+1:]...)
		}
	}
}

// SetTitle replaces the title line.
func SetTitle(d *Document, title string) error {
	s := ParseStructure(d)
	if s.TitleLine < 0 {
		return fmt.Errorf("document has no title")
	}
	d.Lines[s.TitleLine] = "# " + title
	return nil
}

// BodyLines splits authored text into lines, dropping surrounding blank lines.
func BodyLines(body string) []string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	lines := strings.Split(body, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	return lines
}

// CheckBody rejects authored section text that would change the document's
// structure: unclosed fences or top-level headings.
func CheckBody(lines []string) error {
	mask, unclosed := FenceMask(lines)
	if unclosed {
		return fmt.Errorf("a fenced code block is never closed")
	}
	for i, line := range lines {
		if !mask[i] && (strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ")) {
			return fmt.Errorf("line %q is a `#` or `##` heading; use `###` or deeper inside a section", line)
		}
	}
	return nil
}

// ReplaceSection replaces a section's body, inserting the section in
// canonical order when it is missing.
func ReplaceSection(d *Document, name string, body []string, order []string) {
	s := ParseStructure(d)
	sec, ok := s.Section(name)
	if !ok {
		insertSection(d, s, name, body, order)
		return
	}
	region := []string{""}
	region = append(region, body...)
	if sec.End < len(d.Lines) {
		region = append(region, "")
	}
	if sec.End >= len(d.Lines) {
		if len(body) == 0 {
			region = nil
		}
		d.FinalNewline = true
	}
	d.Lines = replaceRange(d.Lines, sec.Heading+1, sec.End, region)
}

// RemoveSection deletes a section and its body.
func RemoveSection(d *Document, name string) {
	s := ParseStructure(d)
	sec, ok := s.Section(name)
	if !ok {
		return
	}
	start := sec.Heading
	// Drop the blank line that separated it from the previous section when it
	// was the last section, so the file doesn't end with stray blank lines.
	if sec.End >= len(d.Lines) {
		for start > 0 && strings.TrimSpace(d.Lines[start-1]) == "" {
			start--
		}
	}
	d.Lines = replaceRange(d.Lines, start, sec.End, nil)
}

func insertSection(d *Document, s *Structure, name string, body []string, order []string) {
	rank := indexOf(order, name)
	at := len(d.Lines)
	for _, sec := range s.Sections {
		if r := indexOf(order, sec.Name); r > rank {
			at = sec.Heading
			break
		}
	}
	block := []string{"## " + name, ""}
	block = append(block, body...)
	if at < len(d.Lines) {
		block = append(block, "")
		d.Lines = insertLines(d.Lines, at, block...)
		return
	}
	for len(d.Lines) > 0 && strings.TrimSpace(d.Lines[len(d.Lines)-1]) == "" {
		d.Lines = d.Lines[:len(d.Lines)-1]
	}
	if len(body) == 0 {
		block = block[:1]
	}
	d.Lines = append(d.Lines, "")
	d.Lines = append(d.Lines, block...)
	d.FinalNewline = true
}

// AppendToSection adds lines at the end of a section, after its last
// non-blank line and separated from it by one blank line.
func AppendToSection(d *Document, name string, lines []string) error {
	s := ParseStructure(d)
	sec, ok := s.Section(name)
	if !ok {
		return fmt.Errorf("document has no %q section", name)
	}
	last := sec.Heading
	for i := sec.End - 1; i > sec.Heading; i-- {
		if strings.TrimSpace(d.Lines[i]) != "" {
			last = i
			break
		}
	}
	block := append([]string{""}, lines...)
	if sec.End < len(d.Lines) {
		block = append(block, "")
		d.Lines = replaceRange(d.Lines, last+1, sec.End, block)
		return nil
	}
	d.Lines = append(d.Lines[:last+1], block...)
	d.FinalNewline = true
	return nil
}

// AppendListItem adds one line at the end of a section. It follows a list
// directly and leaves a blank line after any other content.
func AppendListItem(d *Document, name, line string) error {
	s := ParseStructure(d)
	sec, ok := s.Section(name)
	if !ok {
		return fmt.Errorf("document has no %q section", name)
	}
	last := sec.Heading
	for i := sec.End - 1; i > sec.Heading; i-- {
		if strings.TrimSpace(d.Lines[i]) != "" {
			last = i
			break
		}
	}
	if last == sec.Heading || !strings.HasPrefix(d.Lines[last], "- ") {
		return AppendToSection(d, name, []string{line})
	}
	d.Lines = insertLines(d.Lines, last+1, line)
	if last+1 == len(d.Lines)-1 {
		d.FinalNewline = true
	}
	return nil
}

func indexOf(list []string, v string) int {
	for i, x := range list {
		if x == v {
			return i
		}
	}
	return len(list)
}

func insertLines(lines []string, at int, add ...string) []string {
	out := make([]string, 0, len(lines)+len(add))
	out = append(out, lines[:at]...)
	out = append(out, add...)
	return append(out, lines[at:]...)
}

func replaceRange(lines []string, start, end int, with []string) []string {
	out := make([]string, 0, len(lines)-(end-start)+len(with))
	out = append(out, lines[:start]...)
	out = append(out, with...)
	return append(out, lines[end:]...)
}

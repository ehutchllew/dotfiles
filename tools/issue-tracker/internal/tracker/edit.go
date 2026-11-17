package tracker

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/apperr"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/doc"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/store"
)

func checkTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" || strings.ContainsAny(title, "\r\n") {
		return "", apperr.Usage("invalid_input", "`title` must be one non-empty line")
	}
	if strings.HasPrefix(title, "#") {
		return "", apperr.Usage("invalid_input", "`title` must not start with `#`")
	}
	return title, nil
}

// sectionBody checks authored section text and returns its lines.
func sectionBody(field, text string, required bool) ([]string, error) {
	lines := doc.BodyLines(text)
	if len(lines) == 0 {
		if required {
			return nil, apperr.Usage("invalid_input", "`%s` is required and can't be empty", field)
		}
		return nil, nil
	}
	if err := doc.CheckBody(lines); err != nil {
		return nil, apperr.Usage("invalid_input", "`%s`: %v", field, err)
	}
	return lines, nil
}

func checkStoryText(text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" || strings.ContainsAny(text, "\r\n") {
		return "", apperr.Usage("invalid_input", "each user story must be one non-empty line")
	}
	if strings.HasPrefix(text, "US ") {
		return "", apperr.Usage("invalid_input", "give story text without the `US <n>:` prefix; the CLI numbers stories")
	}
	if strings.HasPrefix(text, "Withdrawn in revision") {
		return "", apperr.Usage("invalid_input", "use `withdraw_stories` to withdraw a story")
	}
	return text, nil
}

func (tr *Tracker) writeResult(r *WriteResult, path string, data []byte, create bool) (*WriteResult, error) {
	if err := tr.Store.Write(path, data, create); err != nil {
		return nil, err
	}
	r.Hash = doc.Hash(data)
	r.Changed = []string{path}
	return tr.finish(r), nil
}

// SpecCreate writes a new spec at revision 1.
func (tr *Tracker) SpecCreate(id string, in *SpecCreateInput) (*WriteResult, error) {
	title, err := checkTitle(in.Title)
	if err != nil {
		return nil, err
	}
	limit := in.AttemptLimit
	if limit == 0 {
		limit = doc.DefaultAttemptLimit
	}
	if limit < 0 {
		return nil, apperr.Usage("invalid_input", "`attempt_limit` must be a positive integer")
	}
	if len(in.UserStories) == 0 {
		return nil, apperr.Usage("invalid_input", "`user_stories` needs at least one story")
	}
	history := strings.TrimSpace(in.History)
	if history == "" {
		history = "Initial spec."
	}
	if strings.ContainsAny(history, "\r\n") {
		return nil, apperr.Usage("invalid_input", "`history` must be one line")
	}

	lines := []string{"# " + title, "", "Revision: 1", "Attempt limit: " + strconv.Itoa(limit)}
	add := func(name string, body []string) {
		lines = append(lines, "", "## "+name, "")
		lines = append(lines, body...)
	}
	for _, s := range []struct {
		name, field, text string
		required          bool
	}{
		{"Problem statement", "problem_statement", in.ProblemStatement, true},
		{"Solution", "solution", in.Solution, true},
	} {
		body, err := sectionBody(s.field, s.text, s.required)
		if err != nil {
			return nil, err
		}
		add(s.name, body)
	}
	var stories []string
	for i, text := range in.UserStories {
		text, err := checkStoryText(text)
		if err != nil {
			return nil, err
		}
		if i > 0 {
			stories = append(stories, "")
		}
		stories = append(stories, fmt.Sprintf("US %d: %s", i+1, text))
	}
	add("User stories", stories)
	for _, s := range []struct {
		name, field, text string
		required          bool
	}{
		{"Implementation decisions", "implementation_decisions", in.ImplementationDecisions, true},
		{"Testing decisions", "testing_decisions", in.TestingDecisions, true},
		{"Out of scope", "out_of_scope", in.OutOfScope, true},
		{"Assumptions", "assumptions", in.Assumptions, false},
		{"Further notes", "further_notes", in.FurtherNotes, false},
	} {
		body, err := sectionBody(s.field, s.text, s.required)
		if err != nil {
			return nil, err
		}
		if body != nil {
			add(s.name, body)
		}
	}
	add("Change history", []string{"- Revision 1: " + history})
	data := []byte(strings.Join(lines, "\n") + "\n")
	if sp := doc.ParseSpec(data); len(sp.Problems) > 0 {
		return nil, apperr.New("render_failed", "the new spec would be malformed").With("problems", sp.Problems)
	}

	unlock, err := tr.lock(id, true)
	if err != nil {
		return nil, err
	}
	defer unlock()
	path := store.SpecPath(id)
	if ok, err := tr.Store.Exists(path); err != nil {
		return nil, err
	} else if ok {
		return nil, apperr.New("feature_exists", "feature %q already has a spec; use `spec edit`", id)
	}
	r := &WriteResult{Feature: id, Operation: "spec create", Revision: 1, CommitSubject: "spec: create " + id}
	return tr.writeResult(r, path, data, true)
}

// SpecEdit changes authored spec content. The caller declares whether the
// change is editorial or a user-approved revision.
func (tr *Tracker) SpecEdit(id, ifMatch string, in *SpecEditInput) (*WriteResult, error) {
	if err := requireIfMatch(ifMatch); err != nil {
		return nil, err
	}
	content := in.Title != nil || in.ProblemStatement != nil || in.Solution != nil || in.ImplementationDecisions != nil ||
		in.TestingDecisions != nil || in.OutOfScope != nil || in.Assumptions != nil || in.FurtherNotes != nil ||
		len(in.AddStories) > 0 || len(in.EditStories) > 0 || len(in.WithdrawStories) > 0
	if !content && in.AttemptLimit == nil {
		return nil, apperr.Usage("invalid_input", "the input changes nothing")
	}
	switch {
	case !content && in.Change != "":
		return nil, apperr.Usage("invalid_input", "`change` applies only to content edits; an attempt limit change doesn't change the revision")
	case content && in.Change != "editorial" && in.Change != "revision":
		return nil, apperr.Usage("invalid_input", "`change` must be `editorial` or `revision` for content edits")
	case in.Change == "revision" && !in.Approved:
		return nil, apperr.Usage("invalid_input", "a revision needs `approved: true`, confirming the user approved the change")
	case in.Change == "revision" && (strings.TrimSpace(in.History) == "" || strings.ContainsAny(in.History, "\r\n")):
		return nil, apperr.Usage("invalid_input", "a revision needs a one-line `history` naming the affected story IDs and what changed")
	case in.Change != "revision" && (in.History != "" || in.Approved):
		return nil, apperr.Usage("invalid_input", "`history` and `approved` apply only to `change: revision`")
	case in.Change != "revision" && (len(in.AddStories) > 0 || len(in.WithdrawStories) > 0):
		return nil, apperr.Usage("invalid_input", "adding or withdrawing stories needs `change: revision`")
	}
	if in.AttemptLimit != nil && *in.AttemptLimit < 1 {
		return nil, apperr.Usage("invalid_input", "`attempt_limit` must be a positive integer")
	}

	unlock, err := tr.lock(id, true)
	if err != nil {
		return nil, err
	}
	defer unlock()
	path := store.SpecPath(id)
	raw, err := tr.Store.Read(path)
	if e, ok := err.(*apperr.Error); ok && e.Code == "not_found" {
		return nil, apperr.New("feature_not_found", "feature %q has no spec at %s", id, path)
	}
	if err != nil {
		return nil, err
	}
	if err := checkCurrent(path, doc.Hash(raw), ifMatch); err != nil {
		return nil, err
	}
	sp := doc.ParseSpec(raw)
	if len(sp.Problems) > 0 {
		return nil, apperr.New("invalid_spec", "the spec has problems; fix them first").With("problems", sp.Problems)
	}
	rev := sp.Revision
	if in.Change == "revision" {
		rev++
	}
	d := sp.Doc

	if in.Title != nil {
		title, err := checkTitle(*in.Title)
		if err != nil {
			return nil, err
		}
		_ = doc.SetTitle(d, title)
	}
	for _, s := range []struct {
		name, field string
		text        *string
		required    bool
	}{
		{"Problem statement", "problem_statement", in.ProblemStatement, true},
		{"Solution", "solution", in.Solution, true},
		{"Implementation decisions", "implementation_decisions", in.ImplementationDecisions, true},
		{"Testing decisions", "testing_decisions", in.TestingDecisions, true},
		{"Out of scope", "out_of_scope", in.OutOfScope, true},
		{"Assumptions", "assumptions", in.Assumptions, false},
		{"Further notes", "further_notes", in.FurtherNotes, false},
	} {
		if s.text == nil {
			continue
		}
		body, err := sectionBody(s.field, *s.text, s.required)
		if err != nil {
			return nil, err
		}
		if body == nil {
			doc.RemoveSection(d, s.name)
		} else {
			doc.ReplaceSection(d, s.name, body, doc.SpecSectionOrder)
		}
	}

	sp.Reparse()
	if len(in.EditStories) > 0 || len(in.WithdrawStories) > 0 {
		targets := map[string]string{}
		for id, text := range in.EditStories {
			text, err := checkStoryText(text)
			if err != nil {
				return nil, err
			}
			targets[normalStoryID(id)] = text
		}
		for _, id := range in.WithdrawStories {
			key := normalStoryID(id)
			if _, dup := targets[key]; dup {
				return nil, apperr.Usage("invalid_input", "%s is both edited and withdrawn", id)
			}
			targets[key] = fmt.Sprintf("Withdrawn in revision %d.", rev)
		}
		for key, text := range targets {
			st, ok := sp.Story(key)
			if !ok {
				return nil, apperr.Usage("unknown_story", "the spec has no story %q", key)
			}
			if st.Withdrawn {
				return nil, apperr.Usage("invalid_input", "%s was withdrawn in revision %d and can't change", key, st.WithdrawnIn)
			}
			d.Lines[st.Line] = fmt.Sprintf("%s: %s", key, text)
		}
		sp.Reparse()
	}
	if len(in.AddStories) > 0 {
		sec, _ := sp.S.Section("User stories")
		at := sec.Heading + 1
		if n := len(sp.Stories); n > 0 {
			at = sp.Stories[n-1].Line + 1
		}
		next := len(sp.Stories) + 1
		var add []string
		for _, text := range in.AddStories {
			text, err := checkStoryText(text)
			if err != nil {
				return nil, err
			}
			add = append(add, "", fmt.Sprintf("US %d: %s", next, text))
			next++
		}
		d.Lines = append(d.Lines[:at], append(add, d.Lines[at:]...)...)
	}
	if in.AttemptLimit != nil {
		_ = doc.SetMeta(d, "Attempt limit", strconv.Itoa(*in.AttemptLimit), doc.SpecMetaOrder)
	}
	if in.Change == "revision" {
		_ = doc.SetMeta(d, "Revision", strconv.Itoa(rev), doc.SpecMetaOrder)
		if err := doc.AppendListItem(d, "Change history", fmt.Sprintf("- Revision %d: %s", rev, strings.TrimSpace(in.History))); err != nil {
			return nil, apperr.New("render_failed", "%v", err)
		}
	}

	sp.Reparse()
	if len(sp.Problems) > 0 {
		return nil, apperr.New("render_failed", "the edit would leave the spec malformed").With("problems", sp.Problems)
	}
	data := d.Bytes()
	r := &WriteResult{Feature: id, Operation: "spec edit", Hash: doc.Hash(raw), Revision: sp.Revision, Changed: []string{}, CommitSubject: "spec: revise " + id}
	if string(data) == string(raw) {
		return tr.finish(r), nil
	}
	return tr.writeResult(r, path, data, false)
}

func normalStoryID(id string) string {
	if covers, err := doc.ParseCovers(id); err == nil && len(covers) == 1 {
		return covers[0]
	}
	return strings.TrimSpace(id)
}

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func slugify(title string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(title) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	s := strings.Trim(b.String(), "-")
	if len(s) > 50 {
		s = strings.Trim(s[:50], "-")
		if i := strings.LastIndex(s, "-"); i > 20 {
			s = s[:i]
		}
	}
	if s == "" {
		return "ticket"
	}
	return s
}

func (f *Feature) checkCovers(t *TicketDoc, values []string) ([]string, error) {
	covers, err := doc.ParseCovers(strings.Join(values, ", "))
	if len(values) == 0 {
		covers, err = []string{}, nil
	} else if len(values) == 1 && strings.TrimSpace(values[0]) == "none" {
		covers, err = []string{}, nil
	}
	if err != nil {
		return nil, apperr.Usage("invalid_input", "`covers` %v", err)
	}
	if problems := f.coverProblems(t, covers); len(problems) > 0 {
		return nil, apperr.New("invalid_reference", "%s", problems[0].Message).With("problems", problems)
	}
	return covers, nil
}

func (f *Feature) checkBlockers(self int, refs []TicketRef) ([]int, error) {
	idx := f.byNumber()
	var out []int
	seen := map[int]bool{}
	for _, r := range refs {
		n := int(r)
		if seen[n] {
			continue
		}
		seen[n] = true
		if n == self {
			return nil, apperr.New("invalid_reference", "a ticket can't block itself")
		}
		if _, err := f.Ticket(n); err != nil {
			return nil, apperr.New("invalid_reference", "blocker %s: %v", doc.TicketID(n), err)
		}
		if idx[n].Status == doc.StatusCancelled {
			return nil, apperr.New("invalid_reference", "blocker %s is cancelled", doc.TicketID(n))
		}
		out = append(out, n)
	}
	edges := f.blockerEdges()
	edges[self] = out
	if cycle := newCycle(f.blockerEdges(), edges); cycle != nil {
		return nil, apperr.New("invalid_reference", "blockers would form a cycle: %s", strings.Join(cycle, " -> "))
	}
	return out, nil
}

func criteriaLines(values []string, ticked map[string]bool) ([]string, error) {
	if len(values) == 0 {
		return nil, apperr.Usage("invalid_input", "`criteria` needs at least one criterion")
	}
	seen := map[string]bool{}
	var lines []string
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" || strings.ContainsAny(v, "\r\n") {
			return nil, apperr.Usage("invalid_input", "each criterion must be one non-empty line")
		}
		if seen[v] {
			return nil, apperr.Usage("invalid_input", "criterion %q appears twice", v)
		}
		seen[v] = true
		mark := " "
		if ticked[v] {
			mark = "x"
		}
		lines = append(lines, fmt.Sprintf("- [%s] %s", mark, v))
	}
	return lines, nil
}

// TicketCreate adds a ticket with the next unused number.
func (tr *Tracker) TicketCreate(id string, in *TicketCreateInput) (*WriteResult, error) {
	title, err := checkTitle(in.Title)
	if err != nil {
		return nil, err
	}
	if in.Kind != doc.KindImplementation && in.Kind != doc.KindAcceptance && in.Kind != doc.KindManual {
		return nil, apperr.Usage("invalid_input", "`kind` must be implementation, acceptance, or manual")
	}
	if in.Covers == nil {
		return nil, apperr.Usage("invalid_input", "`covers` is required; use [] for a ticket that covers no story")
	}
	if in.SpecRevision != 0 && in.Kind != doc.KindAcceptance {
		return nil, apperr.Usage("invalid_input", "`spec_revision` applies only to acceptance tickets")
	}
	slug := in.Slug
	if slug == "" {
		slug = slugify(title)
	} else if !slugRe.MatchString(slug) {
		return nil, apperr.Usage("invalid_input", "`slug` must be lowercase words joined by hyphens")
	}
	what, err := sectionBody("what_to_build", in.WhatToBuild, true)
	if err != nil {
		return nil, err
	}
	criteria, err := criteriaLines(in.Criteria, nil)
	if err != nil {
		return nil, err
	}
	verification, err := sectionBody("verification", in.Verification, true)
	if err != nil {
		return nil, err
	}
	outOfScope, err := sectionBody("out_of_scope", in.OutOfScope, false)
	if err != nil {
		return nil, err
	}

	unlock, err := tr.lock(id, true)
	if err != nil {
		return nil, err
	}
	defer unlock()
	f, err := tr.load(id)
	if err != nil {
		return nil, err
	}
	n := f.nextNumber()
	draft := &TicketDoc{Ticket: &doc.Ticket{Number: n, Status: doc.StatusOpen}, FileNumber: n}
	covers, err := f.checkCovers(draft, *in.Covers)
	if err != nil {
		return nil, err
	}
	blockers, err := f.checkBlockers(n, in.BlockedBy)
	if err != nil {
		return nil, err
	}

	lines := []string{fmt.Sprintf("# %s: %s", doc.TicketID(n), title), "", "Kind: " + in.Kind}
	if in.Kind == doc.KindImplementation {
		lines = append(lines, "Phase: implement")
	}
	lines = append(lines, "Status: open")
	if in.Kind == doc.KindImplementation {
		lines = append(lines, "Attempt: 0")
	}
	if len(blockers) > 0 {
		lines = append(lines, "Blocked by: "+doc.FormatBlockedBy(blockers))
	}
	lines = append(lines, "Covers: "+doc.FormatCovers(covers))
	if in.Kind == doc.KindAcceptance {
		rev := in.SpecRevision
		if rev == 0 {
			rev = f.Spec.Revision
		}
		lines = append(lines, "Spec revision: "+strconv.Itoa(rev))
	}
	section := func(name string, body []string) {
		lines = append(lines, "", "## "+name, "")
		lines = append(lines, body...)
	}
	section("What to build", what)
	section("Acceptance criteria", criteria)
	section("Verification", verification)
	if outOfScope != nil {
		section("Out of scope", outOfScope)
	}
	lines = append(lines, "", "## Comments")
	data := []byte(strings.Join(lines, "\n") + "\n")
	if t := doc.ParseTicket(data); len(t.Problems) > 0 {
		return nil, apperr.New("render_failed", "the new ticket would be malformed").With("problems", t.Problems)
	}
	path := fmt.Sprintf("%s/%s-%s.md", store.IssuesDir(id), doc.TicketID(n), slug)
	if err := checkNoRegressions(f, f.withTicket(path, n, data)); err != nil {
		return nil, err
	}
	r := &WriteResult{Feature: id, Ticket: doc.TicketID(n), Operation: "ticket create", Status: doc.StatusOpen, CommitSubject: "tickets: create " + doc.TicketID(n)}
	return tr.writeResult(r, path, data, true)
}

// TicketEdit changes authored ticket content. It can't change state fields
// or comments, and it refuses done and cancelled tickets.
func (tr *Tracker) TicketEdit(id string, n int, ifMatch string, in *TicketEditInput) (*WriteResult, error) {
	if err := requireIfMatch(ifMatch); err != nil {
		return nil, err
	}
	if *in == (TicketEditInput{}) {
		return nil, apperr.Usage("invalid_input", "the input changes nothing")
	}
	unlock, err := tr.lock(id, true)
	if err != nil {
		return nil, err
	}
	defer unlock()
	f, err := tr.load(id)
	if err != nil {
		return nil, err
	}
	cur, err := f.Ticket(n)
	if err != nil {
		return nil, err
	}
	if err := checkCurrent(cur.Path, cur.Hash, ifMatch); err != nil {
		return nil, err
	}
	if len(cur.Problems) > 0 || cur.Number != cur.FileNumber {
		return nil, apperr.New("invalid_ticket", "ticket %s has problems; fix them first", doc.TicketID(n)).With("problems", cur.Problems)
	}
	if cur.IsFinished() {
		return nil, apperr.New("ticket_finished", "ticket %s is %s; put changes to finished work in a new ticket", doc.TicketID(n), cur.Status)
	}
	t := doc.ParseTicket(cur.Raw)
	d := t.Doc

	if in.Title != nil {
		title, err := checkTitle(*in.Title)
		if err != nil {
			return nil, err
		}
		_ = doc.SetTitle(d, doc.TicketID(n)+": "+title)
	}
	if in.BlockedBy != nil {
		blockers, err := f.checkBlockers(n, *in.BlockedBy)
		if err != nil {
			return nil, err
		}
		if len(blockers) == 0 {
			doc.RemoveMeta(d, "Blocked by")
		} else {
			_ = doc.SetMeta(d, "Blocked by", doc.FormatBlockedBy(blockers), doc.TicketMetaOrder)
		}
	}
	if in.Covers != nil {
		covers, err := f.checkCovers(cur, *in.Covers)
		if err != nil {
			return nil, err
		}
		_ = doc.SetMeta(d, "Covers", doc.FormatCovers(covers), doc.TicketMetaOrder)
	}
	if in.SpecRevision != nil {
		if cur.Kind != doc.KindAcceptance {
			return nil, apperr.Usage("invalid_input", "`spec_revision` applies only to acceptance tickets")
		}
		if *in.SpecRevision < 1 || *in.SpecRevision > f.Spec.Revision {
			return nil, apperr.Usage("invalid_input", "`spec_revision` must be between 1 and the spec's revision %d", f.Spec.Revision)
		}
		_ = doc.SetMeta(d, "Spec revision", strconv.Itoa(*in.SpecRevision), doc.TicketMetaOrder)
	}
	for _, s := range []struct {
		name, field string
		text        *string
		required    bool
	}{
		{"What to build", "what_to_build", in.WhatToBuild, true},
		{"Verification", "verification", in.Verification, true},
		{"Out of scope", "out_of_scope", in.OutOfScope, false},
	} {
		if s.text == nil {
			continue
		}
		body, err := sectionBody(s.field, *s.text, s.required)
		if err != nil {
			return nil, err
		}
		if body == nil {
			doc.RemoveSection(d, s.name)
		} else {
			doc.ReplaceSection(d, s.name, body, doc.TicketSectionOrder)
		}
	}
	if in.Criteria != nil {
		ticked := map[string]bool{}
		for _, c := range cur.Criteria {
			ticked[c.Text] = c.Checked
		}
		lines, err := criteriaLines(*in.Criteria, ticked)
		if err != nil {
			return nil, err
		}
		t.Reparse()
		first, last := t.Criteria[0].Line, t.Criteria[len(t.Criteria)-1].Line
		for i := first; i <= last; i++ {
			if !criterionLine(d.Lines[i]) && strings.TrimSpace(d.Lines[i]) != "" {
				return nil, apperr.New("unsupported_edit", "ticket %s has other text between its criteria, which a criteria edit would lose; edit that section by hand", doc.TicketID(n))
			}
		}
		d.Lines = append(d.Lines[:first], append(lines, d.Lines[last+1:]...)...)
	}

	t.Reparse()
	if len(t.Problems) > 0 {
		return nil, apperr.New("render_failed", "the edit would leave ticket %s malformed", doc.TicketID(n)).With("problems", t.Problems)
	}
	data := d.Bytes()
	r := &WriteResult{Feature: id, Ticket: doc.TicketID(n), Operation: "ticket edit", Hash: cur.Hash, Changed: []string{}, CommitSubject: "tickets: edit " + doc.TicketID(n)}
	ticketState(r, t)
	if string(data) == string(cur.Raw) {
		return tr.finish(r), nil
	}
	if err := checkNoRegressions(f, f.withTicket(cur.Path, n, data)); err != nil {
		return nil, err
	}
	return tr.writeResult(r, cur.Path, data, false)
}

func criterionLine(line string) bool {
	return strings.HasPrefix(line, "- [ ] ") || strings.HasPrefix(line, "- [x] ") || strings.HasPrefix(line, "- [X] ")
}

package tracker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/apperr"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/doc"
)

type opSpec struct {
	label  string
	fields []string
}

var opSpecs = map[string]opSpec{
	"claim":            {"claim", []string{"commit"}},
	"hand-off":         {"hand off", []string{"commit", "criteria_met", "verification", "notes"}},
	"pass":             {"pass", []string{"criteria_met", "all_criteria_met", "verification", "evidence", "not_blocking", "resolved", "notes"}},
	"send-back":        {"send back", []string{"criteria_unmet", "findings", "verification", "not_blocking", "resolved", "notes"}},
	"set-feature-base": {"set feature base", []string{"commit"}},
	"start-review":     {"start review", []string{"commit"}},
	"stop":             {"stop", []string{"reason", "criteria_unmet", "findings", "verification", "evidence", "not_blocking", "resolved", "notes"}},
	"resume":           {"resume", []string{"reason"}},
	"reset-attempts":   {"reset attempts", []string{"reason"}},
	"revise":           {"revise", []string{"reason"}},
	"complete":         {"complete", []string{"evidence", "notes"}},
	"cancel":           {"cancel", []string{"reason", "notes"}},
	"comment":          {"comment", []string{"not_blocking", "resolved", "notes"}},
}

// Operations lists the state operations in documentation order.
var Operations = []string{"claim", "hand-off", "pass", "send-back", "set-feature-base", "start-review", "stop", "resume", "reset-attempts", "revise", "complete", "cancel", "comment"}

// IsOperation reports whether name is a state operation.
func IsOperation(name string) bool {
	_, ok := opSpecs[name]
	return ok
}

// OperationFields returns the input fields an operation accepts.
func OperationFields(name string) []string { return opSpecs[name].fields }

// WriteResult describes a completed write.
type WriteResult struct {
	Feature       string            `json:"feature"`
	Ticket        string            `json:"ticket,omitempty"`
	Operation     string            `json:"operation"`
	Unchanged     bool              `json:"unchanged,omitempty"`
	Hash          string            `json:"hash"`
	Revision      int               `json:"revision,omitempty"`
	Phase         string            `json:"phase,omitempty"`
	Status        string            `json:"status,omitempty"`
	Attempt       *int              `json:"attempt,omitempty"`
	Changed       []string          `json:"changed"`
	Hashes        map[string]string `json:"hashes,omitempty"`
	CommitSubject string            `json:"commit_subject,omitempty"`
	GitIgnored    *bool             `json:"git_ignored,omitempty"`
	Note          string            `json:"note,omitempty"`
}

func (tr *Tracker) finish(r *WriteResult) *WriteResult {
	if len(r.Changed) == 0 {
		r.Unchanged = true
		r.CommitSubject = ""
		return r
	}
	if ignored, ok := tr.Store.IsIgnored(r.Changed[0]); ok {
		r.GitIgnored = &ignored
	}
	return r
}

func ticketState(r *WriteResult, t *doc.Ticket) {
	r.Phase, r.Status = t.Phase, t.Status
	if t.Kind == doc.KindImplementation {
		a := t.Attempt
		r.Attempt = &a
	}
}

func stateText(t *doc.Ticket) string {
	if t.Kind == doc.KindImplementation {
		return t.Phase + "/" + t.Status
	}
	return t.Status
}

func invalidTransition(op string, t *doc.Ticket, need string) error {
	return apperr.New("invalid_transition", "can't %s %s ticket %s while it is %s; it must be %s", opSpecs[op].label, t.Kind, t.ID(), stateText(t), need).
		With("kind", t.Kind).With("state", stateText(t))
}

func requireIfMatch(ifMatch string) error {
	if ifMatch == "" {
		return apperr.Usage("missing_if_match", "pass --if-match with the document hash from your last read")
	}
	return nil
}

func checkCurrent(path, have, want string) error {
	if have != want {
		return apperr.New("stale", "%s changed since hash %s; read it again", path, want).With("current_hash", have)
	}
	return nil
}

// Apply runs one state operation on a ticket.
func (tr *Tracker) Apply(id string, n int, op string, ifMatch string, in *OpInput) (*WriteResult, error) {
	spec, ok := opSpecs[op]
	if !ok {
		return nil, apperr.Usage("unknown_operation", "%q is not a ticket operation", op)
	}
	if err := requireIfMatch(ifMatch); err != nil {
		return nil, err
	}
	if in == nil {
		in = &OpInput{}
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
		problems := cur.Problems
		if cur.Number != cur.FileNumber {
			problems = append(problems, doc.Problem{Code: "number_mismatch", Message: "the title number doesn't match the file name"})
		}
		return nil, apperr.New("invalid_ticket", "ticket %s has problems; fix them first", doc.TicketID(n)).With("problems", problems)
	}

	t := doc.ParseTicket(cur.Raw)
	a := &applier{tr: tr, f: f, t: t, op: op, in: in, label: spec.label}
	if err := a.run(); err != nil {
		return nil, err
	}

	res := &WriteResult{Feature: id, Ticket: t.ID(), Operation: op, Hash: cur.Hash, Changed: []string{}, CommitSubject: fmt.Sprintf("tickets: %s %s", spec.label, t.ID()), Note: a.note}
	t.Reparse()
	if len(t.Problems) > 0 {
		return nil, apperr.New("render_failed", "the operation would leave ticket %s malformed", t.ID()).With("problems", t.Problems)
	}
	ticketState(res, t)

	type pending struct {
		path   string
		id     string
		number int
		data   []byte
	}
	var writes []pending
	if out := t.Doc.Bytes(); string(out) != string(cur.Raw) {
		writes = append(writes, pending{cur.Path, t.ID(), cur.FileNumber, out})
	}
	for _, o := range a.others {
		writes = append(writes, pending{o.path, o.id, o.number, o.data})
	}
	after := f
	for _, w := range writes {
		after = after.withTicket(w.path, w.number, w.data)
	}
	if err := checkNoRegressions(f, after); err != nil {
		return nil, err
	}
	var written []string
	for i, w := range writes {
		if err := tr.Store.Write(w.path, w.data, false); err != nil {
			if i == 0 {
				return nil, err
			}
			var left []string
			for _, p := range writes[i:] {
				left = append(left, p.path)
			}
			return nil, apperr.New("partial_write", "wrote %s but not %s: %v; run the same operation again to finish", strings.Join(written, ", "), strings.Join(left, ", "), err).
				With("written", written).With("pending", left)
		}
		written = append(written, w.path)
		res.Changed = append(res.Changed, w.path)
		h := doc.Hash(w.data)
		if w.path == cur.Path {
			res.Hash = h
		} else {
			if res.Hashes == nil {
				res.Hashes = map[string]string{}
			}
			res.Hashes[w.id] = h
		}
	}
	return tr.finish(res), nil
}

type otherWrite struct {
	path   string
	id     string
	number int
	data   []byte
}

type applier struct {
	tr     *Tracker
	f      *Feature
	t      *doc.Ticket
	op     string
	in     *OpInput
	label  string
	items  []commentItem
	note   string
	others []otherWrite
}

func (a *applier) run() error {
	t, in := a.t, a.in
	impl := t.Kind == doc.KindImplementation
	attempt := t.Attempt
	comment := true
	attemptLabel := impl

	switch a.op {
	case "claim":
		comment = false
		if !impl {
			return invalidTransition(a.op, t, "an implementation ticket")
		}
		if t.Phase == doc.PhaseImplement && t.Status == doc.StatusInProgress {
			a.note = "the ticket was already claimed; the current attempt continues"
			return nil
		}
		if t.Phase != doc.PhaseImplement || t.Status != doc.StatusOpen {
			return invalidTransition(a.op, t, "implement/open or implement/in-progress")
		}
		if t.Attempt >= a.f.Spec.AttemptLimit {
			return apperr.New("attempt_limit_reached", "ticket %s has used %d of %d attempts", t.ID(), t.Attempt, a.f.Spec.AttemptLimit)
		}
		if err := a.requireBlockersDone(); err != nil {
			return err
		}
		if t.Base == "" {
			sha, err := a.commit()
			if err != nil {
				return err
			}
			a.setMeta("Base commit", sha)
		} else if in.Commit != "" {
			a.note = "the ticket already has a base commit, so it was kept"
		}
		a.setMeta("Attempt", strconv.Itoa(t.Attempt+1))
		a.setMeta("Status", doc.StatusInProgress)

	case "hand-off":
		if !impl || t.Phase != doc.PhaseImplement || t.Status != doc.StatusInProgress {
			return invalidTransition(a.op, t, "an implementation ticket in implement/in-progress")
		}
		sha, err := a.commit()
		if err != nil {
			return err
		}
		if len(in.Verification) == 0 {
			return apperr.Usage("invalid_input", "hand-off needs `verification` with each command and its result")
		}
		if err := a.tick(in.CriteriaMet, true); err != nil {
			return err
		}
		a.setMeta("Target commit", sha)
		a.setMeta("Phase", doc.PhaseReview)
		a.setMeta("Status", doc.StatusOpen)

	case "pass":
		switch {
		case impl && t.Phase == doc.PhaseReview && t.Status == doc.StatusOpen:
		case t.Kind == doc.KindAcceptance && t.Status == doc.StatusInProgress:
			if err := a.requireCurrentAcceptance(); err != nil {
				return err
			}
			if err := a.requireBlockersDone(); err != nil {
				return err
			}
		default:
			return invalidTransition(a.op, t, "an implementation ticket in review/open or an acceptance ticket in progress")
		}
		if err := a.tick(in.CriteriaMet, true); err != nil {
			return err
		}
		if in.AllCriteriaMet {
			setAllCriteria(t, true)
		}
		a.setMeta("Status", doc.StatusDone)

	case "send-back":
		if !impl || t.Phase != doc.PhaseReview || t.Status != doc.StatusOpen {
			return invalidTransition(a.op, t, "an implementation ticket in review/open")
		}
		if len(in.CriteriaUnmet) == 0 && len(in.Findings) == 0 {
			return apperr.Usage("invalid_input", "send-back needs `criteria_unmet` or `findings`")
		}
		a.setMeta("Phase", doc.PhaseImplement)
		a.setMeta("Status", doc.StatusOpen)

	case "set-feature-base":
		comment = false
		if t.Kind != doc.KindAcceptance || t.Status != doc.StatusOpen {
			return invalidTransition(a.op, t, "an open acceptance ticket")
		}
		if t.Base != "" {
			return apperr.New("base_already_set", "ticket %s already has base commit %s, which never changes", t.ID(), t.Base)
		}
		sha, err := a.commit()
		if err != nil {
			return err
		}
		a.setMeta("Base commit", sha)

	case "start-review":
		comment = false
		if t.Kind != doc.KindAcceptance {
			return invalidTransition(a.op, t, "an acceptance ticket")
		}
		if t.Status == doc.StatusInProgress {
			a.note = "the review was already started; it continues"
			return nil
		}
		if t.Status != doc.StatusOpen {
			return invalidTransition(a.op, t, "open or in-progress")
		}
		if err := a.requireCurrentAcceptance(); err != nil {
			return err
		}
		if t.Base == "" {
			return apperr.New("missing_base", "ticket %s has no base commit; run set-feature-base first", t.ID())
		}
		if err := a.requireBlockersDone(); err != nil {
			return err
		}
		sha, err := a.commit()
		if err != nil {
			return err
		}
		a.setMeta("Target commit", sha)
		a.setMeta("Status", doc.StatusInProgress)

	case "stop":
		if t.Status != doc.StatusOpen && t.Status != doc.StatusInProgress {
			return invalidTransition(a.op, t, "open or in-progress")
		}
		if err := a.requireReason(); err != nil {
			return err
		}
		a.setMeta("Status", doc.StatusStopped)

	case "resume":
		if t.Status != doc.StatusStopped {
			return invalidTransition(a.op, t, "stopped")
		}
		if err := a.requireReason(); err != nil {
			return err
		}
		a.setMeta("Status", doc.StatusOpen)

	case "reset-attempts":
		if !impl || t.Phase != doc.PhaseImplement || t.Status != doc.StatusStopped {
			return invalidTransition(a.op, t, "an implementation ticket in implement/stopped")
		}
		if err := a.requireReason(); err != nil {
			return err
		}
		a.setMeta("Attempt", "0")
		a.setMeta("Status", doc.StatusOpen)

	case "revise":
		if t.Status != doc.StatusOpen && t.Status != doc.StatusInProgress && t.Status != doc.StatusStopped {
			return invalidTransition(a.op, t, "open, in-progress, or stopped")
		}
		if err := a.requireReason(); err != nil {
			return err
		}
		setAllCriteria(t, false)
		if impl {
			a.setMeta("Phase", doc.PhaseImplement)
			a.setMeta("Attempt", "0")
		}
		a.setMeta("Status", doc.StatusOpen)

	case "complete":
		if t.Kind != doc.KindManual || t.Status != doc.StatusOpen {
			return invalidTransition(a.op, t, "an open manual ticket")
		}
		if len(in.Evidence) == 0 {
			return apperr.Usage("invalid_input", "complete needs `evidence`")
		}
		setAllCriteria(t, true)
		a.setMeta("Status", doc.StatusDone)

	case "cancel":
		if t.Status == doc.StatusDone {
			return invalidTransition(a.op, t, "anything except done")
		}
		if t.Status == doc.StatusCancelled {
			comment = false
			a.note = "the ticket was already cancelled; only its remaining blocker references were removed"
		} else {
			a.setMeta("Status", doc.StatusCancelled)
		}
		if err := a.unblockOthers(); err != nil {
			return err
		}

	case "comment":
		if len(in.NotBlocking) == 0 && len(in.Resolved) == 0 && len(in.Notes) == 0 {
			return apperr.Usage("invalid_input", "comment needs `not_blocking`, `resolved`, or `notes`")
		}
		if len(in.NotBlocking) == 0 && len(in.Notes) == 0 {
			attemptLabel = false
		}
	}

	if !comment {
		return nil
	}
	if err := a.buildItems(); err != nil {
		return err
	}
	header := a.label
	if attemptLabel {
		header = fmt.Sprintf("%s (attempt %d)", a.label, attempt)
	}
	lines := renderComment(header, a.items)
	if _, unclosed := doc.FenceMask(lines); unclosed {
		return apperr.Usage("invalid_input", "the comment would have a fenced code block that is never closed")
	}
	return doc.AppendToSection(t.Doc, "Comments", lines)
}

func (a *applier) setMeta(key, value string) {
	_ = doc.SetMeta(a.t.Doc, key, value, doc.TicketMetaOrder)
	a.t.Reparse()
}

func (a *applier) commit() (string, error) {
	if a.in.Commit == "" {
		return "", apperr.Usage("invalid_input", "%s needs `commit` (or --commit), such as HEAD or a full SHA", a.op)
	}
	return a.tr.Store.ResolveCommit(a.in.Commit)
}

func (a *applier) requireReason() error {
	if strings.TrimSpace(a.in.Reason) == "" {
		return apperr.Usage("invalid_input", "%s needs `reason`", a.op)
	}
	return nil
}

// requireCurrentAcceptance checks, against the current spec and tickets,
// that an acceptance ticket is still the one to review and checks the
// spec's current revision.
func (a *applier) requireCurrentAcceptance() error {
	t := a.t
	if cur := a.f.CurrentAcceptance(); cur == nil || cur.FileNumber != t.Number {
		return apperr.New("not_current_acceptance", "ticket %s isn't the current acceptance ticket", t.ID())
	}
	if t.SpecRevision != a.f.Spec.Revision {
		return apperr.New("stale_acceptance", "ticket %s checks spec revision %d but the spec is at revision %d", t.ID(), t.SpecRevision, a.f.Spec.Revision)
	}
	return nil
}

func (a *applier) requireBlockersDone() error {
	cur, _ := a.f.Ticket(a.t.Number)
	if open := a.f.unfinishedBlockers(cur); len(open) > 0 {
		return apperr.New("not_ready", "ticket %s is blocked by %s, which aren't done", a.t.ID(), strings.Join(open, ", ")).With("blocked_by", open)
	}
	return nil
}

func (a *applier) tick(texts []string, checked bool) error {
	for _, text := range texts {
		i, err := findCriterion(a.t, text)
		if err != nil {
			return err
		}
		setCriterion(a.t, i, checked)
	}
	return nil
}

func (a *applier) unblockOthers() error {
	for _, o := range a.f.Tickets {
		if o.FileNumber == a.t.Number {
			continue
		}
		var keep []int
		found := false
		for _, b := range o.BlockedBy {
			if b == a.t.Number {
				found = true
			} else {
				keep = append(keep, b)
			}
		}
		if !found {
			continue
		}
		if len(o.Problems) > 0 {
			return apperr.New("invalid_ticket", "ticket %s is blocked by %s but has problems, so cancel changed nothing", doc.TicketID(o.FileNumber), a.t.ID()).With("problems", o.Problems)
		}
		w := doc.ParseTicket(o.Raw)
		if len(keep) == 0 {
			doc.RemoveMeta(w.Doc, "Blocked by")
		} else {
			_ = doc.SetMeta(w.Doc, "Blocked by", doc.FormatBlockedBy(keep), doc.TicketMetaOrder)
		}
		w.Reparse()
		if len(w.Problems) > 0 {
			return apperr.New("render_failed", "removing %s from ticket %s would leave it malformed", a.t.ID(), doc.TicketID(o.FileNumber)).With("problems", w.Problems)
		}
		a.others = append(a.others, otherWrite{path: o.Path, id: doc.TicketID(o.FileNumber), number: o.FileNumber, data: w.Doc.Bytes()})
	}
	return nil
}

func (a *applier) buildItems() error {
	in := a.in
	if r := strings.TrimSpace(in.Reason); r != "" {
		vals, err := checkValues("reason", []string{r}, false)
		if err != nil {
			return err
		}
		label := "Reason"
		if a.op == "resume" {
			label = "Resolution"
		}
		a.items = append(a.items, commentItem{label: label, values: vals})
	}
	if len(in.CriteriaUnmet) > 0 {
		var quoted []string
		for _, text := range in.CriteriaUnmet {
			i, err := findCriterion(a.t, text)
			if err != nil {
				return err
			}
			setCriterion(a.t, i, false)
			quoted = append(quoted, strconv.Quote(a.t.Criteria[i].Text))
		}
		a.items = append(a.items, commentItem{label: "Unmet criteria", values: quoted})
	}
	if len(in.Findings) > 0 {
		vals, err := checkValues("findings", in.Findings, false)
		if err != nil {
			return err
		}
		for _, v := range vals {
			if strings.HasPrefix(strings.ToLower(v), "not blocking") {
				return apperr.Usage("invalid_input", "put not-blocking findings in `not_blocking`, not `findings`")
			}
		}
		a.items = append(a.items, commentItem{values: vals, raw: true})
	}
	for _, f := range []struct {
		name   string
		label  string
		values []string
	}{{"verification", "Verification", in.Verification}, {"evidence", "Evidence", in.Evidence}} {
		if len(f.values) == 0 {
			continue
		}
		vals, err := checkValues(f.name, f.values, false)
		if err != nil {
			return err
		}
		a.items = append(a.items, commentItem{label: f.label, values: vals, list: true})
	}
	if len(in.NotBlocking) > 0 {
		vals, err := checkValues("not_blocking", in.NotBlocking, true)
		if err != nil {
			return err
		}
		for _, v := range vals {
			a.items = append(a.items, commentItem{label: "Not blocking", values: []string{v}})
		}
	}
	if len(in.Resolved) > 0 {
		open := map[string]bool{}
		unresolved := a.t.UnresolvedNotBlocking()
		for _, f := range unresolved {
			open[doc.FindingKey(f)] = true
		}
		for _, r := range in.Resolved {
			vals, err := checkValues("resolved.finding", []string{r.Finding}, true)
			if err != nil {
				return err
			}
			finding := strings.TrimSuffix(vals[0], ".")
			if !open[doc.FindingKey(finding)] {
				return apperr.New("unknown_finding", "ticket %s has no unresolved not-blocking finding %q", a.t.ID(), r.Finding).With("unresolved", unresolved)
			}
			delete(open, doc.FindingKey(finding))
			var text string
			switch r.Resolution {
			case "fixed", "ignored":
				if r.Ticket != 0 {
					return apperr.Usage("invalid_input", "`ticket` only applies to the `moved` resolution")
				}
				text = finding + ". " + strings.ToUpper(r.Resolution[:1]) + r.Resolution[1:] + "."
			case "moved":
				if r.Ticket == 0 {
					return apperr.Usage("invalid_input", "the `moved` resolution needs `ticket`")
				}
				if _, err := a.f.Ticket(int(r.Ticket)); err != nil {
					return err
				}
				text = fmt.Sprintf("%s. Moved to ticket %s.", finding, doc.TicketID(int(r.Ticket)))
			default:
				return apperr.Usage("invalid_input", "resolution must be fixed, ignored, or moved, not %q", r.Resolution)
			}
			a.items = append(a.items, commentItem{label: "Not blocking resolved", values: []string{text}})
		}
	}
	if len(in.Notes) > 0 {
		vals, err := checkValues("notes", in.Notes, false)
		if err != nil {
			return err
		}
		for _, v := range vals {
			a.items = append(a.items, commentItem{label: "Note", values: []string{v}})
		}
	}
	return nil
}

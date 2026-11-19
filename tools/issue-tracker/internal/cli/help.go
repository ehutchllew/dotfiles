package cli

import (
	"slices"
	"sort"
	"strings"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/tracker"
)

const mainHelp = `issue-tracker reads and writes feature specs and tickets.

Usage: issue-tracker [--repo DIR] [--pretty] COMMAND ...

Read commands:
  feature list                   List features and their titles.
  status FEATURE                 Compact state for orchestration, with full validation.
  validate FEATURE [--draft]     Validate a feature. --draft skips whole-feature checks.
  spec get FEATURE               Print the spec and its hash.
  ticket list FEATURE            Ticket summaries without bodies.
  ticket get FEATURE NN          Print a ticket and its hash.

Write commands (JSON input through --input FILE, or --input - for stdin):
  spec create FEATURE            Create a spec at revision 1.
  spec edit FEATURE              Edit spec content, as an editorial change or a revision.
  ticket create FEATURE          Add a ticket with the next number.
  ticket edit FEATURE NN         Edit an unfinished ticket's content.
  ticket OPERATION FEATURE NN    Apply a state operation:
      ` + "{{OPS}}" + `

Every command prints one JSON object. Failures print {"error":{"code":...}}
and exit 1, or 2 for bad arguments or input.

Run "issue-tracker help COMMAND", such as "issue-tracker help ticket send-back",
for flags and an input example.
`

var commandHelp = map[string]string{
	"feature list": `Usage: issue-tracker feature list

Lists every directory in .scratch/ that has a spec.md.

Output: {"features":[{"feature":"order-summary","title":"Order summary"}]}
`,
	"status": `Usage: issue-tracker status FEATURE

Prints compact state without document bodies: spec revision, spec hash,
attempt limit, each ticket's metadata, hash, and readiness, the current
acceptance ticket, full validation results, and every unresolved
not-blocking finding.

"valid": false means the state isn't safe to act on. Read "errors".
"last_comment" appears on stopped tickets. "earlier_acceptance_base" is
the base commit of the most recent earlier acceptance ticket, if any.
`,
	"validate": `Usage: issue-tracker validate FEATURE [--draft]

Runs every validation check and exits 1 with code validation_failed and
the error list when any fails.

--draft  Skip checks that need the full breakdown: story coverage by
         implementation tickets and the current acceptance ticket's
         coverage, spec revision, and blockers.
`,
	"spec get": `Usage: issue-tracker spec get FEATURE [--if-changed HASH]

Prints the spec body and its hash.

--if-changed HASH  If the spec still has this hash, print only
                   {"hash":...,"unchanged":true}. Pass only a hash whose
                   content you still have in context.
`,
	"spec create": `Usage: issue-tracker spec create FEATURE --input FILE

Creates .scratch/FEATURE/spec.md at revision 1. Stories are numbered
US 1, US 2, ... in order. Optional: attempt_limit (default 3),
assumptions, further_notes, history (default "Initial spec.").

Input:
{"title":"Order summary",
 "problem_statement":"...",
 "solution":"...",
 "user_stories":["As a shopper, I want ..., so that ..."],
 "implementation_decisions":"- ...",
 "testing_decisions":"- Run ` + "`npm test`" + `.",
 "out_of_scope":"- ...",
 "assumptions":"- ...",
 "history":"Initial spec."}
`,
	"spec edit": `Usage: issue-tracker spec edit FEATURE --input FILE --if-match HASH

Replaces the given sections and leaves the rest unchanged.

change      "editorial" or "revision". Required for any content change.
            You decide which; the CLI doesn't judge wording.
approved    true. Required for a revision: the user approved it.
history     One line for the Change history entry. Revision only.

Sections: title, problem_statement, solution, implementation_decisions,
testing_decisions, out_of_scope, assumptions, further_notes. Set
assumptions or further_notes to "" to remove them.

Stories: add_stories (revision only) appends with new IDs.
edit_stories maps an ID to new text. withdraw_stories (revision only)
replaces the text with "Withdrawn in revision N.". IDs never change.

attempt_limit changes the limit without a revision; omit change for it.

Input:
{"change":"revision","approved":true,
 "history":"US 2 added: show discounts.",
 "add_stories":["As a shopper, I want ..., so that ..."],
 "implementation_decisions":"- ..."}
`,
	"ticket list": `Usage: issue-tracker ticket list FEATURE

Prints each ticket's number, title, kind, phase, status, attempt,
blockers, covered stories, commits, readiness, hash, and path.
`,
	"ticket get": `Usage: issue-tracker ticket get FEATURE NN [--if-changed HASH] [--full]

Prints the ticket body, its hash, and its criteria with their IDs.
The hash always covers the whole file.

By default, comments before the ticket's last "revise" comment are left
out and "omitted_comments" says how many. Earlier blocking findings and
unresolved not-blocking findings since that revise stay in the body.

--full             Include every comment. Use for diagnosis and audits.
--if-changed HASH  If the ticket still has this hash, print only
                   {"hash":...,"unchanged":true}.
`,
	"ticket create": `Usage: issue-tracker ticket create FEATURE --input FILE

Creates .scratch/FEATURE/issues/NN-<slug>.md with the next unused number
and the starting state for its kind. covers is required; use [] for none.
Acceptance tickets get spec_revision from the spec unless given. Optional:
blocked_by, out_of_scope, slug.

Input:
{"title":"Show order item counts","kind":"implementation",
 "blocked_by":["01"],"covers":["US 1"],
 "what_to_build":"Show the total item quantity in the order summary.",
 "criteria":["An empty order shows an item count of zero."],
 "verification":"Run ` + "`npm test`" + `.",
 "out_of_scope":"- Order totals."}
`,
	"ticket edit": `Usage: issue-tracker ticket edit FEATURE NN --input FILE --if-match HASH

Changes content of an open, in-progress, or stopped ticket. It can't
change kind, phase, status, attempt, commits, or comments; use state
operations. Criteria whose text is unchanged keep their IDs and ticks.
An edit that would break validation fails with would_invalidate; for example,
revise an in-progress acceptance ticket before adding blockers to it.

Fields: title, blocked_by, covers, spec_revision (acceptance),
what_to_build, criteria, verification, out_of_scope ("" removes it).

Input:
{"criteria":["An empty order shows an item count of zero.",
             "The count updates when an item is removed."],
 "blocked_by":["01","04"]}
`,
}

var opHelp = map[string]string{
	"claim": `Implementation, implement/open -> implement/in-progress. Adds 1 to
Attempt and sets Base commit if missing. On implement/in-progress it
changes nothing and the attempt continues. Fails with not_ready when a
blocker isn't done and attempt_limit_reached at the spec's limit.

Input: {"commit":"HEAD"}, or pass --commit HEAD.`,
	"hand-off": `Implementation, implement/in-progress -> review/open. Sets Target
commit, ticks criteria_met, and records the verification summary.
Also finishes a pending hand-off from stopped without another claim.
Omit commit in that case to use the recorded commit.

Input:
{"commit":"HEAD",
 "criteria_met":["C01"],
 "verification":["Required checks passed."],
 "notes":["..."]}`,
	"pass": `Implementation review/open -> review/done, or acceptance
in-progress -> done. all_criteria_met ticks every criterion. An acceptance
pass fails with stale_acceptance if the spec was revised during the review,
and with not_current_acceptance if a newer acceptance ticket exists.

Input:
{"verification":["` + "`npm test`" + ` passed."],
 "evidence":["Ticket 02: phone screenshot recorded."],
 "all_criteria_met":true,
 "not_blocking":["The local n doesn't say it holds the item count."],
 "resolved":[{"finding":"...","resolution":"fixed"}]}`,
	"send-back": `Implementation, review/open -> implement/open. Unticks
criteria_unmet. Needs criteria_unmet or findings. Put each blocking
finding in findings as written, such as "Blocking (missing): ...".

Input:
{"criteria_unmet":["C01"],
 "findings":["Blocking (missing): an empty order shows no count line."],
 "verification":["` + "`npm test`" + ` passed."],
 "not_blocking":["..."],
 "resolved":[{"finding":"...","resolution":"fixed"}]}`,
	"set-feature-base": `Acceptance, open. Sets Base commit. Fails with base_already_set
if it has one.

Input: {"commit":"<sha>"}, or pass --commit HEAD.`,
	"start-review": `Acceptance, open -> in-progress. Sets Target commit. Needs the
current acceptance ticket, a Base commit, the spec's current revision,
and every blocker done. On in-progress it changes nothing.

Input: {"commit":"HEAD"}, or pass --commit HEAD.`,
	"stop": `Any kind, open or in-progress -> stopped. Phase doesn't change.
Unticks criteria_unmet. A failed acceptance review uses stop.

Input:
{"reason":"Verification blocked: the test database is down.",
 "criteria_unmet":["C01"],"findings":["Blocking (wrong): ..."],
 "verification":["..."]}

For a report-only problem on a claimed implementation:
{"reason":"The report used an unknown criterion ID.",
 "stop_kind":"handoff-pending","commit":"HEAD","spec_hash":"sha256:...",
 "verification":["` + "`npm test`" + ` passed."]}`,
	"resume": `Any kind, stopped -> open. Discards a pending hand-off; use hand-off
directly to submit its corrected report without starting a new attempt.

Input: {"reason":"<how the cause was resolved>"}`,
	"reset-attempts": `Implementation, implement/stopped -> implement/open. Sets
Attempt to 0.

Input: {"reason":"..."}`,
	"revise": `Unticks every criterion. Implementation: any phase, open,
in-progress, or stopped -> implement/open, and Attempt to 0. Acceptance
and manual: open, in-progress, or stopped -> open. Commits are kept.

Input: {"reason":"Spec revision 2 changed US 1."}`,
	"complete": `Manual, open -> done. Ticks every criterion.

Input: {"evidence":["Pixel 8, Chrome: count readable without zoom. Screenshot: ..."]}`,
	"cancel": `Any status except done -> cancelled. Removes the ticket from every
Blocked by line. Writes several files; run it again after a partial_write
to finish. On a cancelled ticket it only removes leftover references.

Input (optional): {"reason":"..."}`,
	"comment": `Any status. Appends a comment. A comment that only resolves
findings has no attempt label. resolution is fixed, ignored, or moved
(with ticket).

Input:
{"resolved":[{"finding":"The local n is unclear.","resolution":"moved","ticket":"05"}],
 "notes":["..."]}`,
}

func opUsage(op string) string {
	fields := tracker.OperationFields(op)
	commit := slices.Contains(fields, "commit")
	usage := "Usage: issue-tracker ticket " + op + " FEATURE NN --if-match HASH [--input FILE]"
	if commit {
		usage += " [--commit REF]"
	}
	text := usage + "\n\n" + opHelp[op] + "\n\nAccepted input fields: " + strings.Join(fields, ", ") + `

--if-match HASH  The ticket's hash from your last read. Required. A
                 mismatch fails with code stale and the current hash.
`
	if commit {
		text += `--commit REF     Shorthand for the commit field. HEAD and other refs are
                 resolved to a full SHA with git rev-parse.
`
	}
	return text
}

func helpFor(words []string) (string, bool) {
	key := strings.Join(words, " ")
	if len(words) == 0 || key == "help" {
		ops := append([]string(nil), tracker.Operations...)
		return strings.Replace(mainHelp, "{{OPS}}", strings.Join(ops, ", "), 1), true
	}
	if h, ok := commandHelp[key]; ok {
		return h, true
	}
	if len(words) == 2 && words[0] == "ticket" && tracker.IsOperation(words[1]) {
		return opUsage(words[1]), true
	}
	return "", false
}

func knownCommands() []string {
	var out []string
	for k := range commandHelp {
		out = append(out, k)
	}
	for _, op := range tracker.Operations {
		out = append(out, "ticket "+op)
	}
	sort.Strings(out)
	return out
}

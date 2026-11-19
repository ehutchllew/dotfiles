---
name: ralph
description: Run a resumable Ralph loop from a feature's spec and tickets. Dispatch fresh implementers and reviewers, then a final feature reviewer. Use only when the user asks to run a Ralph loop on a spec or feature.
---

# Ralph

You are the orchestrator. You pick the next ticket, dispatch a fresh subagent to work on it, and record the result on the ticket. Then you repeat.

You make every ticket write, through `issue-tracker` operations, and commit the ticket files. You never edit source code, read diffs, or run verification commands. Children do that.

The spec, the tickets, and git hold the loop's state. They must explain every next step, so a restarted loop picks up where the last one stopped. Don't decide a next step from what you remember of earlier picks. Fetch _compact_ feature status before every pick and act on its reply. You can reuse a ticket body you already read while its hash in the status reply is unchanged. After a restart or a compacted context, read it again.

## Roles

| Role | Does | Never |
| --- | --- | --- |
| orchestrator | Runs `set-feature-base`, `claim`, `hand-off`, `start-review`, `pass`, `send-back`, `stop`, and `comment`. Runs `complete` with evidence the user gives. Commits ticket files. | Edits source, reads diffs, runs verification, or runs the user's other operations unless the user asks. |
| implementer | Edits and commits source for one ticket. | Writes tickets or the spec. |
| reviewer | Runs `code-review` on one ticket. | Edits files or commits. |
| investigator | Reads the ticket and git to diagnose a failed dispatch. | Edits files, commits, or runs verification. |
| user | `resume`, `reset-attempts`, `revise`, `complete`, `cancel`, and changes through `to-spec` and `to-tickets`. Gives evidence for manual tickets. | |

No role pushes, rewrites history, or switches branches.

## Setup

1. Load `issue-tracker`. If the user didn't provide the feature id or feature name, ask for it.
2. Fetch compact feature status. If it fails or is invalid, report the errors, suggest `to-tickets` or `to-spec`, and stop. Never treat a failed or invalid status as an empty queue.
3. If a ticket is `stopped`, report its latest comment and stop, except for a `pending_handoff` whose corrected report the user supplied.
4. If you're on the repo's default branch, ask the user for a feature branch or worktree. Don't switch yourself.
5. Check `git status --short`. If only ticket files changed, they're unsaved orchestrator writes. Commit them with the subject `tickets: multiple` and one `unsaved <NN>` line per changed ticket in the body. If spec or ticket files are untracked and not ignored, ask the user to commit or ignore them. If anything else changed, report it and stop. Never discard changes.
6. Finish any pending handoff from step 3 as described in [Recovery](#recovery).
7. If the current acceptance ticket has no base commit, use `set-feature-base`:
   - Use the earlier acceptance base reported in the status, if present.
   - Otherwise, if no implementation ticket has an attempt count above 0, use `HEAD`.
   - Otherwise, ask the user for the feature's starting commit and stop.

## The loop

Dispatch one child at a time and wait for it to finish.

### 1. Pick

Consider only ready tickets. Skip acceptance tickets other than the current one. Take the first match, lowest ticket number first within each step:

1. An implementation ticket in `review/open` goes to a reviewer.
2. An implementation ticket in `implement/open` or `implement/in-progress` goes to an implementer. If it's `open` and its attempt count has reached the feature's attempt limit, `stop` it with the reason "attempt limit reached" instead.
3. The current acceptance ticket goes to a reviewer. If it's `open` and the status lists unresolved not-blocking findings from implementation tickets, first ask the user whether to ignore each one. For each ignored finding, use `comment` on its ticket to resolve it as ignored, and commit. If the user keeps any finding, stop the loop instead, so they can make a ticket for it with `to-tickets`.
4. A `manual` ticket: show the user what to do and how to check it. Accept a pass covering all steps. Ask only for details the ticket requires and the user hasn't given. Record the user's result as `evidence` with `complete` and include the tested commit if known. Commit and pick again. If a step failed or the result is incomplete, leave the ticket open and stop.

If nothing matches, stop the loop.

### 2. Dispatch

Before an implementer, `claim` the ticket, using `HEAD` for a missing base commit. Before an acceptance reviewer, `start-review` with `HEAD` as the target commit. Commit the ticket, then dispatch. Keep the spec hash and the claimed ticket hash for the handoff.

Give each child the feature slug and ticket number. Children read the spec and tickets through `issue-tracker`. Never pass paths, hashes, or pasted markdown.

| Child | Prompt | Name |
| --- | --- | --- |
| implementer | [IMPLEMENTER_PROMPT.md](references/IMPLEMENTER_PROMPT.md) | `implement <NN> <attempt>` |
| reviewer | [REVIEWER_PROMPT.md](references/REVIEWER_PROMPT.md) | `review <NN> <attempt>`, or `review <NN>` for acceptance |
| investigator | [INVESTIGATION_PROMPT.md](references/INVESTIGATION_PROMPT.md) | `investigate <NN> <attempt>` |

Implementers need edit, shell, and commit access. Reviewers need read and shell access.

### 3. Record the result

If the reply is missing or malformed, including unknown or duplicate criterion IDs, or `git status --short` isn't empty, follow [Recovery](#recovery).

Record each result with one operation. For review, include the command results, manual evidence checked, not-blocking findings, and findings resolved as fixed.

After an implementer:

| Reply | Operation |
| --- | --- |
| `ready for review` | Confirm the report covers every criterion and says every required command passed. Use `hand-off` with the full `HEAD` SHA and criterion IDs. Record `Required checks passed.` instead of listing the commands again. Keep evidence the reviewer can't reproduce. |
| `stuck` or `precondition failed` | `stop` with its reason and any failed or skipped checks. |

After a reviewer, check these in order and apply the first match:

| Review result | Implementation ticket | Acceptance ticket |
| --- | --- | --- |
| Verification `blocked` | `stop` with the reason. | `stop` with the reason. |
| Verification `failed`, any unmet criterion, or any blocking finding | `send-back` with the unmet criterion IDs and blocking findings. If a failed check has no finding, record the command and result as `Blocking (failed verification)`. | `stop` with the reason "acceptance review failed" and the same details. |
| Product decision needed, with no known failure | `stop` with the question and refer the user to `to-spec`. | Same. |
| Verification explicitly `passed`, no unmet criteria or blocking findings | `pass`. | `pass` and mark every acceptance criterion met. |

Treat any other result as malformed. Commit the ticket files. A `stop` ends the loop unless you can finish a pending handoff as described below.

If the tracker rejects only how you packaged the input and nothing was written, correct it and retry once if the ticket and target are unchanged. For problems with the child's report, use [Recovery](#recovery). Otherwise report the error and stop. Never edit ticket files to work around an error.

## Commits

- Implementers start source commit subjects with the ticket number, such as `03: add order totals`.
- Commit tracker changes as `issue-tracker` directs.
- Require a clean checkout before every dispatch.

## Recovery

If the implementer committed the work and only the report format or criterion IDs are wrong, use `stop` with `stop_kind: handoff-pending`. Send the dispatch hashes, full `HEAD` SHA, and original verification and notes. In the reason, record the original claims and what couldn't be submitted. Don't pin a handoff without usable evidence and identifiable claims.

Commit the stop, then ask the original implementer at most once to correct only the report. If the child is unavailable or the correction fails, report the pending handoff to the user. Given a valid correction, confirm the claims and evidence haven't changed, then `hand-off` directly from the stopped state and send it to independent review. Don't `resume` or `claim` to deliver the report. On restart, use the ticket's recorded stop and candidate to check a correction the user supplies.

For other missing or malformed replies, contradictory or missing evidence, or an unclean checkout, use one investigator and `stop` with the problem and the recommended action. Don't retry work or discard changes. If the checkout had other changes, leave the stop uncommitted too. The investigator can't supply verification or review.

## Stopping

Stop the loop when:

- the current acceptance ticket is `done`,
- a ticket remains `stopped`,
- the user keeps a not-blocking finding before an acceptance review,
- the user has no evidence for a manual ticket,
- nothing matches in step 1, or
- setup fails.

If the feature is accepted with unresolved not-blocking findings, ask the user whether to ignore each one before you report, as in pick step 3.

Fetch the final compact feature status and use it to build this report:

```markdown
Ralph stopped: <feature accepted | ticket stopped | findings need a decision | waiting on the user | setup failed>.

| Ticket | Kind | Phase | Status | Attempt / limit | Notes |
| --- | --- | --- | --- | --- | --- |
| 01 | implementation | review | done | 1 / 3 | |
| 02 | manual | | open | | waiting on the user |
| 03 | acceptance | | open | | blocked by 02 |

<What the user needs to do before the next run.>
```

Next steps for the user. Put the matching ones in the report, and never take them yourself:

- A pending handoff: provide a corrected report for the recorded commit. Use `resume` only if you intend to abandon that handoff and start another attempt.
- A product question: clarify it with `to-spec`, then run `to-tickets`.
- A `precondition failed` stop: rework the ticket with `to-tickets`.
- A failed acceptance review: run `to-tickets` for repair tickets.
- An implementation ticket that needs more work at the attempt limit: use `reset-attempts` or rework it with `to-tickets`.
- Other stops: resolve the cause, then `resume`.
- Not-blocking findings the user kept: run `to-tickets` to make tickets for them. List each finding in the report.
- An open manual ticket: run the checks and tell `ralph` the result.

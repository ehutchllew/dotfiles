---
name: ralph
description: Run a resumable Ralph loop from a feature's spec and tickets. Dispatch fresh implementers and reviewers, then a final feature reviewer. Use only when the user asks to run a Ralph loop on a spec or feature.
---

# Ralph

You are the **orchestrator**. Read the current files, pick the next ticket and role, construct a prompt with file paths, and dispatch a fresh subagent. Then reread the saved result and repeat. You never write or review source code, repair a subagent's work, audit code diffs, or run verification commands yourself.

Load `issue-tracker` and request the ticket scheduling, execution, review, persistence, and human hand-back rules. It owns metadata, immutable baselines, lifecycle transitions, and attempt budgets. Follow that contract rather than inventing additional transitions.

Ralph has no authoritative in-context execution state. The spec, tickets, and git must explain every next action after a restart. Do not use a loop-start SHA, session-local retry counters, or an earlier subagent report as the source of truth.

## Setup

1. Ask for the spec and ticket paths if the user did not provide them. Confirm the paths exist and read the spec, every ticket's title and metadata, and the current acceptance ticket. Read a selected ticket's full comments before dispatch.
2. Require `git status --short` to be empty before a fresh handoff. A dirty checkout needs deliberate recovery; never discard changes. On the default branch, ask the user for a feature branch or worktree rather than switching yourself.
3. Find the current final acceptance ticket using the tracker contract. Require its `Spec revision:` to match the spec's `Revision:`. If the gate or revision metadata is missing or mismatched, ask the user to reconcile with `to-spec` or `to-tickets`. Do not create a gate or guess missing execution metadata yourself.
4. If the current acceptance ticket has no `Base commit:`, record the feature base before any implementation begins and persist it. Copy an earlier acceptance ticket's recorded feature base when one exists. If any implementation has already started and no feature base was recorded, stop for an explicit recovery baseline.
5. Read [references/PROMPTS.md](references/PROMPTS.md). If the user supplies a different attempt limit, persist that approved limit on the affected unfinished tickets before dispatch. Do not reset attempt counts.

Tell the user which work is runnable and the stored attempt limits. An existing `in-progress` or `needs-review` ticket is valid resumable work; do not require a ready `open` ticket.

## The loop

Dispatch only one immediate child at a time, in the foreground. A reviewer may use separate Standards and Spec subagents when supported.

### 1. Pick the ticket and role

Reread the current spec and ticket state before each dispatch. Stop for reconciliation if the spec and current acceptance revisions no longer match. Consider only unblocked tickets with `Mode: afk`, excluding historical acceptance tickets. Take the first match, lowest ticket number first:

1. A `needs-review` ticket goes to a **reviewer**, keeping its current attempt.
2. An `in-progress` implementation ticket goes to an **implementer**.
3. A ready implementation ticket goes to an **implementer**.
4. The current acceptance ticket goes directly to a **reviewer** once every implementation ticket is `done`. Follow the tracker contract to start or resume its attempt and persist `needs-review` before dispatch.

Check the selected ticket's metadata and comments against the shared contract. A new attempt needs remaining budget; an interrupted attempt or pending review can finish at the limit. If a new attempt would exceed the limit, mark the ticket `hitl`, persist the reason, and consider other work.

If the current acceptance ticket is already `done` but unfinished implementation work exists, stop for a revised acceptance ticket. If its blockers omit feature work, stop for reconciliation. Report missing fields, unknown blockers, or dependency cycles when they prevent selection; do not silently repair them or treat them as completion. Ticket structure and granularity are validated by `to-tickets`, not by a second design review here.

### 2. Start a fresh subagent

Require a clean checkout, then fill in the matching prompt from `references/PROMPTS.md`. Pass the spec and selected ticket paths, the feature's ticket directory or explicit paths for a reviewer, and a glossary path when present for an implementer. Do not copy document contents or calculate a review base. Each agent reads the current files and the reviewer reads `Base commit:` from its ticket.

Implementers need editing, shell, and commit access. Reviewers need reading, shell access for independent verification, permission to save their ticket edits, and nested subagents when supported. Reviewers must not edit source or commit. Name the child after its role and ticket, such as `implement 03` or `review 03`.

### 3. Check the result

Reread the selected ticket rather than trusting the report. Check only the handoff contract, not code quality or commit contents:

- After an implementer, require a clean checkout and either `needs-review` with recorded verification or `in-progress` with an early-stop reason. An implementer cannot finish its own review.
- After a reviewer, require `done` or `in-progress` with the matching attempt's review record, or `needs-review` with a blocked-review record. The only permitted outstanding local change is the selected ticket. Persist the reviewer's saved edits yourself.
- On a review send-back at the attempt limit, set `Mode: hitl` and record the exhausted budget. On any final acceptance send-back, set `Mode: hitl` and request an approved repair plan. Persist these updates with the review result.
- Before another dispatch, require `git status --short` to be empty. Intentionally ignored tickets are saved through the tracker without forcing them into git.

Independent verification belongs to the reviewer. Do not rerun tests, count commits, reconstruct commit ranges, or inspect source diffs to validate the report.

### 4. Handle failures

An implementation send-back with remaining budget is ordinary work for the next fresh implementer. An exhausted or human-owned ticket stays blocked while unrelated runnable work may continue.

On an early stop, blocked review, denied command, missing baseline, invalid handoff, or unresolved scope decision, stop and report the saved state. Do not automatically redispatch the same blocked work. An interrupted invocation can resume the recorded attempt after recovery; it does not receive a new budget.

If a child returns no usable report, reread the files. A clean, complete handoff is enough to continue. Otherwise stop for recovery. Never repair code, discard changes, reset counters, change scope, or infer success from a report alone.

Pause while the user revises the spec or breakdown. Resume only from their persisted, approved changes and the tracker human hand-back rules.

## Stopping

Stop when:

- the current final acceptance ticket and all implementation tickets are `done`,
- no work can be selected because it needs a human, has exhausted its budget, or is blocked, or
- a handoff or recovery problem requires stopping.

If final acceptance already passed, report it without rerunning the review. Do not confuse completed implementation tickets with a passed feature review.

Then report to the user with this template:

```markdown
Ralph stopped: <feature accepted | waiting on human work | blocked tickets | recovery needed>.

| Ticket | Kind | Status | Mode | Attempt / limit | Notes |
| --- | --- | --- | --- | --- | --- |
| 01 | implementation | done | afk | 1 / 2 | |
| 02 | implementation | in-progress | hitl | 2 / 2 | <reason> |
| 03 | acceptance | open | afk | 0 / 2 | blocked by 02 |

<When work remains: the saved blocker and the explicit action needed to resume.>
```

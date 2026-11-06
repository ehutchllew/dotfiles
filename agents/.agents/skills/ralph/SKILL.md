---
name: ralph
description: Run a resumable Ralph loop from a feature's spec and tickets. Dispatch fresh implementers and reviewers, then a final feature reviewer. Use only when the user asks to run a Ralph loop on a spec or feature.
---

# Ralph

You are the **orchestrator**. Read the feature spec and its associated tickets, pick the next ticket and role, construct a prompt with file paths, and dispatch a fresh subagent. Then reread the saved result and repeat. You never write or review source code, repair a subagent's work, audit code diffs, or run verification commands yourself.

The spec files and their associated tickets are contracts. Adhere to their conventions and rules. It owns metadata, immutable baselines, lifecycle transitions, and attempt budgets. Follow that contract rather than inventing additional transitions.

Ralph has no authoritative in-context execution state. The spec, tickets, and git must explain every next action after a restart. Do not use a loop-start SHA, session-local retry counters, or an earlier subagent report as the source of truth.

## Setup

1. Load the `issue-tracker` skill.
2. Request the spec and tickets from the `issue-tracker`. If the user did not provide the feature name to work on, ask the user for it, then request them. Verify with the user that you've requested the correct files.
3. Read the spec and every ticket's title and metadata. Read a selected ticket's full comments before dispatch.
4. Require `git status --short` to be empty before a fresh handoff. A dirty checkout needs deliberate recovery; never discard changes. On the default branch, ask the user for a feature branch or worktree rather than switching yourself.
5. Find the current final final-review ticket using the tracker contract. Require its `Spec revision:` to match the spec's `Revision:`. If the gate or revision metadata is missing or mismatched, ask the user to reconcile with `to-spec` or `to-tickets`. Do not create a gate or guess missing execution metadata yourself.
6. If the current final-review ticket has no `Base commit:`, record the feature base before any implementation begins and persist it. Copy an earlier final-review ticket's recorded feature base when one exists. If any implementation has already started and no feature base was recorded, stop for an explicit recovery baseline.

Tell the user which work is runnable and the stored attempt limits. An existing `in-progress` or `needs-review` ticket is valid resumable work; do not require a ready `open` ticket.

## The loop

Dispatch only one immediate child at a time, in the foreground. A reviewer may use separate Standards and Spec subagents when supported.

### 1. Pick the ticket and role

Reread the current spec and ticket state before each dispatch. Stop for reconciliation if the spec's `Revision` and the latest final-review ticket's `Spec revision` do not match. Consider only unblocked tickets with `Mode: afk`, excluding historical final-review tickets. Take the first match, lowest ticket number first:

1. A `needs-review` ticket goes to a **reviewer** subagent.
2. An `open` or `in-progress` implementation ticket goes to an **implementer** subagent.
3. The current/latest `final-review` ticket goes directly to a **reviewer** subagent once every implementation ticket is `done`.

Check the selected ticket's metadata and comments against the shared contract. A new attempt needs remaining budget; an interrupted attempt or pending review can finish at the limit. If a new attempt would exceed the limit, mark the ticket `hitl`, persist the reason, and consider other work.

If the current final-review ticket is already `done` but unfinished implementation work exists, stop for a revised final-review ticket. If its blockers omit feature work, stop for reconciliation. Report missing fields, unknown blockers, or dependency cycles when they prevent selection; do not silently repair them or treat them as completion. 

### 2. Start a fresh subagent

Require a clean checkout.

Implementers need editing, shell, and commit access. Reviewers need reading, shell access for independent verification, permission to save their ticket edits, and nested subagents when supported. Reviewers must not edit source or commit. Name the child after its role, ticket and attempt, such as `implement 03 1` or `review 03 1`.

Each type of subagent receives a specific prompt:
- For **reviewer** subagents: Read and fill out the [references/REVIEWER_PROMPT.md](reviewer prompt), then hydrate the subagent with it.
- For **implementer** subagents: Read and fill out the [references/IMPLEMENTER_PROMPT.md](implementer prompt), then hydrate the subagent with it.

### 3. Check the result

Read the summary and ticket returned by the subagent. Check only the handoff contract, not code quality or commit contents:

- After an implementer, require a clean checkout and either `needs-review` with recorded verification or `in-progress` with an early-stop reason. An implementer cannot finish its own review.
- After a reviewer, require `done` or `in-progress` with the matching attempt's review record, or `needs-review` with a blocked-review record. 
- On a review send-back at the attempt limit, set `Mode: hitl` and record the exhausted budget. On any final final-review send-back, set `Mode: hitl` and request an approved repair plan. Persist these updates with the review result.
- Before another dispatch, require `git status --short` to be empty. Intentionally ignored tickets are saved through the tracker without forcing them into git.

Independent verification belongs to the reviewer. Do not rerun tests, count commits, reconstruct commit ranges, or inspect source diffs to validate the report.

### 4. Handle failures

An implementation send-back with remaining budget is ordinary work for the next fresh implementer. An exhausted or human-owned ticket stays blocked while unrelated runnable work may continue.

On an early stop, blocked review, denied command, missing baseline, invalid handoff, or unresolved scope decision, stop and report the saved state. Do not automatically redispatch the same blocked work. An interrupted invocation can resume the recorded attempt after recovery; it does not receive a new budget.

If a child returns no summary and ticket, reread the ticket you passed it. If the ticket was updated successfully continue, otherwise kick off a **reviewer** subagent to investigate.

Pause while the user revises spec or scope changes. Resume once the user approves or signals for you to continue.

## Stopping

Stop when:

- the current final final-review ticket and all implementation tickets are `done`,
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
| 03 | final-review | open | afk | 0 / 2 | blocked by 02 |

<When work remains: the saved blocker and the explicit action needed to resume.>
```

---
name: issue-tracker
description: The shared ticket format, lifecycle, review baselines, and storage operations. Load whenever a skill needs to read or write issues or specs.
---

# Issue Tracker

## Per-repo override

If the repo has `docs/agents/issue-tracker.md`, use it for storage locations and operations instead of the default storage instructions below. The ticket contract and lifecycle in this skill still apply. If the override cannot preserve them, stop and explain the incompatibility rather than translating statuses or dropping metadata silently.

## Default storage: local markdown

Issues and specs live as markdown files in `.scratch/` at the repo root. Whether `.scratch/` is committed or gitignored is up to the user; don't change it.

### Layout

- One feature per directory: `.scratch/<feature-slug>/`
- The spec is `.scratch/<feature-slug>/spec.md`
- Each ticket is its own file at `.scratch/<feature-slug>/issues/<NN>-<slug>.md`, numbered from `01`. Never combine tickets into one file.

## Spec revisions

Each spec has a plain `Revision: 1` line below its title. `to-spec` increments it when the user approves changes to requirements, implementation decisions, or verification, and records the changes in the spec's history. Editorial corrections alone do not need a new revision.

The current acceptance ticket has a `Spec revision:` line identifying the revision its approved breakdown and final review cover. `to-tickets` sets this only after the user approves reconciliation. If it differs from the spec's revision, Ralph must pause for reconciliation, even if an earlier final review passed. A changed spec is not approved execution scope merely because its file exists.

## Ticket contract

```markdown
# <NN>: <title>

Status: open
Mode: afk
Kind: implementation
Attempt: 0
Attempt limit: 2
Blocked by: 01, 02
Covers: US 4, US 7

## What to build

<the end-to-end behaviour this ticket makes work, from the user's point of view>

## Acceptance criteria

- [ ] <a criterion a test or command can check>

## Verification

<the tests or commands that prove each criterion, using commands from the spec's Testing decisions>

## Out of scope

- <nearby work this ticket must not touch>

## Comments
```

- `Status:` is one of `open`, `in-progress`, `needs-review`, `done`.
- `Mode:` is `afk` when an agent can finish the ticket alone, and `hitl` when it needs a person, such as a manual QA pass, a third-party dashboard, or a product decision.
- `Kind:` is `implementation` or `acceptance`. An acceptance ticket records the final review of the whole feature, not another implementation task.
- `Spec revision:` is required on acceptance tickets and matches the approved spec revision they verify. It is not required on implementation tickets, whose completed history remains intact across revisions.
- `Attempt:` is a nonnegative integer identifying the current or most recently completed attempt. New tickets start at `0`.
- `Attempt limit:` is a positive integer limiting complete attempts, normally `2`. A restart does not reset either field. Only a user-approved change may increase the limit.
- `Blocked by:` lists ticket numbers. Leave the line out when nothing blocks the ticket. A ticket is unblocked when every ticket it lists is `done`.
- `Covers:` lists the spec's user stories this ticket contributes to or verifies. It does not claim that a single slice implements the whole story. Prefactoring tickets can contribute through the implementation tickets they unblock.
- `Base commit:` is added during execution, as described below. Do not publish a placeholder SHA.
- Write the metadata lines plain, without bold, so a script can match `^Status: `.
- Comments and conversation history go under `## Comments`, newest last.

### Review baselines

- An implementation ticket's `Base commit:` is the full SHA returned by `git rev-parse HEAD` before its first implementation. The implementer records it when claiming the ticket, before changing source files.
- The current acceptance ticket's `Base commit:` is the feature's starting commit. Ralph records it before dispatching the first implementation agent. Never initialize it when final review begins.
- Keep both kinds of base unchanged across repairs, reviews, and restarts. A reviewer reads the base from its ticket; Ralph does not reconstruct it from commit subjects or pass it in prompts.
- A baseline is the start of a diff, not a list of changes owned by a ticket. Other tickets may have landed since that base. Ticket reviews must keep findings scoped to their ticket.
- If implementation has already started and a required base is missing, ask for an explicit recovery baseline. Never substitute the current `HEAD`. If an earlier acceptance ticket has the recorded feature base, copy that base to a new acceptance ticket.
- Scope changes do not reset a baseline. New follow-up implementation tickets get their own bases when work starts. A replacement acceptance ticket retains the original feature base.

### Final acceptance

The highest-numbered `Kind: acceptance` ticket is the current final acceptance ticket. Earlier completed acceptance tickets remain as history.

The current acceptance ticket must be blocked by every implementation ticket in the feature, including human work, and cover every current user story. Its criteria require the complete feature to work together, all blocking findings to be resolved, and the feature-level verification to pass. Its reviewer reads the full spec and all associated tickets. Ticket checkboxes alone are not evidence of feature acceptance.

Once its blockers are done, Ralph sends an `afk` acceptance ticket directly to a reviewer. No implementer or baseline-finding agent is needed. A failed final review needs a human-approved repair plan, not an implementer assigned the whole feature. Keep the findings in the acceptance ticket, set `Mode: hitl`, and use `to-spec` or `to-tickets` to agree the follow-up work before handing it back.

### Attempts

An implementation attempt includes implementation and its review. Two attempts allow an initial implementation and review, then one repair and review. For an acceptance ticket, an attempt is one final feature review after its blockers are done.

- Start attempt `1` when first claiming a ticket. An implementer also sets `Status: in-progress` immediately.
- Record the start in a new comment beginning `Attempt N: started` and persist it before doing the work.
- A reviewer finishes an attempt with a comment beginning `Attempt N: review passed` or `Attempt N: review sent back`. Include verification results and any findings below that line.
- After a send-back, increment `Attempt:` only when starting the next attempt. An interrupted or early-stopped implementation with no completed review resumes the same attempt. An interrupted review resumes the same attempt too.
- If a review cannot finish, record `Attempt N: review blocked`, explain why, and keep `Status: needs-review`. A blocked review does not consume another attempt.
- Before starting another attempt, compare `Attempt:` with `Attempt limit:`. When the budget is exhausted, Ralph sets `Mode: hitl` and records why. Never reject a pending review merely because it is the last allowed attempt.
- A `needs-review` ticket must have a started attempt with no completed review for that attempt. Repairs after a completed send-back require a new attempt, whether a human or an agent makes them. Never requeue a completed attempt to bypass the limit.

### Lifecycle

| Event | Owner | Result |
| --- | --- | --- |
| Implementation starts | Implementer | Set `in-progress`, initialize the base once, and start or resume the attempt |
| Implementation and its checks finish | Implementer | Set `needs-review` and record verification evidence |
| Implementer stops early | Implementer | Remain `in-progress` and record the reason |
| Review passes, including independent verification | Reviewer | Set `done` |
| Review sends work back | Reviewer | Set `in-progress`, untick unmet criteria, and record blocking findings |
| Review cannot finish | Reviewer | Remain `needs-review` and record the blocker |
| Final acceptance starts | Ralph | Start or resume its attempt and set `needs-review` before dispatching the reviewer |
| Final acceptance fails or the attempt budget is exhausted | Ralph | Set `Mode: hitl` and record the required human action |

Implementers may change only their ticket's `Status:`, initial `Base commit:`, `Attempt:`, acceptance-criteria checkboxes, and append-only comments. They cannot mark their own ticket `done`. Reviewers may change only their ticket's `Status:`, checkboxes, and append-only comments. They never edit source code or commit. Neither role edits the spec or another ticket.

Ralph may persist ticket metadata and comments as this contract requires, but never writes or reviews source code. It changes attempt limits only when the user explicitly approves the new value.

### Persistence and human hand-back

Persist every checkpoint before the next agent starts. For tracked local tickets, implementers commit their initialization before editing source, then commit code and ticket updates before handoff. Ralph commits its own ticket updates and the reviewer's saved ticket edits. Use ticket-prefixed commit subjects and include only the intended files. Reviewers save edits but never commit.

For intentionally gitignored local tickets, save updates without forcing them into git, and commit source changes normally. If a local ticket is untracked and not ignored, ask the user to commit or ignore it rather than changing their policy. Storage overrides must provide an equivalent durable update before handoff.

For human work, record the action or decision and its evidence in the ticket. Set `done` only when the work is actually complete. To return unfinished work to agents, explicitly set `Mode: afk` and leave `in-progress` for agent repairs. A human who completes repairs must start the next attempt under the same budget rules before handing off as `needs-review`. An unfinished review can remain `needs-review` on its existing attempt. An acceptance ticket sent back for repairs stays `in-progress` until Ralph starts its next final review after the approved blockers are done.

Preserve the baseline, attempt history, and findings. An exhausted budget needs an explicitly approved higher limit before another attempt. Resolve the cause of an early stop before resuming; a restart alone is not a resolution.

Pause Ralph while the spec or ticket breakdown changes. Preserve completed tickets and append-only execution history. Record which requirements changed and which follow-up tickets address them. Do not mark unfinished work `done` merely to unblock a revised plan.

## Operations

- **Publish to the issue tracker**: create a new file under `.scratch/<feature-slug>/`, creating the directory if needed. New tickets start as `Status: open`.
- **Fetch the relevant ticket**: read the file at the referenced path. The user will normally pass the path or the ticket number directly; if a number is ambiguous across features, ask.
- **List issues**: scan `.scratch/<feature-slug>/issues/` and read each file's `Status:` line.
- **List ready tickets**: tickets with `Status: open` and `Mode: afk` that are unblocked, lowest number first.
- **Persist a checkpoint**: save the update through the configured storage and follow the persistence rules above.
- **Comment**: append to the `## Comments` section.
- **Close**: follow the lifecycle above, append the evidence for completion, then set `Status: done`.

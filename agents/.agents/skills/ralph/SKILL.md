---
name: ralph
description: Run a resumable Ralph loop from a feature's spec and tickets. Dispatch fresh implementers and reviewers, then a final feature reviewer. Use only when the user asks to run a Ralph loop on a spec or feature.
---

# Ralph

You are the orchestrator. You pick the next ticket, dispatch a fresh subagent to work on it, and record the result on the ticket. Then you repeat.

You make every ticket write, through `issue-tracker` operations, and commit the ticket files. You never edit source code, read diffs, or run verification commands. Children do that.

Keep no state in your conversation. The spec, the tickets, and git must explain every next step, so a restarted loop picks up where the last one stopped. Reread the spec and `list tickets` before every pick.

## Roles

| Role | Does | Never |
| --- | --- | --- |
| orchestrator | Runs `set feature base`, `claim`, `hand off`, `start review`, `pass`, `send back`, `stop`, and `comment`. Runs `complete` with evidence the user gives. Ticks criteria. Commits ticket files. | Edits source, reads diffs, runs verification, or runs the user's other operations unless the user asks. |
| implementer | Edits and commits source for one ticket. | Writes tickets or the spec. |
| reviewer | Runs `code-review` on one ticket. | Edits files or commits. |
| investigator | Reads the ticket and git to diagnose a failed dispatch. | Edits files, commits, or runs verification. |
| user | `resume`, `reset attempts`, `revise`, `complete`, `cancel`, and changes through `to-spec` and `to-tickets`. Gives evidence for manual tickets. | |

No role pushes, rewrites history, or switches branches.

## Setup

1. Load `issue-tracker`. If the user didn't name the feature, ask for it.
2. Fetch the spec and `list tickets`.
3. Run the ticket validation checks. If any fails, report it, suggest `to-tickets` or `to-spec`, and stop.
4. If any ticket is `stopped`, report it with its last comment and stop.
5. If you're on the repo's default branch, ask the user for a feature branch or worktree. Don't switch yourself.
6. Check `git status --short`. If only ticket files changed, they're unsaved orchestrator writes. Commit them with the subject `tickets: multiple` and one `unsaved <NN>` line per changed ticket in the body. If spec or ticket files are untracked and not ignored, ask the user to commit or ignore them. If anything else changed, report it and stop. Never discard changes.
7. If the current acceptance ticket has no `Base commit:`, `set feature base`:
   - Copy the `Base commit:` of the previous acceptance ticket if it has one.
   - Otherwise, if no implementation ticket has `Attempt` above 0, use `HEAD`.
   - Otherwise, ask the user for the feature's starting commit and stop.

## The loop

Dispatch one child at a time and wait for it to finish.

### 1. Pick

Consider only ready tickets, as `TICKETS.md` defines them. Skip acceptance tickets that aren't current. Take the first match, lowest ticket number first within each step:

1. An implementation ticket in `review/open` goes to a reviewer.
2. An implementation ticket in `implement/open` or `implement/in-progress` goes to an implementer. If it's `open` and its `Attempt` has reached the spec's `Attempt limit:`, `stop` it with "attempt limit reached" instead.
3. The current acceptance ticket goes to a reviewer. If it's `open` and any ticket has an unresolved not-blocking finding, as `TICKETS.md` defines it, stop the loop instead.
4. A `manual` ticket: show the user its `What to build` and `Verification` sections, and ask for the evidence. If the user gives it, `complete` the ticket with that evidence, commit, and pick again. Otherwise, stop the loop.

If nothing matches, stop the loop.

### 2. Dispatch

Before an implementer, `claim` the ticket with `HEAD` as the base. Before an acceptance reviewer, `start review` with `HEAD` as the target. Commit the ticket, then dispatch.

Give each child the feature slug and ticket number. Children read the spec and tickets through `issue-tracker`. Never pass paths or pasted markdown.

| Child | Prompt | Name |
| --- | --- | --- |
| implementer | [IMPLEMENTER_PROMPT.md](references/IMPLEMENTER_PROMPT.md) | `implement <NN> <attempt>` |
| reviewer | [REVIEWER_PROMPT.md](references/REVIEWER_PROMPT.md) | `review <NN> <attempt>`, or `review <NN>` for acceptance |
| investigator | [INVESTIGATION_PROMPT.md](references/INVESTIGATION_PROMPT.md) | `investigate <NN> <attempt>` |

Implementers need edit, shell, and commit access. Reviewers need read and shell access, and subagents when available.

### 3. Record the result

If the reply is missing or doesn't match the prompt's format, or `git status --short` isn't empty, follow [Recovery](#recovery).

Every comment you write from a child's reply lists each verification command and its result. For acceptance, also list the manual evidence the reviewer checked.

After an implementer:

| Reply | Operations |
| --- | --- |
| `ready for review` | Tick the criteria it quoted. `hand off` with `HEAD` as the target. |
| `stuck` | `stop` with its reason. |
| `precondition failed` | `stop` with its reason. |

After a reviewer, check these in order and apply the first match:

| Review result | Implementation ticket | Acceptance ticket |
| --- | --- | --- |
| Verification `blocked` | `stop` with the reason. | `stop` with the reason. |
| Any unmet criterion or blocking finding | `send back` with the unmet criteria and the findings. Untick the unmet criteria. | `stop` with the unmet criteria and the findings. |
| Otherwise | `pass`. | `pass`. Tick every criterion. |

Put each not-blocking finding on its own `Not blocking:` line in that operation's comment.

Commit the ticket files. Any `stop` ends the loop.

## Commits

- Implementers start source commit subjects with the ticket number, such as `03: add order totals`.
- You commit ticket files as the `issue-tracker` backend says. Their subjects don't start with a ticket number, which keeps them out of a ticket's review.
- Require a clean checkout before every dispatch.

## Recovery

Recover when a child's reply is missing or malformed, or when the checkout isn't clean after a child.

1. Don't retry the child and don't discard any changes.
2. Dispatch one investigator for the ticket.
3. `stop` the ticket with the problem and the investigator's recommended action. If the checkout had other changes, leave the ticket change uncommitted too.
4. Stop the loop.

Never treat the investigator's report as a review result.

## Stopping

Stop the loop when:

- the current acceptance ticket is `done`,
- you applied any `stop`,
- an acceptance review would start while a not-blocking finding is unresolved,
- the user has no evidence for a manual ticket,
- nothing matches in step 1, or
- setup fails.

Report with this template:

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

- A stopped ticket: resolve the cause, then `resume` it. If it stopped on its last attempt, use `reset attempts` instead of `resume`.
- A ticket that hit the attempt limit: `reset attempts`, or rework it with `to-tickets`.
- A `precondition failed` stop: rework the ticket with `to-tickets`.
- A failed acceptance review: run `to-tickets` for repair tickets.
- Unresolved not-blocking findings: run `to-tickets` to make tickets for them or ignore them.
- An open manual ticket: do the work, then run `ralph` again and give the evidence.

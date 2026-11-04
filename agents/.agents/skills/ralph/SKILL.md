---
name: ralph
description: Run a Ralph loop over a feature's tickets. One ticket at a time goes to a fresh implementer subagent, then to a fresh reviewer subagent, and the orchestrator checks every result before the next step. Use only when the user asks to run a Ralph loop on a spec or feature.
---

# Ralph

You are the **orchestrator**. You pick the next ticket, start a fresh subagent to work on it, check what changed, and repeat. You never write or review code yourself, and you never fix a subagent's work. State lives in the ticket files and in git, not in your context.

Load the `issue-tracker` skill for the ticket format, statuses, and the definition of a ready ticket.

## Setup

1. **Find the spec.** The user tells you where the spec and tickets are. If they didn't, ask. Don't search for them.
2. **Read** the spec in full and every ticket's title and metadata lines (`Status:`, `Mode:`, `Blocked by:`). Read the rest of a ticket only when the preflight below says to.
3. **Preflight.** Stop and report to the user if any of these fail:
   - The current branch is not the default branch (`main`, `master`, or whatever `origin/HEAD` points to). Suggest a new branch or worktree.
   - `git status --short` is empty.
   - The spec and every ticket file are tracked by git (`git ls-files --error-unmatch <path>`). The checks below diff ticket files, which only works for tracked files.
   - Every ticket has `Status:` and `Mode:` lines, every `Blocked by:` number exists, and the blocking edges have no cycle.
   - Every `afk` ticket that isn't `done` holds one deliverable: one behaviour that can be checked end to end on its own. Read its acceptance criteria to judge. If a ticket looks like more than one deliverable, ask the user to split it. Don't split it yourself.
   - The spec's Testing decisions or the tickets' `## Verification` sections name the commands that prove the work. If neither does, ask the user for them.
   - At least one ticket is ready (the `issue-tracker` definition).
4. **Record** `git rev-parse HEAD` as the **loop start**.
5. **Set the limit.** The iteration limit is 3 times the number of `afk` tickets not yet `done`, unless the user gave one.
6. **Read** [references/PROMPTS.md](references/PROMPTS.md). It holds the two subagent prompts and the report format they reply with.

Tell the user, in two or three lines, which tickets you expect to run and the iteration limit. Then start.

## The loop

Run one iteration at a time. Never run two subagents at once.

### 1. Pick the ticket and role

Take the first match, lowest ticket number first:

1. A `needs-review` ticket goes to a **reviewer**.
2. An `in-progress` ticket with `Mode: afk` goes to an **implementer**.
3. A ready ticket goes to an **implementer**.

If nothing matches, stop (see [Stopping](#stopping)). This order finishes a ticket before starting the next, so each ticket's commits stay together in history.

### 2. Start a fresh subagent

Record `git rev-parse HEAD` as the **iteration start**.

Fill in the matching prompt from `references/PROMPTS.md` and start a new subagent in the foreground. Implementers need to edit files, run shell commands, and commit. Reviewers need to read files, run shell commands, edit the ticket file, and start their own subagents if possible. Name the subagent after the role and ticket, for example `implement 03` or `review 03`.

Pass paths, not file contents. The subagent reads the current files itself.

For a reviewer, you also pass the **ticket base**: the parent of the oldest commit since the loop started whose subject begins with the ticket number (`git log --reverse --format=%H --grep='^<NN>: ' <loop start>..HEAD`, then that commit's parent).

### 3. Check the result

Don't trust the report. Check the repository.

After an **implementer**:

- At least one new commit since the iteration start, and `git status --short` is empty.
- `HEAD` descends from the iteration start (`git merge-base --is-ancestor <iteration start> HEAD`), so no history was rewritten.
- The ticket moved from `open` or `in-progress` to `needs-review`. Otherwise it kept its status with a new comment explaining why the implementer stopped, which step 4 handles.
- `git diff <iteration start> -- <ticket>` changes only the `Status:` line, acceptance-criteria checkboxes, and new lines under `## Comments`.
- No other ticket file and not the spec changed (`git diff --stat <iteration start>`).
- If the ticket is now `needs-review`, run its `## Verification` commands yourself, or the spec's Testing decisions commands when the ticket has none. Read only the tail of any failing output.

After a **reviewer**:

- `HEAD` equals the iteration start. Reviewers don't commit.
- `git status --short` shows only the ticket file.
- The ticket moved from `needs-review` to `done`, or to `in-progress` with the findings in a new comment.
- `git diff -- <ticket>` changes only the `Status:` line, acceptance-criteria checkboxes, and new lines under `## Comments`.

When the reviewer's checks pass, commit the ticket file yourself: `<NN>: review passed` or `<NN>: review sent back`.

### 4. Handle failures

You decide whether a failure is worth retrying. Retry with a fresh subagent of the role that failed, at most once for the same failure on the same ticket.

Usually worth retrying:

- The subagent returned no report, or the report doesn't match the format.
- Your verification run failed after the implementer reported a pass. Add a comment to the ticket with the failing command and the end of its output, set `Status: in-progress`, commit as `<NN>: verification failed`, and send the ticket to a new implementer.
- The implementer stopped early and its comment points to something a fresh attempt could get past, such as running out of context or a flaky test.

Not worth retrying. Stop and report:

- The implementer found that the ticket holds more than one deliverable. The user needs to split it.
- The implementer stopped early because the ticket contradicts the spec or needs a decision.
- The subagent changed something it may not change: the spec, another ticket, a protected part of its ticket, or history.
- A reviewer committed.
- Uncommitted changes are left after an implementer. Don't discard them.
- The harness denied a command the work needs.
- The same failure happened again after a retry.

### 5. Hand repeated send-backs to a human

When a reviewer sends the same ticket back for the second time, set its `Mode: hitl`, add a comment linking the two review comments, and commit as `<NN>: needs a human`. Tickets blocked by it stay blocked. The loop carries on with any other ready ticket.

## Stopping

Stop when:

- no ticket can be picked in step 1, because every ticket is `done`, is `hitl`, or is blocked (directly or through other tickets) by a `hitl` ticket that isn't `done`,
- the iteration limit is reached, or
- a failure isn't worth retrying.

Then report to the user with this template:

```markdown
Ralph stopped: <all tickets done | waiting on hitl tickets NN, NN | iteration limit reached | failure on ticket NN>.

Iterations: <n> of <limit>. Commits: `<loop start>..<HEAD>`.

| Ticket | Status | Mode | Notes |
| --- | --- | --- | --- |
| 01 | done | afk | |
| 02 | in-progress | hitl | sent back twice: <one-line reason> |

<Only on a failure: what you expected, what the check found, and the git commands to inspect it.>
```

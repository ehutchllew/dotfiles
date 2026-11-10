# Tickets

## Template

Text in `<...>` is a placeholder. Include only the metadata lines that apply to the ticket's kind. The Fields table lists them, with their allowed and starting values.

```markdown
# <NN>: <title>

Kind: <kind>
Phase: <phase>
Status: <status>
Attempt: <attempt>
Blocked by: <ticket numbers>
Covers: <story IDs>
Spec revision: <spec revision>

## What to build

<the behavior this ticket makes work, from the user's point of view>

## Acceptance criteria

- [ ] <a criterion that can be checked>

## Verification

<how to check each criterion>

## Out of scope

- <nearby work this ticket must not touch>

## Comments
```

## Fields

Write metadata lines plain, one per line, directly under the title, so `list tickets` can read them. Operations add `Base commit:` and `Target commit:` after the last metadata line.

| Field | Kinds | Values | Set by |
| --- | --- | --- | --- |
| `Kind:` | all | `implementation`, `acceptance`, `manual` | `create` |
| `Phase:` | implementation | `implement`, `review`. Starts at `implement`. | state operations |
| `Status:` | all | `open`, `in-progress`, `stopped`, `done`, `cancelled`. Starts at `open`. | state operations |
| `Attempt:` | implementation | Integer. Starts at `0`. | `claim`, `revise`, `reset attempts` |
| `Blocked by:` | all | Ticket numbers, such as `01, 02`. Omit the line when empty. | `create`, `edit` |
| `Covers:` | all | Story IDs, such as `US 1, US 2`, or `none` | `create`, `edit` |
| `Spec revision:` | acceptance | The spec revision this ticket checks | `create`, `edit` |
| `Base commit:` | implementation, acceptance | Full commit SHA | `claim`, `set feature base` |
| `Target commit:` | implementation, acceptance | Full commit SHA | `hand off`, `start review` |

What the fields mean:

- `Kind:`
  - `implementation` is code work.
  - `acceptance` is the final review of the whole feature.
  - `manual` is work only the user can do, such as a manual QA pass, a third-party dashboard, or a product decision.
- `Covers:` lists the stories a ticket contributes to. A ticket doesn't have to complete a story alone. Prefactor and refactor tickets use `none`.
- `Base commit:` is where the review range starts. It never changes once set.
  - Implementation: the commit before the ticket's first attempt.
  - Acceptance: the commit before the feature's first attempt.
- `Target commit:` is where the review range ends.
  - Implementation: the last source commit when the ticket is handed off.
  - Acceptance: the commit under review when its review starts.

## Sections

- What to build:
  - Implementation: the behavior this ticket makes work.
  - Acceptance: "Review the complete feature against the spec."
  - Manual: the action the user takes.
- Acceptance criteria: each criterion can be checked on its own.
- Verification:
  - Implementation and acceptance: commands from the spec's Testing decisions.
  - Manual: the steps the user follows and the evidence to record.
- Out of scope: nearby work this ticket must not touch, usually work that belongs to another ticket.
- Comments: history, newest last.
  - Append only. Never edit or delete a comment.
  - Start each comment with the operation's name. On implementation tickets, add the attempt, such as `send back (attempt 2): <findings>`.
  - Put each not-blocking review finding on its own line in the operation's comment, starting with `Not blocking:`.
  - A not-blocking finding is resolved once a later comment on the same ticket either names the new ticket for it or says it's ignored. Use one of these formats, and leave out `(attempt N)` on tickets without attempts:
    - New ticket: `comment (attempt N): Not blocking resolved: <finding> moved to ticket <NN>`
    - Ignored: `comment (attempt N): Not blocking resolved: <finding> ignored`

## Rules

- Never renumber a ticket. Other tickets refer to it by number.
- Don't edit a `done` or `cancelled` ticket except to add a comment. Changes to finished work go in a new ticket.
- Follow the rules for file paths, code, and commands in `SPECS.md`.

## State operations

Each operation changes one ticket. `From` and `To` show `Phase/Status` for implementation tickets and `Status` for other kinds. Operations that set a commit take the SHA as input.

| Operation | Kind | From | To | Also |
| --- | --- | --- | --- | --- |
| `claim` | implementation | implement/open | implement/in-progress | Add 1 to `Attempt`. Set `Base commit` if missing. |
| `claim` | implementation | implement/in-progress | no change | Resumes the current attempt. |
| `hand off` | implementation | implement/in-progress | review/open | Set `Target commit`. Comment with verification results. |
| `pass` | implementation | review/open | review/done | Comment. |
| `send back` | implementation | review/open | implement/open | Untick unmet criteria. Comment with findings. |
| `set feature base` | acceptance | open | no change | Set `Base commit`. Fails if already set. |
| `start review` | acceptance | open | in-progress | Set `Target commit`. |
| `start review` | acceptance | in-progress | no change | Resumes the current review. |
| `pass` | acceptance | in-progress | done | Comment. |
| `complete` | manual | open | done | Tick every criterion. Comment with the evidence. |
| `stop` | any | open or in-progress | stopped | Comment with the reason. |
| `resume` | any | stopped | open | Comment with the resolution. |
| `reset attempts` | implementation | implement/stopped | implement/open | Set `Attempt` to `0`. Comment with the reason. |
| `revise` | implementation | any phase, open, in-progress, or stopped | implement/open | Set `Attempt` to `0`. Untick every criterion. Comment with the reason. |
| `revise` | acceptance | open, in-progress, or stopped | open | Untick every criterion. Comment with the reason. |
| `cancel` | any | any except done | cancelled | Comment. Remove the ticket from every `Blocked by:` line. |
| `comment` | any | any | no change | Append under `## Comments`. |

- `stop` and `resume` never change `Phase`.
- `revise` keeps `Base commit` and the old `Target commit`. The next `hand off` or `start review` replaces `Target commit`.
- A failed acceptance review uses `stop`.

## Readiness

A ticket is ready when all of these hold:

- `Status:` is `open` or `in-progress`.
- Every ticket in `Blocked by:` is `done`.

## Current acceptance ticket

The current acceptance ticket is the highest-numbered acceptance ticket that isn't cancelled. Earlier acceptance tickets are history.

## Validation checks

A feature's tickets are valid when every check below holds.

### Each ticket

1. It has the fields its kind needs, with allowed values.
2. An implementation ticket in the `review` phase is `open`, `stopped`, or `done`.
3. An implementation ticket with `Attempt` above `0` has a `Base commit`. One in the `review` phase also has a `Target commit`.

### Blockers

4. Every ticket in a `Blocked by:` line exists and isn't cancelled.
5. No chain of blockers leads back to its starting ticket, such as `01` blocked by `02` blocked by `03` which is blocked by `01`.

6. An `in-progress` acceptance ticket has every blocker `done`.

### The whole feature

7. Every story ID in a `Covers:` line exists in the spec and isn't withdrawn.
8. Every story that isn't withdrawn is covered by at least one implementation ticket that isn't cancelled.
9. A current acceptance ticket exists. It covers every story that isn't withdrawn, and its `Spec revision:` matches the spec's `Revision:`.
10. The current acceptance ticket is blocked by every implementation and manual ticket that isn't cancelled.

## Examples

These examples assume a spec at revision `1` with one story. `US 1` asks for item counts in the order summary. Its Testing decisions name `npm test` and one manual check on a phone.

Implementation ticket:

```markdown
# 01: Show order item counts

Kind: implementation
Phase: implement
Status: open
Attempt: 0
Covers: US 1

## What to build

Show the total item quantity in the order summary.

## Acceptance criteria

- [ ] An empty order shows an item count of zero.
- [ ] The item count equals the sum of the order's item quantities.

## Verification

Run `npm test`.

## Out of scope

- Order totals.

## Comments
```

Manual ticket:

```markdown
# 02: Check item counts on a phone

Kind: manual
Status: open
Blocked by: 01
Covers: US 1

## What to build

Check the order summary on a phone.

## Acceptance criteria

- [ ] The item count is readable without zooming.

## Verification

Open the order summary on a phone. Record the device and a screenshot in a comment.

## Out of scope

- Desktop layout.

## Comments
```

Acceptance ticket:

```markdown
# 03: Accept the order summary feature

Kind: acceptance
Status: open
Blocked by: 01, 02
Covers: US 1
Spec revision: 1

## What to build

Review the complete feature against the spec.

## Acceptance criteria

- [ ] Every story in the spec works.
- [ ] Feature verification passes.

## Verification

Run `npm test`. Check the evidence on ticket 02.

## Out of scope

- Implementing repairs or changing the spec.

## Comments
```

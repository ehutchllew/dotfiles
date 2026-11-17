# Writing tickets

This guide covers ticket content and relationships.

## Kinds

- `implementation` is code work.
- `acceptance` is the final review of the whole feature.
- `manual` is work only the user can do, such as a manual QA pass, a third-party dashboard, or a product decision.

## Relationships

- List the stories a ticket contributes to. A ticket doesn't have to complete a story alone. Prefactor and refactor tickets cover no stories.
- A blocker must be done before the ticket can start.
- An acceptance ticket covers every story that isn't withdrawn, is blocked by every implementation and manual ticket that isn't cancelled, and checks the spec's current revision.

## Sections

- What to build:
  - Implementation: the behavior this ticket makes work, from the user's point of view.
  - Acceptance: "Review the complete feature against the spec."
  - Manual: the action the user takes.
- Acceptance criteria: each criterion can be checked on its own. Each is one line.
- Verification:
  - Implementation and acceptance: commands from the spec's Testing decisions.
  - Manual: the steps the user follows and the evidence to record.
- Out of scope: nearby work this ticket must not touch, usually work that belongs to another ticket.

## Rules

- Follow the spec's rules for file paths, code, and commands: no file paths or code, and commands exactly as typed. They don't apply to comments.
- Done and cancelled tickets can't be edited. Changes to finished work go in a new ticket.

## Example

The spec is at revision 1 with one story. `US 1` asks for item counts in the order summary. Its Testing decisions name `npm test` and one manual check on a phone. All three tickets cover `US 1`.

| Kind | What to build | Acceptance criteria | Verification | Blockers |
| --- | --- | --- | --- | --- |
| implementation | Show the total item quantity in the order summary. | An empty order shows zero. The count equals the sum of item quantities. | Run `npm test`. | None. |
| manual | Check item counts on a phone. | The count is readable without zooming. | Record the device and a screenshot in a comment. | The implementation ticket. |
| acceptance | Review the complete feature against the spec. | Every story works. Feature verification passes. | Run `npm test` and check the manual evidence. | Both other tickets. |

Order totals are outside the implementation ticket. Desktop layout is outside the manual ticket. The acceptance ticket excludes repairs and spec changes.

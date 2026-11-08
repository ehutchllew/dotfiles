---
name: to-tickets
description: Turn a spec into tickets, or update existing tickets after a spec revision or a failed acceptance review. Use when the user wants a spec or plan turned into tickets.
disable-model-invocation: true
metadata:
    opencode/autoinvoke: false
---

# To tickets

Break a spec into implementation tickets, manual tickets, and one acceptance ticket. When tickets already exist, update them instead of creating duplicates.

This skill reads code but changes only tickets, through `issue-tracker`. It never changes source code or the spec.

Use this skill when:

- a spec has no tickets yet,
- a spec revision changed what the tickets must cover, or
- the current acceptance ticket is `stopped` after a failed review and the user wants repair tickets, with no spec change.

## 1. Gather context

1. Load `issue-tracker` and read the spec and ticket formats.
2. `fetch` the spec. If there is none, suggest `to-spec` and stop.
3. If the spec's Testing decisions name no commands, ask the user to add them with `to-spec` and stop.
4. `list tickets`. If tickets exist, `fetch` each one and read it in full, including comments.
5. If any ticket is `in-progress`, work on this feature may still be running. Ask the user to confirm it has stopped before you change anything.
6. For a failed acceptance review, read the findings in the acceptance ticket's comments.

## 2. Explore the code

If you haven't inspected the code that the spec touches, do so. Use the terms in `GLOSSARY.md` and respect relevant ADRs.

Look for prefactoring that would make the feature tickets smaller or simpler. Plan each prefactor as its own ticket that comes before the tickets it helps.

## 3. Draft the tickets

### Implementation tickets

- Each ticket holds one deliverable: a change that can be checked on its own once it lands. If the title needs "and", or the criteria check unrelated behaviors, split the ticket.
- Acceptance criteria state what the deliverable must do. They are not a todo list.
- Most deliverables are vertical slices:
  - A slice cuts a narrow path through every layer it needs, such as schema, API, UI, and tests.
  - A finished slice can be shown working or checked on its own.
  - A slice fits in one fresh context window.
- A prefactor keeps behavior the same, changes the code's shape, and leaves tests passing.

### Wide refactors

A wide refactor is one mechanical change, such as renaming a column, that touches too many call sites to land as one passing slice. Split it into horizontal tickets:

1. Expand: add the new form beside the old. Nothing breaks.
2. Migrate: one ticket per batch of call sites, such as per package or directory. Each is blocked by the expand ticket.
3. Contract: remove the old form. It is blocked by every migrate ticket.

Each ticket must pass its own Verification. If a batch can't, split it or fix the expand ticket.

### Manual tickets

Add a manual ticket for each check in the spec's Testing decisions that only the user can do, and for any product decision or third-party setup the work depends on.

### Acceptance ticket

Add one acceptance ticket for the feature.

### All tickets

- Block a ticket only on tickets that must be done before it can start.
- Number tickets in the order the work should happen.

## 4. Update existing tickets

Skip this step when the feature has no tickets.

- `edit` unfinished tickets to match the change, and `comment` why. If an implementation ticket in the `review` phase has changed criteria, `send back` with the reason.
- Add new tickets for changes to `done` work. `comment` on each affected `done` ticket with the new ticket's number and the reason.
- `cancel` work that no longer applies, only after the user agrees.
- If the current acceptance ticket is `done`, add a new one. Otherwise, `edit` it to match the new breakdown. If it is `stopped`, `resume` it after its new blockers are published.
- If a conflict prevents a runnable plan, name it and let the user decide.

## 5. Review with the user

1. Check your plan. Each ticket must follow the rules in step 3. All the tickets together, old and new, must pass the validation checks in the ticket format. Fix any problems you find.
2. Show the breakdown as a numbered list. For each ticket, show its title, kind, blockers, covered stories, and what it delivers. Show changes to existing tickets apart from new tickets.
3. List anything you're unsure of, such as a ticket that may hold two deliverables or a check that may need the user.
4. Ask the user to approve or change the breakdown. Repeat until they approve.

## 6. Publish

Publish the approved changes through `issue-tracker`.

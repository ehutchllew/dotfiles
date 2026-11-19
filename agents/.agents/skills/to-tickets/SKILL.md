---
name: to-tickets
description: Turn a spec into tickets, or update existing tickets after a spec revision, a failed acceptance review, or a stopped implementation ticket. Use when the user wants a spec or plan turned into tickets.
disable-model-invocation: true
metadata:
    opencode/autoinvoke: false
---

# To tickets

Break a spec into implementation tickets, manual tickets, and one acceptance ticket. When tickets already exist, update them instead of creating duplicates.

This skill reads code but changes only tickets, through `issue-tracker`. It never changes source code or the spec.

Use this skill when:

- a spec has no tickets yet,
- a spec revision changed what the tickets must cover,
- the current acceptance ticket is `stopped` after a failed review and the user wants repair tickets, with no spec change,
- an implementation ticket is `stopped` and the user wants it reworked or split, or
- tickets have unresolved not-blocking review findings.

## 1. Gather context

1. Load `issue-tracker` and read [TICKET-GUIDE.md](references/TICKET-GUIDE.md).
2. Read the spec through `issue-tracker`. If there is no spec, suggest `to-spec` and stop.
3. If the spec's Testing decisions name no commands, ask the user to add them with `to-spec` and stop.
4. Fetch compact feature status. Read each existing ticket and its comments. For stopped tickets, request full history so you see every attempt.
5. If any ticket is `in-progress`, work on this feature may still be running. Ask the user to confirm it has stopped before you change anything.
6. For a failed acceptance review, read the acceptance ticket's `stop` comment. Draft repairs for findings. If it records blocked verification or a product question instead, report that to the user and stop.
7. For a stopped implementation ticket, read its comments to learn why each attempt failed. Use that to decide how to rework or split it.
8. Note every unresolved not-blocking finding reported in status, with its ticket number.

## 2. Explore the code

If you haven't inspected the code that the spec touches, do so. Use the terms in `GLOSSARY.md` and respect relevant ADRs.

Look for prefactoring that would make the feature tickets smaller or simpler. Plan each prefactor as its own ticket that comes before the tickets it helps.

## 3. Draft the tickets

### Implementation tickets

- Each ticket holds one deliverable you can check on its own without shipping incorrect behavior. Put a safety requirement in the first ticket that introduces the risk it guards against. Treat "and" in the title as a cue to check for unrelated deliverables, not an automatic split.
- Acceptance criteria state what the deliverable must do. They are not a todo list.
- Every criterion must be something the ticket's Verification can confirm. If only a person can check it, such as a visual check when the repo has no browser automation, put it in a manual ticket instead.
- Carry over agreed examples where they apply. If the spec leaves an observable requirement undecided, ask for a decision through `to-spec` rather than deciding it in a ticket.
- Most deliverables are vertical slices:
  - A slice cuts a narrow path through every layer it needs, such as schema, API, UI, and tests.
  - A finished slice can be shown working or checked on its own.
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

### Not-blocking findings

Draft one implementation ticket for each unresolved not-blocking finding, or one for a group of related findings. It covers no stories unless the finding concerns a story. The user decides in step 5 which drafts to keep.

### All tickets

- Block a ticket only on tickets that must be done before it can start.
- Blockers define work order. If new work must come before an unfinished ticket, make it a blocker of that ticket.

## 4. Update existing tickets

Skip this step when the feature has no tickets.

- For every unfinished ticket you change, `revise` it with the reason before editing its content or blockers. This applies to every kind, in any phase, including `in-progress` and `stopped` tickets.
- Add new tickets for changes to `done` work, including any manual scenarios that need to run again after a repair. `comment` on each affected `done` ticket with a note naming the new ticket and the reason.
- `cancel` work that no longer applies.
- If the current acceptance ticket is `done`, add a new one. Otherwise, after its new blockers are published, `revise` it and update its blockers, covered stories, and spec revision to match the new breakdown.
- For each not-blocking finding the user resolved in step 5, use `comment` on its original ticket to resolve it as ignored or moved to the new ticket.
- If a conflict prevents a runnable plan, name it and let the user decide.

## 5. Review with the user

1. Check your plan. Each ticket must follow the rules in step 3 and the guide. All the tickets together, old and new, must cover every story that isn't withdrawn, and the current acceptance ticket must be blocked by every implementation and manual ticket that isn't cancelled. Fix any problems you find.
2. Show the breakdown as a numbered list. For each ticket, show its title, kind, blockers, covered stories, and what it delivers. Show changes to existing tickets apart from new tickets.
3. List anything you're unsure of, such as a ticket that may hold two deliverables or a check that may need the user.
4. Show each unresolved not-blocking finding with its draft ticket, and ask the user to keep the ticket or ignore the finding.
5. Ask the user to approve or change the breakdown. Repeat until they approve.

## 6. Publish

Publish only the changes the user approved:

1. Create new tickets in the approved order, blockers first and the acceptance ticket last. Check that the assigned ticket numbers match the plan.
2. Apply the edits and operations from step 4.
3. Validate the complete feature through `issue-tracker`. If validation fails, fix the tickets you published and validate again. If an edit failed after its `revise`, retry the edit before you report, so no reopened ticket keeps its old content.
4. Commit the published changes as `issue-tracker` directs.

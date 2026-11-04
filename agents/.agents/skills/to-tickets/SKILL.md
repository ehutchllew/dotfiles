---
name: to-tickets
description: Break a spec into tickets, one deliverable each, with blocking edges, and publish them through the issue tracker. Use when the user wants a spec or plan turned into tickets.
disable-model-invocation: true
---

# To Tickets

Break a spec into **tickets**. Each ticket holds one **deliverable** and declares the tickets that **block** it.

## Process

### 1. Gather context

Load the `issue-tracker` skill and fetch the spec the user points to, including its comments. Also use whatever is already in the conversation.

Tickets need a spec. `Covers:` refers to the spec's user story IDs, and `## Verification` uses the commands in its Testing decisions. If there is no spec, suggest running `to-spec` first. If the spec's Testing decisions name no commands, ask the user for them.

### 2. Explore the codebase (optional)

If you have not already explored the codebase, do so to understand the current state of the code. Ticket titles and descriptions should use the project's domain glossary vocabulary, and respect ADRs in the area you're touching.

Look for opportunities to prefactor the code to make the implementation easier. "Make the change easy, then make the easy change."

### 3. Draft the tickets

A ticket holds exactly one deliverable: a change that can be checked on its own once it lands. Acceptance criteria check that one deliverable; they are not a todo list. If the title needs "and", or the criteria check unrelated behaviours, split the ticket.

Most deliverables are **vertical slices**, also called tracer bullets:

<vertical-slice-rules>

- Each slice cuts a narrow but complete path through every layer (schema, API, UI, tests). It is vertical, not a horizontal slice of one layer.
- A completed slice is demoable or verifiable on its own.
- Each slice fits in a single fresh context window.
- Prefactoring comes first, as its own tickets. A prefactor keeps the behaviour the same, gives the code a new shape, and leaves the tests passing.

</vertical-slice-rules>

Give each ticket its **blocking edges**: the other tickets that must complete before it can start. A ticket with no blockers can start immediately.

**Wide refactors are the exception, and their deliverables are horizontal slices.** A **wide refactor** is one mechanical change (rename a column, retype a shared symbol) whose **blast radius** fans across the whole codebase, so a single edit breaks thousands of call sites at once and no vertical slice can land green. Don't force it into a tracer bullet; sequence it as **expand-contract**. First expand: add the new form beside the old so nothing breaks. Then migrate the call sites over in batches sized by blast radius (per package, per directory), each batch its own ticket blocked by the expand. Finally contract: delete the old form once no caller remains, in a ticket blocked by every migrate batch. Every batch must pass its own Verification, which works because the old form still exists. If a batch can't, it is too big or the expand step is missing something, so rework the sequence.

Then fill in each ticket:

- **Mode**: `hitl` when the ticket needs a person, such as a manual QA pass, a third-party dashboard, or a product decision. Otherwise `afk`. Each check the spec's Testing decisions lists as needing a person belongs in a `hitl` ticket.
- **Covers**: the spec's user stories this ticket implements. Every user story should be covered by at least one ticket.
- **Verification**: the tests or commands that prove each acceptance criterion. Use only commands from the spec's Testing decisions.
- **Out of scope**: nearby work this ticket must not touch, usually work that belongs to another ticket.

### 4. Quiz the user

Present the proposed breakdown as a numbered list. For each ticket, show:

- **Title**: short descriptive name
- **Mode**: `afk` or `hitl`
- **Blocked by**: which other tickets (if any) must complete first
- **Covers**: which user stories it implements
- **What it delivers**: the change this ticket makes work

After the list, name any user stories no ticket covers.

Ask the user:

- Does the granularity feel right? (too coarse / too fine)
- Does each ticket hold one deliverable?
- Are the blocking edges correct: does each ticket only depend on tickets that genuinely gate it?
- Is each `hitl` ticket really one that needs a person, and is any `afk` ticket missing one?
- Should any tickets be merged or split further?

Iterate until the user approves the breakdown.

### 5. Publish the tickets

Publish each approved ticket in the `issue-tracker` ticket format. Publish in dependency order, blockers first, so each `Blocked by:` refers to tickets that already exist.

Don't close or change the spec or any parent issue.

If the tickets are local files, offer to commit the feature directory with the spec and tickets. Ralph only runs on committed tickets.

Avoid specific file paths or code snippets: they go stale fast. Exception: if a prototype produced a snippet that encodes a decision more precisely than prose can (state machine, reducer, schema, type shape), inline it and note briefly that it came from a prototype. Trim to the decision-rich parts, not a working demo, just the important bits.

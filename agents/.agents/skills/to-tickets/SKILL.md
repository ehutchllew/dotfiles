---
name: to-tickets
description: Turn a spec into implementation tickets and a final acceptance ticket, or reconcile existing tickets with an approved spec revision. Use when the user wants a spec or plan turned into tickets.
disable-model-invocation: true
metadata:
    opencode/autoinvoke: false
---

# To Tickets

Break a spec into implementation tickets, each with one deliverable and explicit blockers, plus a final acceptance ticket. Reconcile existing work instead of generating duplicate tickets when the spec changes.

## Process

### 1. Gather context

Load `issue-tracker` and request the spec requirements relevant to ticket breakdown, plus the ticket template and authoring, reconciliation, and final acceptance guidelines. Fetch the spec the user points to, including its revision and history. Read existing tickets and their execution metadata and comments. Also use whatever is already in the conversation.

Before changing an existing breakdown, confirm that no one is implementing or reviewing affected tickets.

Tickets need a spec. `Covers:` refers to the spec's user story IDs, and `## Verification` uses the commands in its Testing decisions. If there is no spec, suggest running `to-spec` first. If the spec's Testing decisions name no commands, ask the user for them.

### 2. Explore the codebase (optional)

If you have not already explored the codebase, do so to understand the current state of the code. Ticket titles and descriptions should use the project's domain glossary vocabulary, and respect ADRs in the area you're touching.

Look for opportunities to prefactor the code to make the implementation easier and record them as preliminary tickets; do NOT do the work yourself. "Make the change easy, then make the easy change."

### 3. Draft the tickets

An implementation ticket holds exactly one deliverable: a change that can be checked on its own once it lands. Acceptance criteria are not a todo list, they are high-level requirements for the deliverable. If the title needs "and", or the criteria check unrelated behaviours, split the ticket. Final acceptance is a review gate, not a feature-sized implementation ticket.

Most deliverables are **vertical slices**, also called tracer bullets:

<vertical-slice-rules>

- Each slice cuts a narrow but complete path through every layer (schema, API, UI, tests). It is vertical, not a horizontal slice of one layer.
- A completed slice is demoable or verifiable on its own.
- Each slice fits in a single fresh context window.
- Prefactoring comes first, as its own tickets. A prefactor keeps the behaviour the same, gives the code a new shape, and leaves the tests passing.

</vertical-slice-rules>

Give each ticket its **blocking edges**: the other tickets that must complete before it can start. A ticket with no blockers can start immediately.

**Wide refactors are the exception, and their deliverables are horizontal slices.** A **wide refactor** is one mechanical change (rename a column, retype a shared symbol) whose **blast radius** fans across a large portion of the codebase, so a single edit breaks thousands of call sites at once and no vertical slice can land green. Don't force it into a tracer bullet; sequence it as **expand-contract**. First expand: add the new form beside the old so nothing breaks. Then migrate the call sites over in batches sized by blast radius (per package, per directory), each batch its own ticket blocked by the expand. Finally contract: delete the old form once no caller remains, in a ticket blocked by every migrate batch. Every batch must pass its own Verification, which works because the old form still exists. If a batch can't, it is too big or the expand step is missing something, so rework the sequence.

Then fill in each ticket:

- **Mode**: `hitl` when the ticket needs a person, such as a manual QA pass, a third-party dashboard, or a product decision. Otherwise `afk`. Each check the spec's Testing decisions lists as needing a person belongs in a `hitl` ticket.
- **Covers**: the spec's user stories this ticket contributes to or verifies. Every current story needs implementation coverage, not just a final acceptance check.
- **Verification**: the tests or commands that prove each acceptance criterion. Use only commands from the spec's Testing decisions.
- **Out of scope**: nearby work this ticket must not touch, usually work that belongs to another ticket.

Use the tracker's initial metadata for new implementation tickets. Do not invent a `Base commit:` before work starts or reset an existing ticket's base, attempt count, limit, or review history.

### Final acceptance and revisions

Add one current `Kind: acceptance` ticket using the tracker contract. It must depend on every implementation ticket, cover all current stories, and verify feature-level behavior and integration through the spec's commands. Include human checks in explicit `hitl` work before final acceptance, or make the acceptance ticket `hitl` if the final check itself needs a person.

When reconciling existing tickets:

- Keep ticket numbers stable and assign new numbers after the highest existing number. Preserve completed tickets and create follow-up work for changed behavior.
- Propose precise changes to unfinished tickets, including blockers and unmet criteria. Preserve execution metadata and append a comment explaining approved scope changes. If a changed ticket was awaiting review, return it to `in-progress` for implementation against the revised criteria.
- Do not erase earlier findings, reset budgets, mark unfinished work `done` to remove a blocker, or retire contradictory work without the user's explicit resolution. Name any conflict that prevents a runnable plan.
- Reuse an unfinished acceptance ticket, preserving its base and attempt history. If final acceptance already passed, append a new acceptance ticket and copy the original feature base. Keep the earlier pass as history.
- After approval, set the current acceptance ticket's `Spec revision:` to the spec revision being reconciled. Add new blockers for all follow-up work. A failed acceptance ticket remains `hitl` until the user explicitly approves its repair plan and hand-back, including a higher limit if its budget is exhausted.

### 4. Quiz the user

Present the proposed breakdown as a numbered list. For each ticket, show:

- **Title**: short descriptive name
- **Kind**: implementation or final acceptance
- **Mode**: `afk` or `hitl`
- **Blocked by**: which other tickets (if any) must complete first
- **Covers**: which user stories it contributes to or verifies
- **What it delivers**: the change this ticket makes work

After the list, name any user stories no implementation ticket contributes to. Acceptance coverage alone does not count as implementation coverage. Show changes to existing tickets separately from new tickets.

Ask the user:

- Does the granularity feel right? (too coarse / too fine)
- Does each implementation ticket hold one deliverable?
- Are the blocking edges correct: does each ticket only depend on tickets that genuinely gate it?
- Is each `hitl` ticket really one that needs a person, and is any `afk` ticket missing one?
- Should any tickets be merged or split further?

Iterate until the user approves the breakdown.

### 5. Publish the tickets

Before publishing, check that every blocker exists, the dependency graph has no cycles, current story IDs are valid and covered, each implementation ticket has one deliverable, and its Verification names commands from the spec. Check that the current final acceptance ticket depends on all implementation work and names the approved spec revision. Resolve invalid structure with the user before publishing.

Publish only the approved additions and edits in the `issue-tracker` format, blockers first. Preserve existing ticket history and record the approved reconciliation with the changes.

Don't close or change the spec or any parent issue.

For local files, offer to commit the approved spec and tickets. Respect an intentional gitignore policy rather than force-adding files.

Avoid specific file paths or code snippets: they go stale fast. Exception: if a prototype produced a snippet that encodes a decision more precisely than prose can (state machine, reducer, schema, type shape), inline it and note briefly that it came from a prototype. Trim to the decision-rich parts, not a working demo, just the important bits.

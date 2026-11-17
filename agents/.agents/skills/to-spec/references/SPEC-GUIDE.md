# Writing a spec

This guide covers the spec's content and revision rules.

## Rules for the whole spec

- Don't include file paths or code. They go stale. Sample terminal output and sample data are fine.
- Name modules, but don't give their paths. Describe interfaces by their inputs and outputs.
- Write a command exactly as typed in a terminal when it names no module path, such as `go test ./...`. Otherwise describe it by the module it runs and its arguments, such as "run the `ordersummary` command with `-code SAVE10`".
- If a prototype produced a snippet that states a decision more precisely than prose, such as a state machine, reducer, schema, or type shape, include it in the matching implementation decision. Keep only the part that carries the decision, and note that it came from a prototype.

## Sections

### Problem statement and Solution

Write both from the user's point of view.

### User stories

- Write each story as "As a <actor>, I want <capability>, so that <benefit>". Stories have IDs such as `US 1` and `US 2`.
- Example: "As a mobile bank customer, I want to see the balance on my accounts, so that I can make better decisions about my spending"
- Cover every behavior the feature needs. Don't add stories nobody asked for.
- Add new stories at the end. Edit or withdraw existing stories without renumbering or reusing their IDs.

### Implementation decisions

List the decisions made. These can include:

- modules to build or change, and their interfaces
- technical clarifications from the user
- architectural decisions
- schema changes
- API contracts
- specific interactions

### Testing decisions

List the decisions made. Include:

- which modules get tests
- prior art: similar tests already in the codebase
- the exact commands that run the tests and other checks, such as typecheck and lint
- checks only the user can do, such as a manual QA pass

### Out of scope

List work this spec excludes.

### Assumptions

List what the spec takes as given without a decision. Omit the section when empty.

### Further notes

Add open risks and anything else worth knowing. Omit the section when empty.

## Revisions

A revision is a change the user approved to User stories, Implementation decisions, Testing decisions, Out of scope, or Assumptions. Each revision increases the revision number and adds a one-line Change history entry naming the affected story IDs and what changed.

Fixes that don't change meaning, such as typos and wording, are editorial and keep the revision.

The attempt limit sets how many implementation attempts each ticket gets. Keep the default of 3 unless the user asks for another positive integer. Changing it isn't a revision.

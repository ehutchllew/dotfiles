# Specs

## Template

Text in `<...>` is a placeholder. Replace it with real content. The section guidance below the template explains each section.

```markdown
# <Feature name>

Revision: 1
Attempt limit: 3

## Problem statement

<the problem, from the user's point of view>

## Solution

<the solution, from the user's point of view>

## User stories

US 1: As a <actor>, I want <capability>, so that <benefit>

## Implementation decisions

- <decision>

## Testing decisions

- <decision>

## Out of scope

- <excluded work>

## Assumptions

- <assumption>

## Further notes

<anything else>

## Change history

- Revision 1: <summary>
```

## Rules for the whole spec

- Don't include file paths or code. They go stale. Sample terminal output and sample data are fine.
- Name modules, but don't give their paths. Describe interfaces by their inputs and outputs.
- Write a command exactly as typed in a terminal when it names no module path, such as `go test ./...`. Otherwise describe it by the module it runs and its arguments, such as "run the `ordersummary` command with `-code SAVE10`".
- If a prototype produced a snippet that states a decision more precisely than prose, such as a state machine, reducer, schema, or type shape, include it in the matching implementation decision. Keep only the part that carries the decision, and note that it came from a prototype.

## Section guidance

### Metadata

- `Revision:` starts at `1`. Increase it by one when the user approves a change to User stories, Implementation decisions, Testing decisions, Out of scope, or Assumptions, and add a Change history entry. Editorial fixes don't change the revision.
- `Attempt limit:` sets how many implementation attempts each ticket gets. It defaults to `3`. If the line is missing, the limit is `3`. Use another positive integer only when the user asks. Changing it doesn't change the revision.

### Problem statement and Solution

- Write both from the user's point of view.

### User stories

- Give each story an ID in the form `US <n>`.
- Example: `US 1: As a mobile bank customer, I want to see the balance on my accounts, so that I can make better decisions about my spending`
- Never renumber or reuse an ID. Add new stories at the end.
- To withdraw a story, keep its ID and replace its text with `Withdrawn in revision <n>.`
- Cover every behavior the feature needs. Don't add stories nobody asked for.

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

- List work this spec excludes.

### Assumptions

- List what the spec takes as given without a decision. Omit the section when empty.

### Further notes

- Add open risks and anything else worth knowing. Omit the section when empty.

### Change history

- Add one entry per revision: its number, the affected story IDs, and what changed.
- Never edit or delete earlier entries.

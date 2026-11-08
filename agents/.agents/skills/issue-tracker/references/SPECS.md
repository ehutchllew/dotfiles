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

## Further notes

<anything else>

## Change history

- Revision 1: <summary>
```

## Rules for the whole spec

- Don't include file paths or code. They go stale.
- Exception: if a prototype produced a snippet that states a decision more precisely than prose, such as a state machine, reducer, schema, or type shape, include it in the matching implementation decision. Keep only the part that carries the decision, and note that it came from a prototype.

## Section guidance

### Metadata

- `Revision:` starts at `1`. Increase it by one when the user approves a change to user stories, implementation decisions, or testing decisions, and add a Change history entry. Editorial fixes don't change the revision.
- `Attempt limit:` sets how many implementation attempts each ticket gets. It defaults to `3`. Use another positive integer only when the user asks. Changing it doesn't change the revision.

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
- technical clarifications from the developer
- architectural decisions
- schema changes
- API contracts
- specific interactions

### Testing decisions

List the decisions made. Include:

- what makes a good test: it checks external behavior, not implementation details
- which modules get tests
- prior art: similar tests already in the codebase
- the exact commands that run the tests and other checks, such as typecheck and lint, as typed in a terminal
- checks only the user can do, such as a manual QA pass

### Out of scope

- List work this spec excludes.

### Further notes

- Add open risks and anything else worth knowing. Omit the section when empty.

### Change history

- Add one entry per revision: its number, the affected story IDs, and what changed.
- Never edit or delete earlier entries.

# Specs

## Revision rules

Each spec has a plain `Revision: <number>` line below its title. Increment it when the user approves changes to requirements, implementation decisions, or verification, and records the changes in the spec's history. Editorial corrections alone do not need a new revision.

## Template

<spec-template>

# <Feature name>

Revision: 1

## Problem statement

The problem that the user is facing, from the user's perspective.

## Solution

The solution to the problem, from the user's perspective.

## User stories

A numbered list of user stories. Give each one a fixed ID in the format:

US 1: As an <actor>, I want a <feature>, so that <benefit>

<user-story-example>
US 1: As a mobile bank customer, I want to see balance on my accounts, so that I can make better informed decisions about my spending
</user-story-example>

Tickets refer to stories by ID, so never renumber them. Add new stories at the end.

Cover every behaviour the feature needs. Don't add stories nobody asked for.

## Implementation decisions

A list of implementation decisions that were made. This can include:

- The modules that will be built/modified
- The interfaces of those modules that will be modified
- Technical clarifications from the developer
- Architectural decisions
- Schema changes
- API contracts
- Specific interactions

Don't include specific file paths or code snippets. They may end up being outdated very quickly.

Exception: if a prototype produced a snippet that encodes a decision more precisely than prose can (state machine, reducer, schema, type shape), inline it within the relevant decision and note briefly that it came from a prototype. Trim to the decision-rich parts, not a working demo, just the important bits.

## Testing decisions

A list of testing decisions that were made. Include:

- A description of what makes a good test (only test external behavior, not implementation details)
- Which modules will be tested
- Prior art for the tests (i.e. similar types of tests in the codebase)
- The exact commands that run the tests and other checks (typecheck, lint), as typed in a terminal
- Checks only a person can do, such as a manual QA pass

## Out of scope

A description of the things that are out of scope for this spec.

## Further notes

Any further notes about the feature.

## Change history

For each approved revision, record its number, the affected user-story IDs, and what changed. Keep earlier entries. Treat this section as an append-only log.

</spec-template>

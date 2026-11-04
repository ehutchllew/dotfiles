---
name: to-spec
description: Turn the current conversation into a spec and publish it to the project issue tracker. No interview, just a synthesis of what you've already discussed. Use after a grilling session, or when the user asks for a spec.
disable-model-invocation: true
---

This skill takes the current conversation context and codebase understanding and produces a spec. Don't interview the user; synthesize what you already know.

## Process

1. Explore the repo to understand the current state of the codebase, if you haven't already. Use the vocabulary in `GLOSSARY.md` throughout the spec, and respect any ADRs in the area you're touching.

2. If the conversation has a shared-understanding summary from a grilling session, build the spec from it:
   - Decisions go into Implementation decisions.
   - Out of scope goes into Out of scope.
   - Open risks go into Further notes.
   - Reread the Assumptions. If one would change a user story, raise it with the user before writing.

3. Sketch out the seams at which you're going to test the feature. Existing seams should be preferred to new ones. Use the highest seam possible. If new seams are needed, propose them at the highest point you can. The fewer seams across the codebase, the better. The ideal number is one.

   Find the exact commands that run the tests and other checks (typecheck, lint) by reading the repo: `package.json`, `Makefile`, CI config, and similar. Check with the user that the seams and commands match their expectations.

4. Write the spec using the template below. Load the `issue-tracker` skill and publish the spec.

<spec-template>

# <Feature name>

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

</spec-template>

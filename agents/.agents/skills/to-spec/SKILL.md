---
name: to-spec
description: Turn the current conversation into a new or revised spec and publish it to the project issue tracker. Synthesize what has already been discussed. Use after a grilling session, or when the user asks for a spec.
disable-model-invocation: true
metadata:
    opencode/autoinvoke: false
---

# To spec

Turn the current conversation into a new or revised spec. Don't start a new interview. Ask a focused question if a gap would change observable behavior. Confirm the testing approach and commands, and ask for approval of revisions.

## Process

1. Load `issue-tracker` and read [SPEC-GUIDE.md](references/SPEC-GUIDE.md). Check for an existing spec and read it if present. Revise it instead of creating a second one. For a new feature, use the slug the user gave, or choose a short one such as `order-summary`.
2. Use the terms in `GLOSSARY.md` and respect relevant ADRs. Explore the code as needed.
3. If the conversation has a summary with decisions, assumptions, out-of-scope items, or open risks, build from it:
   - Decisions go into Implementation decisions.
   - Out-of-scope items go into Out of scope.
   - Assumptions go into Assumptions.
   - Open risks go into Further notes.
4. Where a behavior rule has a boundary that readers could interpret differently, write the agreed examples of what passes and what fails next to it.
5. Choose what this feature needs tested and tie each test to a user story or decision. Leave how to write tests to the Testing section of `CODING_STANDARDS.md`, and don't restate its rules. If there is no such section, use the repo's existing tests and conventions.
6. Collect the commands for Testing decisions:
   - Copy every command in the Checks section of `CODING_STANDARDS.md`, if present.
   - Find the repo's test and check commands in `package.json`, `Makefile`, CI config, and similar.
   - Confirm the testing approach and commands with the user.
7. Keep the default attempt limit unless the user asks for a different one.
8. Publish the spec through `issue-tracker`:
   - Create a new spec if none exists.
   - For a revision, show the user the proposed changes and publish only after approval. Treat fixes that don't change meaning as editorial.
   - Commit spec changes as `issue-tracker` directs.

## Revising a spec

- Follow the guide's rules for revisions, story IDs, and change history.
- Read ticket summaries through `issue-tracker` to see which work has started or finished. If a change contradicts that work, tell the user. Don't assume the work continues unchanged.
- Don't edit tickets. After you publish a revision, tell the user to run `to-tickets`.

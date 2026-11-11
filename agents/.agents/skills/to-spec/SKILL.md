---
name: to-spec
description: Turn the current conversation into a new or revised spec and publish it to the project issue tracker. Synthesize what has already been discussed. Use after a grilling session, or when the user asks for a spec.
disable-model-invocation: true
metadata:
    opencode/autoinvoke: false
---

# To spec

Turn the current conversation into a new or revised spec. Build it from what has already been discussed. Don't start a new interview. Ask the user only to confirm the testing approach and commands and, for a revision, the proposed changes.

## Process

1. Load `issue-tracker` and read the spec format. If the feature already has a spec, `fetch` it and revise it instead of creating a second one. For a new feature, use the slug the user gave, or choose a short one such as `order-summary`.
2. Use the terms in `GLOSSARY.md` and respect relevant ADRs. Explore the code as needed.
3. If the conversation has a summary with decisions, assumptions, out-of-scope items, or open risks, build from it:
   - Decisions go into Implementation decisions.
   - Out-of-scope items go into Out of scope.
   - Assumptions go into Assumptions.
   - Open risks go into Further notes.
   - If an assumption would change a user story, confirm it with the user first.
4. Choose what to test. Follow the Testing section of `CODING_STANDARDS.md`, and tie each test to a user story or decision. If the file or its Testing section is missing, suggest `to-standards` and stop.
5. Collect the commands for Testing decisions:
   - Copy every command in the Checks section of `CODING_STANDARDS.md`.
   - Find the commands that run the tests in the repo: `package.json`, `Makefile`, CI config, and similar.
   - Confirm the testing approach and commands with the user.
6. Keep the template's default `Attempt limit:` unless the user asks for a different one.
7. Write the spec. For a new spec, `create` it. For a revision, show the user the proposed changes and `edit` the spec only after they approve. Then commit the spec as the `issue-tracker` backend says.

## Revising a spec

- Follow the spec format's rules for revisions, story IDs, and change history.
- Use `list tickets` to see which work has started or finished. If a change contradicts that work, tell the user. Don't assume the work continues unchanged.
- Don't edit tickets. After you publish a revision, tell the user to run `to-tickets`.

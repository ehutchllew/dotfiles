---
name: to-spec
description: Turn the current conversation into a new or revised spec and publish it to the project issue tracker. Synthesize what has already been discussed. Use after a grilling session, or when the user asks for a spec.
disable-model-invocation: true
---

This skill takes the current conversation context and codebase understanding and produces a spec. Don't interview the user; synthesize what you already know.

## Process

1. Load `issue-tracker` and request the spec template, authoring guidelines, and revision rules. If the feature already has a spec, fetch it and its tickets before drafting. Request ticket metadata rules too if needed to interpret their execution state. Revise the existing spec rather than publishing a duplicate. Explore the repo if needed, use the vocabulary in `GLOSSARY.md`, and respect relevant ADRs.

2. If the conversation has a shared-understanding summary from a grilling session, build the spec from it:
   - Decisions go into Implementation decisions.
   - Out of scope goes into Out of scope.
   - Open risks go into Further notes.
   - Reread the Assumptions. If one would change a user story, raise it with the user before writing.

3. Sketch out the seams at which you're going to test the feature. Existing seams should be preferred to new ones. Use the highest seam possible. If new seams are needed, propose them at the highest point you can. The fewer seams across the codebase, the better. The ideal number is one.

   Find the exact commands that run the tests and other checks (typecheck, lint) by reading the repo: `package.json`, `Makefile`, CI config, and similar. Check with the user that the seams and commands match their expectations.

4. Write the spec using the template and guidelines obtained from `issue-tracker`, then publish it through the tracker. For a revision, show the proposed requirement and verification changes for approval before publishing them. Apply the revision process below.

## Revising an existing spec

- Pause Ralph before changing its approved scope. Do not edit the spec while an implementer or reviewer is running.
- Preserve existing user-story IDs. Add new stories at the end, never renumber or reuse IDs, and record changed or withdrawn requirements explicitly rather than silently deleting their history.
- Follow the tracker's revision contract. Record the approved changes, affected story IDs, and reason under `## Change history`. Preserve previous history.
- Do not rewrite completed tickets, reset execution metadata, or claim that a previous final review covers the revision. Leave ticket reconciliation to `to-tickets` and tell the user that Ralph must wait for it.
- If a new requirement contradicts already-started work, make that conflict explicit for the user rather than assuming the implementation should continue unchanged.

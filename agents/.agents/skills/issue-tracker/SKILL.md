---
name: issue-tracker
description: How to publish, fetch, list, comment on, and close issues and specs for the current repo. Load whenever a skill says "publish to the issue tracker", "fetch the relevant ticket", or otherwise needs to read or write issues.
---

# Issue Tracker

## Per-repo override

If the repo has `docs/agents/issue-tracker.md`, follow that file instead and ignore the rest of this skill.

## Default: local markdown

Issues and specs live as markdown files in `.scratch/` at the repo root. Whether `.scratch/` is committed or gitignored is up to the user; don't change it.

### Layout

- One feature per directory: `.scratch/<feature-slug>/`
- The spec is `.scratch/<feature-slug>/spec.md`
- Each ticket is its own file at `.scratch/<feature-slug>/issues/<NN>-<slug>.md`, numbered from `01`. Never combine tickets into one file.

### Ticket format

```markdown
# <NN>: <title>

Status: open
Mode: afk
Blocked by: 01, 02
Covers: US 4, US 7

## What to build

<the end-to-end behaviour this ticket makes work, from the user's point of view>

## Acceptance criteria

- [ ] <a criterion a test or command can check>

## Verification

<the tests or commands that prove each criterion, using commands from the spec's Testing Decisions>

## Out of scope

- <nearby work this ticket must not touch>

## Comments
```

- `Status:` is one of `open`, `in-progress`, `needs-review`, `done`.
- `Mode:` is `afk` when an agent can finish the ticket alone, and `hitl` when it needs a person, such as a manual QA pass, a third-party dashboard, or a product decision.
- `Blocked by:` lists ticket numbers. Leave the line out when nothing blocks the ticket. A ticket is unblocked when every ticket it lists is `done`.
- `Covers:` lists the spec's user stories this ticket implements.
- Write the metadata lines plain, without bold, so a script can match `^Status: `.
- Comments and conversation history go under `## Comments`, newest last.

### Operations

- **Publish to the issue tracker**: create a new file under `.scratch/<feature-slug>/`, creating the directory if needed. New tickets start as `Status: open`.
- **Fetch the relevant ticket**: read the file at the referenced path. The user will normally pass the path or the ticket number directly; if a number is ambiguous across features, ask.
- **List issues**: scan `.scratch/<feature-slug>/issues/` and read each file's `Status:` line.
- **List ready tickets**: tickets with `Status: open` and `Mode: afk` that are unblocked, lowest number first.
- **Comment**: append to the `## Comments` section.
- **Close**: append a comment explaining the outcome, then set `Status: done`.

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

Near the top of each ticket:

- `Status:` one of `open`, `in-progress`, `done`
- `Blocked by: NN, NN` (optional). A ticket is unblocked when every ticket it lists is `done`.

Comments and conversation history go at the bottom of the file under a `## Comments` heading, newest last.

### Operations

- **Publish to the issue tracker**: create a new file under `.scratch/<feature-slug>/`, creating the directory if needed. New tickets start as `Status: open`.
- **Fetch the relevant ticket**: read the file at the referenced path. The user will normally pass the path or the ticket number directly; if a number is ambiguous across features, ask.
- **List issues**: scan `.scratch/<feature-slug>/issues/` and read each file's `Status:` line.
- **Comment**: append to the `## Comments` section.
- **Close**: append a comment explaining the outcome, then set `Status: done`.

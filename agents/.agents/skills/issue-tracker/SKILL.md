---
name: issue-tracker
description: Storage operations and task-based access to spec and ticket templates and guidelines. Load whenever a skill needs to read or write issues or specs.
---

# Issue tracker

Callers state what information or operation they need, not which reference file to read. Resolve that request from the current task using the routing below. This skill supplies instructions to the current agent; it does not dispatch a separate tracker agent.

## Load the requested information

Before providing a template, interpreting contract metadata, or changing an artifact, read the matching reference:

| Requested information or operation | Read |
| --- | --- |
| Spec template, authoring guidelines, user-story IDs, or revision rules | [Spec contract](references/SPECS.md) |
| Spec requirements for drafting or reconciling a ticket breakdown | Both the spec contract and [ticket contract](references/TICKETS.md) |
| Ticket template, metadata, authoring, or reconciliation guidelines | Ticket contract |
| Ticket scheduling, implementation, review, baselines, attempts, final acceptance, persistence, or human hand-back | Ticket contract |
| Simple document fetch or listing, without interpreting contract metadata or changing state | Storage instructions below; no reference needed |

Read only the references the operation requires. When the task changes, load additional references as needed. Do not ask the caller to choose files, hardcode reference paths into callers, or load every reference by default.

Consuming an actual feature spec during implementation or review does not require the spec-authoring template. Ticket revision-matching rules are part of the ticket contract. Any operation that changes a spec or ticket must load its corresponding contract first.

## Per-repo override

If the repo has `docs/agents/issue-tracker.md`, use it for storage locations and operations instead of the default storage instructions below. The requested contracts in this skill's references still apply. If the override cannot preserve them, stop and explain the incompatibility rather than translating statuses or dropping metadata silently.

## Default storage: local markdown

Issues and specs live as markdown files in `.scratch/` at the repo root. Whether `.scratch/` is committed or gitignored is up to the user; don't change it.

### Layout

- One feature per directory: `.scratch/<feature-slug>/`
- The spec is `.scratch/<feature-slug>/spec.md`
- Each ticket is its own file at `.scratch/<feature-slug>/issues/<NN>-<slug>.md`, numbered from `01`. Never combine tickets into one file.

## Operations

- **Publish to the issue tracker**: create the approved spec or ticket files in the configured layout, creating directories as needed. Apply the requested contract before writing.
- **Fetch a spec or ticket**: read the file at the referenced path. The user will normally pass the path or ticket number directly; if a number is ambiguous across features, ask.
- **List issues**: scan `.scratch/<feature-slug>/issues/` and read each file's `Status:` line.
- **List ready tickets**: apply the ticket contract's readiness rules to the current metadata.
- **Persist a checkpoint**: save the update through the configured storage using the requested contract's persistence and ownership rules.
- **Comment**: append history in the location defined by the document's contract; do not overwrite earlier entries.
- **Close a ticket**: follow the ticket lifecycle and record completion evidence before closing it.

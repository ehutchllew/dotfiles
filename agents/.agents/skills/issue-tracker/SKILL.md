---
name: issue-tracker
description: Stores specs and tickets and defines their formats and operations. Load whenever a skill needs to read or write a spec or ticket.
---

# Issue tracker

This skill stores specs and tickets and defines their formats and operations. Read and write specs and tickets only through these operations, never directly from storage. This skill does not decide who runs an operation or when. The agent's task decides that.

## Backend

1. If `docs/agents/issue-tracker.md` exists, read its `Backend:` line. Otherwise use `local`.
2. Read `references/backends/<backend>.md` and no other backend file.
3. If that file doesn't exist, stop and report the unknown backend.
4. If the backend file has a Commits section, follow it whenever you commit specs or tickets.

## Features

A feature is one spec plus its tickets. Its slug, such as `order-summary`, is its ID. Every operation on a feature takes the slug.

## Formats

- For specs, read [SPECS.md](references/SPECS.md).
- For tickets, read [TICKETS.md](references/TICKETS.md).

Read the matching format before you create or edit a document, or when you need to interpret its fields.

## Secrets

Before you write a spec or ticket, replace secrets, such as API keys, tokens, passwords, and private keys, with `[REDACTED]`. This includes command output, evidence, and review findings in comments.

## Operations

### Storage operations

These work on specs and tickets. Every backend implements them.

| Operation | Result |
| --- | --- |
| `fetch` | Returns one spec or ticket. |
| `list features` | Returns each feature and the location of its spec. |
| `list tickets` | Returns each ticket in a feature with its number, title, and metadata lines. |
| `create` | Adds a new spec or ticket and saves it. |
| `edit` | Changes an existing spec or ticket and saves it. |

If a request matches no document or more than one, and you're a subagent, stop and report what you found. Otherwise ask the user.

### Ticket state operations

`claim`, `hand off`, `pass`, `send back`, `set feature base`, `start review`, `stop`, `resume`, `reset attempts`, `revise`, `complete`, `cancel`, `comment`.

`TICKETS.md` defines what each one changes. Each is a `fetch` followed by an `edit`, so backends don't implement them. Change a ticket's status, phase, attempt, or commits only through one of these.

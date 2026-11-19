---
name: issue-tracker
description: Reads and writes specs and tickets through the issue-tracker CLI. Load whenever a skill needs to read or write a spec or ticket.
---

# Issue tracker

Specs and tickets are read and written only through the `issue-tracker` command. Never read or edit their files directly. The command parses the documents, applies the ticket state rules, and validates features. This skill doesn't decide who runs an operation or when. The agent's task decides that.

If `issue-tracker` isn't on `PATH`, stop and notify the user.

## IDs and help

A feature is one spec plus its tickets. Its slug, such as `order-summary`, is its ID. Tickets are numbered `01`, `02`, and so on.

Run `issue-tracker help` for the command list. Before your first use of a write command, run `issue-tracker help <command>`, such as `issue-tracker help ticket send-back`. It lists the accepted fields with an input example.

## Reads

This is a non-exhaustive list:
| Command | Returns |
| --- | --- |
| `issue-tracker feature list` | Each feature and its title. |
| `issue-tracker status FEATURE` | Spec revision, attempt limit, each ticket's metadata, readiness, and hash, the current acceptance ticket, full validation, and unresolved not-blocking findings. No bodies. |
| `issue-tracker validate FEATURE [--draft]` | Validation errors. `--draft` skips the checks that need a complete breakdown. |
| `issue-tracker spec get FEATURE` | The spec and its hash. |
| `issue-tracker ticket list FEATURE` | Ticket summaries without bodies. |
| `issue-tracker ticket get FEATURE NN [--full]` | The ticket, hash, and `criteria` with their IDs. Comments before its last `revise` are left out unless you pass `--full`. |

Read with `--if-changed HASH` only when the content for that hash is still in your context. If nothing changed, the reply has `"unchanged": true` and no body. A fresh agent, or one whose context was compacted, reads without it. A hash from another agent doesn't mean you have the content.

## Writes

Writes take a JSON object on stdin. Use a quoted heredoc, so the shell doesn't expand anything in it:

```sh
issue-tracker ticket stop order-summary 03 --if-match sha256:0123456789abcdef --input - <<'EOF'
{"reason": "Verification blocked: the test database is down."}
EOF
```

- `spec create`, `spec edit`, `ticket create`, and `ticket edit` change authored content.
- Ticket state changes only through operations: `claim`, `hand-off`, `pass`, `send-back`, `set-feature-base`, `start-review`, `stop`, `resume`, `reset-attempts`, `revise`, `complete`, `cancel`, `comment`. Each takes its criteria changes, findings, verification, evidence, and reason in the same input and writes the comment.
- Writes to an existing document need `--if-match` with the hash from your last read or write. Each write returns the new hash.

Every reply is JSON. A failure has `error.code`:

| Code | Do this |
| --- | --- |
| `stale` | The document changed. Read it again, then decide whether the write still applies. |
| `invalid_transition`, `not_ready`, `attempt_limit_reached`, `stale_acceptance`, `not_current_acceptance` | The ticket or feature isn't in a state that allows the operation. Don't retry it as is. |
| `would_invalidate` | The write would break validation. The message and details say how. |
| `unknown_criterion`, `duplicate_criterion`, `ambiguous_criterion` | Use each applicable ID from `ticket get` once. Don't guess which criterion was meant. |
| `unknown_finding` | Quote the finding exactly as the ticket has it. |
| `ticket_not_found`, `ambiguous_ticket`, `feature_not_found` | If you're a subagent, stop and report what you found. Otherwise ask the user. |
| `partial_write` | Run the same operation again to finish it. |
| anything else | Report the message. |

## Commits

Each write reply lists `changed` paths and a `commit_subject`. Commit only those paths when you commit. If `git_ignored` is true, skip the commit. For several operations in one commit, use the subject `tickets: multiple` with one `<operation> <NN>` line per operation in the body, such as `cancel 04`.

## Terms

- Kinds: `implementation` is code work, `acceptance` is the final review of the whole feature, and `manual` is work only the user can do.
- Implementation tickets have a phase, `implement` or `review`, and an attempt count. Every ticket has a status: `open`, `in-progress`, `stopped`, `done`, or `cancelled`.
- `Base commit:` starts a review range and never changes. `Target commit:` ends it.
- A ticket is ready when it's `open` or `in-progress` and every blocker is `done`. Being ready doesn't mean someone should start it.
- The current acceptance ticket is the highest-numbered acceptance ticket that isn't cancelled.
- A `Not blocking:` finding stays unresolved until a later comment on its ticket resolves it as fixed, ignored, or moved to another ticket.

## Secrets

Before you write, replace secrets, such as API keys, tokens, passwords, and private keys, with `[REDACTED]`. This includes command output, evidence, and review findings. The CLI doesn't scan for secrets.

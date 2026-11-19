# issue-tracker

`issue-tracker` reads and writes feature specs and tickets for the agent skills in `agents/`. The skills call it instead of editing ticket files by hand. It parses the documents, enforces the ticket state rules, validates features, and prints small JSON responses.

It doesn't judge code, approve plans, run verification commands, or choose review outcomes. The skills still do that.

## Install

Go comes from mise. `stow agents` links the skills but doesn't build this binary, so install it separately and again after pulling changes to this directory:

```sh
mkdir -p ~/.local/bin
cd tools/issue-tracker
GOBIN="$HOME/.local/bin" go install ./cmd/issue-tracker
```

`zsh/.config/zsh/exports.zsh` adds `~/.local/bin` to `PATH` only when the directory exists, so open a new shell after creating it.

## Use

```sh
issue-tracker help                      # command list
issue-tracker help ticket send-back     # flags and an input example
issue-tracker status order-summary
issue-tracker ticket get order-summary 03
echo '{"reason":"Out of disk."}' | issue-tracker ticket stop order-summary 03 --if-match sha256:... --input -
```

The target repository is the nearest directory above the working directory with a `.git` entry. A git worktree has a `.git` file, so it uses its own `.scratch/`. `--repo DIR` overrides the search. The install location never matters.

Every command prints one JSON object on stdout. `--pretty` indents it. Failures print `{"error":{"code":"...","message":"...","details":{...}}}` and exit with 1, or 2 for bad arguments or input. A conditional read that finds no change is a success. If stdout can't be written, the command explains on stderr. A command that succeeded then exits with 1. A command that failed keeps its exit status of 1 or 2, and stderr shows the JSON error code and message it couldn't print. Changes may already be on disk, so read the documents again before retrying.

Writes take one JSON object through `--input FILE`, or `--input -` for stdin. Unknown or misspelled fields fail. Writes to an existing document need `--if-match HASH` with the hash from the caller's last read. A mismatch fails with code `stale` and the current hash. Write responses list the changed paths, their new hashes, and a commit subject. The CLI never commits, stages, or changes git configuration. It does read git to resolve refs such as `HEAD` and to report whether storage is ignored.

## Storage

The backend comes from the `Backend:` line in `docs/agents/issue-tracker.md`. Without the file or the line, the backend is `local`. Any other backend fails with `unsupported_backend`.

The local backend keeps one directory per feature:

```text
.scratch/
└── <feature-slug>/
    ├── spec.md
    └── issues/
        ├── 01-<ticket-slug>.md
        └── 02-<ticket-slug>.md
```

Feature IDs are single path segments such as `order-summary`. The CLI refuses symlinks anywhere under `.scratch/`. Locks live in `$TMPDIR/issue-tracker-<uid>/` (or `ISSUE_TRACKER_LOCK_DIR`), keyed by the canonical working-tree path and feature, so reads don't add files to the checkout. Each file is replaced atomically. An operation that changes several files, such as `cancel`, writes the target ticket first and reports `partial_write` if a later file fails. Running the same operation again finishes it.

### Commit subjects

Write responses include `commit_subject`:

| Change | Subject |
| --- | --- |
| Create a spec | `spec: create <feature-slug>` |
| Edit a spec | `spec: revise <feature-slug>` |
| One ticket operation | `tickets: <operation> <NN>`, such as `tickets: send back 03` |
| Several ticket operations | `tickets: multiple`, with one `<operation> <NN>` line per operation in the body |

Ticket commit subjects never start with a ticket number, which keeps them out of a ticket's review range. If `git_ignored` is true, skip the commit.

## Hashes and conditional reads

A document's hash is `sha256:` plus the first 16 hex digits of the SHA-256 of its file bytes. The spec and each ticket have separate hashes, so a ticket write doesn't change the spec hash. The spec's `Revision:` isn't a cache key: editorial fixes and attempt-limit changes change the hash without a new revision.

`spec get --if-changed HASH` and `ticket get --if-changed HASH` print only `{"hash":...,"unchanged":true}` when nothing changed. Callers pass a hash only for content still in their own context. Nothing stores a "last read" marker.

## Spec format

```markdown
# <Feature name>

Revision: 1
Attempt limit: 3

## Problem statement
## Solution
## User stories

US 1: As a <actor>, I want <capability>, so that <benefit>

## Implementation decisions
## Testing decisions
## Out of scope
## Assumptions          (optional)
## Further notes        (optional)
## Change history

- Revision 1: <summary>
```

- `Attempt limit:` defaults to 3 when missing.
- Story IDs run from `US 1` without gaps. They never change. A withdrawn story keeps its ID with the text `Withdrawn in revision <n>.`
- `spec edit` needs `change: editorial` or `change: revision` for content changes. A revision needs `approved: true` and a `history` line. The CLI adds 1 to `Revision:` and appends the Change history entry. It doesn't decide whether a change is editorial.

## Ticket format

```markdown
# <NN>: <title>

Kind: implementation | acceptance | manual
Phase: implement | review              (implementation)
Status: open | in-progress | stopped | done | cancelled
Attempt: <n>                           (implementation)
Blocked by: 01, 02                     (omitted when empty)
Covers: US 1, US 2 | none
Spec revision: <n>                     (acceptance)
Base commit: <sha>                     (set by claim or set-feature-base)
Target commit: <sha>                   (set by hand-off or start-review)

## What to build
## Acceptance criteria

- [ ] C01: <criterion>

## Verification
## Out of scope                        (optional)
## Comments
```

`ticket edit` changes content only. Status, phase, attempt, commits, and comments change only through state operations. Done and cancelled tickets can't be edited, only commented on.

`ticket get` returns the criterion IDs and text. Create and edit still take criterion text. Unchanged text keeps its ID when reordered. New or replaced criteria get new IDs, and `Next criterion:` keeps deleted IDs from being reused. Older tickets get IDs in their original order without a rewrite. If you add a criterion by hand without an ID, it gets one from its place in the list. When that matches an ID written on another criterion, validation reports `duplicate_criterion`, and you fix it by writing an unused ID on the new line. State operations take criterion IDs and still accept exact text from older callers.

### State operations

`From` and `To` show `phase/status` for implementation tickets and `status` for the other kinds.

| Operation | Kind | From | To | Also |
| --- | --- | --- | --- | --- |
| `claim` | implementation | implement/open | implement/in-progress | Adds 1 to `Attempt`. Sets `Base commit` if missing. Needs every blocker done and `Attempt` below the limit. |
| `claim` | implementation | implement/in-progress | no change | The current attempt continues. |
| `hand-off` | implementation | implement/in-progress | review/open | Sets `Target commit`. Ticks met criteria. Records verification. |
| `pass` | implementation | review/open | review/done | Comments. |
| `send-back` | implementation | review/open | implement/open | Unticks unmet criteria. Comments with findings. |
| `set-feature-base` | acceptance | open | no change | Sets `Base commit`. Fails if already set. |
| `start-review` | acceptance | open | in-progress | Sets `Target commit`. Needs the current acceptance ticket, a base commit, the spec's revision, and every blocker done. |
| `start-review` | acceptance | in-progress | no change | The review continues. |
| `pass` | acceptance | in-progress | done | Needs the current acceptance ticket, the spec's revision, and every blocker done. Comments. Can tick every criterion. |
| `complete` | manual | open | done | Ticks every criterion. Comments with evidence. |
| `stop` | any | open or in-progress | stopped | Comments with the reason. Phase doesn't change. |
| `resume` | any | stopped | open | Comments with the resolution. |
| `reset-attempts` | implementation | implement/stopped | implement/open | Sets `Attempt` to 0. |
| `revise` | implementation | any phase; open, in-progress, or stopped | implement/open | Sets `Attempt` to 0. Unticks every criterion. Keeps both commits. |
| `revise` | acceptance, manual | open, in-progress, or stopped | open | Unticks every criterion. |
| `cancel` | any | anything except done | cancelled | Removes the ticket from every `Blocked by:` line. |
| `comment` | any | any | no change | Appends a comment. |

### Comments

Comments are append-only. Headers use the phase before the operation, or the ticket kind when there is no phase. Old headers remain readable.

```markdown
review:send-back (attempt 1):
- Unmet criteria: C01: "An empty order shows an item count of zero."
- Blocking (missing): an empty order shows no item count line.
- Verification:
  - `npm test` passed.
- Not blocking: the local `n` doesn't say it holds the item count.

review:pass (attempt 2):
- Verification:
  - `npm test` passed.
- Not blocking resolved: the local `n` doesn't say it holds the item count. Fixed.
```

A multi-line value goes on the lines below its label, so fenced blocks keep their own lines.

A not-blocking finding stays unresolved until a later comment on the same ticket has a `Not blocking resolved:` item for it ending in `Fixed.`, `Ignored.`, or `Moved to ticket <NN>.` A comment that only resolves findings has no attempt label.

`ticket get` leaves out comments before the ticket's last `revise` comment by default. `--full` returns the whole file.

### Readiness and the current acceptance ticket

A ticket is ready when its status is `open` or `in-progress` and every ticket in its `Blocked by:` line is `done`. The current acceptance ticket is the highest-numbered acceptance ticket that isn't cancelled.

## Validation

`validate --draft` runs the checks that hold while a breakdown is incomplete. A ticket write fails with `would_invalidate` if it would give any ticket one of these problems that it didn't already have. Problems that existed before the write don't block it, so a broken feature can be repaired one write at a time:

1. Each ticket has the fields its kind needs, with allowed values, and no fields for other kinds.
2. An implementation ticket in the review phase is open, stopped, done, or cancelled.
3. An implementation ticket with `Attempt` above 0 has a base commit. One in the review phase also has a target commit.
4. Every blocker exists and isn't cancelled.
5. No chain of blockers forms a cycle.
6. An in-progress acceptance ticket has every blocker done.
7. Every covered story exists. Unfinished tickets don't cover withdrawn stories.

`validate` and `status` also check the whole feature:

8. Every story that isn't withdrawn is covered by an implementation ticket that isn't cancelled.
9. A current acceptance ticket exists, covers every story that isn't withdrawn, and has the spec's revision.
10. The current acceptance ticket is blocked by every implementation and manual ticket that isn't cancelled.

## Decisions beyond the earlier written rules

- `cancel` works on an implementation ticket in the review phase, and validation accepts review/cancelled.
- `revise` on a manual ticket reopens it, unticks its criteria, and comments with the reason.
- Check 7 doesn't flag a withdrawn story on a done or cancelled ticket, because those tickets can't be edited.
- `claim` checks readiness and the attempt limit. `start-review` checks the base commit, the spec revision, and readiness. An acceptance `pass` checks the spec revision, the current acceptance ticket, and readiness again, so a review of an outdated spec can't pass. These checks read the current blockers and spec under the feature lock.
- Adding blockers to an in-progress acceptance ticket fails until the ticket is revised, so `to-tickets` revises a ticket before editing it.
- Unknown metadata lines such as `Owner:` fail validation, so a misspelled field doesn't go unnoticed.

## Development

```sh
cd tools/issue-tracker
go test ./...
go vet ./...
```

The tests run the CLI the way agents do. Each test copies hand-written features from `internal/cli/testdata/features` into a temporary directory, passes arguments and stdin to `cli.Run`, and checks the JSON reply, the exit code, and the bytes on disk. `order-summary` holds the documented examples, and `checkout` has two stories and four tickets. One test runs the CLI as separate processes to check locking. Most tests use an empty `.git` directory and full SHAs, so one test creates a real git repository to check ref resolution and `git_ignored`. Finished documents are compared with files in `internal/cli/testdata/golden`. After an intended format change, run `go test ./internal/cli -update` and review the diff.

The only test below the CLI is in `internal/tracker`. It swaps in a storage backend that fails one write, because a disk error on a single file can't be caused from outside.

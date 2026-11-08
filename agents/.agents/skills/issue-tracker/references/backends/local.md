# Local backend

Specs and tickets are markdown files under `.scratch/` at the repo root.

## Layout

```text
.scratch/
└── <feature-slug>/
    ├── spec.md
    └── issues/
        ├── 01-<ticket-slug>.md
        └── 02-<ticket-slug>.md
```

- Each feature has one directory.
- Each ticket has its own file, numbered from `01`. Never put two tickets in one file.

## Operations

| Operation | How |
| --- | --- |
| `fetch` | Read the file. |
| `list features` | List the directories in `.scratch/` that contain a `spec.md`. |
| `list tickets` | Read the title and metadata lines of each file in the feature's `issues/` directory. |
| `create` | Write a new file and create directories as needed. A new ticket takes the next unused number. |
| `edit` | Rewrite the file in place. |

Don't change whether `.scratch/` is tracked or ignored by git. Committing these files is up to the caller.

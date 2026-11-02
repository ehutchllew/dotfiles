---
name: code-review
description: "Review the changes since a fixed point (commit, branch, tag, or merge-base) along two separate axes: Standards (does the code follow this repo's documented coding standards?) and Spec (does it do what the originating issue or spec asked for?). Use when the user wants a branch, PR, or work-in-progress reviewed, asks whether changes are ready to merge, says \"check my changes\" or \"review since X\", or wants to know if an implementation matches its ticket."
---

Two-axis review of the diff between `HEAD` and a fixed point the user supplies:

- **Standards**: does the code conform to this repo's documented coding standards?
- **Spec**: does the code faithfully implement the originating issue / spec?

Both axes run as **parallel sub-agents** so they don't pollute each other's context, then this skill aggregates their findings. Start every sub-agent with a fresh context window.

If issue was not provided by the user, request it before proceeding.

## Process

### 1. Pin the fixed point

Whatever the user said is the fixed point (a commit SHA, branch name, tag, `main`, `HEAD~5`, etc.). If they didn't specify one, ask for it.

Capture the diff command once: `git diff <fixed-point>...HEAD` (three-dot, so the comparison is against the merge-base). Also note the list of commits via `git log <fixed-point>..HEAD --oneline`.

Then check `git status --short` for uncommitted changes. If they belong to the same work as the commits since the fixed point (same feature, same files), include them: use `git diff $(git merge-base <fixed-point> HEAD)` instead, which covers both commits and uncommitted tracked changes, and list any relevant untracked files for the sub-agents to read. If they look unrelated, leave them out and say so in the final report.

Before going further, confirm the fixed point resolves (`git rev-parse <fixed-point>`) and the diff is non-empty. A bad ref or empty diff should fail here, not inside two parallel sub-agents.

### 2. Identify the spec source

Look for the originating spec, in this order:

1. Issue references in the commit messages (`#123`, `Closes #45`, GitLab `!67`, etc.), fetched via the `issue-tracker` skill.
2. A path the user passed as an argument.
3. A spec file under `docs/`, `specs/`, or `.scratch/` matching the branch name or feature.
4. If nothing is found, ask the user where the spec is. If they say there isn't one, the **Spec** sub-agent will skip and report "no spec available".

### 3. Identify the standards sources

Search the repo for every file that documents how code should be written. When `CODING_STANDARDS.md` or `CONTRIBUTING.md` exists, it must be on the list.

On top of whatever the repo documents, the Standards axis always carries the **smell baseline** in [references/smell-baseline.md](references/smell-baseline.md): a fixed set of Fowler code smells that applies even when a repo documents nothing. You don't need to read it yourself; the Standards sub-agent reads it.

### 4. Spawn both sub-agents in parallel

Issue both sub-agent calls together, in the foreground, and aggregate the reports they return.

**Standards sub-agent prompt** should include:

- The full diff command and commit list.
- The list of standards-source files you found in step 3, **plus the absolute path to `references/smell-baseline.md`** in this skill's directory, with an instruction to read it.
- The brief: "Report, per file/hunk where relevant, (a) every place the diff violates a documented standard: cite the standard (file + the rule); and (b) any baseline smell you spot: name it and quote the hunk. Distinguish hard violations from judgement calls: documented-standard breaches can be hard, but baseline smells are always judgement calls, and a documented repo standard overrides the baseline. Skip anything tooling enforces. Under 400 words."

**Spec sub-agent prompt** should include:

- The diff command and commit list.
- The path or fetched contents of the spec.
- The brief: "Report: (a) requirements the spec asked for that are missing or partial; (b) behaviour in the diff that wasn't asked for (scope creep); (c) requirements that look implemented but where the implementation looks wrong. Quote the spec line for each finding. Under 400 words."

If the spec is missing, skip the Spec sub-agent and note this in the final report.

### 5. Aggregate

Present the two reports verbatim or lightly cleaned. Do **not** merge or rerank findings, because the two axes are deliberately separate (see _Why two axes_). Use this template:

```markdown
Reviewed `<fixed-point>...HEAD` (<N> commits<, plus uncommitted changes>).

## Standards

<Standards sub-agent report>

## Spec

<Spec sub-agent report, or "No spec available.">

**Summary:** Standards: <n> findings, worst: <issue or "none">. Spec: <n> findings, worst: <issue or "none">.
```

Don't pick a single winner across axes in the summary: that's the reranking the separation exists to prevent.

## Why two axes

A change can pass one axis and fail the other:

- Code that follows every standard but implements the wrong thing → **Standards pass, Spec fail.**
- Code that does exactly what the issue asked but breaks the project's conventions → **Spec pass, Standards fail.**

Reporting them separately stops one axis from masking the other.

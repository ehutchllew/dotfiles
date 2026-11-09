---
name: code-review
description: Review changes, an implementation ticket, or final feature acceptance along separate Standards and Spec axes. Use when reviewing a branch or work in progress, checking an implementation against its ticket, or verifying a complete feature.
---

# Code review

Review code on two separate axes:

- **Standards**: does the changed code follow the repo's documented standards?
- **Spec**: does the code do what the requirements in scope ask?

This skill reports findings and verification results. It doesn't decide whether the work passes. That's the caller's job. It never edits files, commits, or writes specs or tickets.

Use fresh, read-only subagents for the two axes, in parallel, when available. Otherwise review the axes one after the other yourself and keep their findings separate. Don't call that fallback an independent review.

## 1. Inputs

The caller may supply any of these. Use what's supplied and don't ask for it again.

| Input | Meaning |
| --- | --- |
| mode | Change review, ticket review, or feature acceptance. |
| range | The base commit and the target commit. |
| feature | The feature whose spec holds the requirements. |
| ticket | A ticket number in that feature. |
| verification commands | Commands to run instead of the ones the mode names. |
| earlier blocking findings | Findings from earlier reviews of the same work. |

Load `issue-tracker` to read specs and tickets, and read them only through its operations. If the caller gives a ticket and no range, read the ticket's `Base commit:` and `Target commit:`.

### Choose the mode

If the caller names no mode:

- An acceptance ticket means feature acceptance.
- An implementation ticket with a `Target commit:` means ticket review.
- Anything else means change review.

### Modes

| Mode | Range | Code to review | Requirements | Verification commands |
| --- | --- | --- | --- | --- |
| Change review | The caller's base. The target defaults to the working tree. | The diff. | A spec or ticket, if one is available. | The caller's, if any. |
| Ticket review | The ticket's `Base commit:` to `Target commit:`. | Commits in the range whose subject starts with `<NN>: `, plus the code at the target. | The ticket and the spec. | The ticket's Verification section, or the spec's Testing decisions if it has none. |
| Feature acceptance | The acceptance ticket's `Base commit:` to `Target commit:`. | The whole range. | The current spec. | The acceptance ticket's Verification section and the spec's Testing decisions. |

In a change review with no requirements supplied:

1. Look for ticket numbers at the start of commit subjects in the range. `fetch` those tickets and their spec.
2. If there are none, ask the user for a feature or ticket.
3. If the user says there is none, skip the Spec axis and report "No spec available."

If the caller supplied no earlier blocking findings, take them from the ticket's `send back` comments in a ticket review, ignoring any made before its last `revise`. In feature acceptance, take them from the `stop` comments on the current and earlier acceptance tickets.

## 2. Pin the range

1. Resolve each end with `git rev-parse --verify <ref>^{commit}`. If one doesn't resolve, stop and report it.
2. Check `git merge-base --is-ancestor <base> <target>`. If the base isn't an ancestor, stop and report it.
3. If a ticket is missing `Base commit:` or `Target commit:`, stop and report it. Never substitute `HEAD`, another ticket's commit, or a guess.
4. Collect the code to review:
   - Change review: `git diff <base>...<target>`. When the target is the working tree, use `git diff $(git merge-base <base> HEAD)` and list relevant untracked files. Leave out unrelated changes and say so in the report.
   - Ticket review: list the commits with `git log --format=%H --grep='^<NN>: ' <base>..<target>`, then read each with `git show`. An empty list isn't a finding by itself. Review the code at the target either way.
   - Feature acceptance: `git diff <base>..<target>` and `git log --oneline <base>..<target>`.
5. An empty diff ends a change review with "No changes to review." It doesn't end a ticket review or feature acceptance.

## 3. Verify

1. In ticket review and feature acceptance, check `git status --short` first. If the checkout has changes, report verification as `blocked` and list the changes.
2. Run the mode's verification commands yourself, even when someone reports they passed. Record each command and its result.
3. In feature acceptance, read the evidence on every `done` manual ticket. Don't perform a manual check yourself.
4. Don't change source, tests, snapshots, or requirements to make a check pass. If a command changes files, leave the changes and report them.

Report one verification status:

| Status | When |
| --- | --- |
| `passed` | Every command ran and succeeded, and every required piece of manual evidence exists. |
| `failed` | A command ran and failed. Each failure is also a blocking finding. |
| `blocked` | There are no commands, a command couldn't run, manual evidence is missing, or the checkout had changes. |
| `not run` | Change review only, when the caller gave no commands. |

Give the results to both axes so they don't rerun the commands.

## 4. Find standards sources

Read these, in order, when present:

1. `CODING_STANDARDS.md`
2. `CONTRIBUTING.md`
3. Other documented coding rules in the repo.

The Standards axis also reads [references/smell-baseline.md](references/smell-baseline.md). Documented repo rules override it.

## 5. Review the two axes

Start both subagents together. End each prompt with:

> Perform this review directly. Do not load the `code-review` skill, spawn subagents, or modify any files.

Give both the mode, the range, the code-to-review commands or commit list, the feature and ticket numbers, earlier blocking findings, and the verification results. They read specs and tickets through `issue-tracker`.

### Standards prompt

Include the path of every standards source and the absolute path of `references/smell-baseline.md`, and tell the subagent to read them.

Ask for:

- Violations of documented rules. These are hard violations.
- Baseline smells. These are judgement calls.
- Earlier blocking Standards findings that are still present, even if the latest change didn't touch that code.

Cite the file, the hunk, and the rule for each finding. Skip anything tooling enforces. In a ticket review, review only the ticket's commits. Keep the report under 400 words.

### Spec prompt

Ask for:

- Requirements in scope that are missing or partly done.
- Behavior nobody asked for. Unused abstractions and parameters belong to Standards, not here.
- Behavior that is wrong or contradicts the requirements.
- Earlier blocking findings that are still unresolved.
- Failed verification that shows unmet behavior.

Scope by mode:

- Ticket review: the ticket sets the scope. `Covers:` links it to stories but doesn't make it responsible for other tickets' work. Read the code at the target beyond the changed lines when that's needed to tell whether the ticket works.
- Feature acceptance: the whole current spec sets the scope. Check that all stories work together. Read current code and tests even when they predate the range. Don't treat ticket statuses as proof. Read tickets only for manual evidence, and report ticket problems as not blocking.

Quote the requirement and cite the code for each finding. Keep the report under 400 words.

## 6. Mark each finding

| Finding | Blocking |
| --- | --- |
| A hard violation of a documented standard | yes |
| Missing, partial, wrong, or unrequested behavior in scope | yes |
| An earlier blocking finding that is still unresolved | yes |
| A failed verification command | yes |
| A baseline smell or other judgement call | no |
| A ticket problem found during feature acceptance | no |

Blocked verification isn't a finding. It shows only in the verification status.

## 7. Report

Present the two axis reports as written, lightly cleaned. Don't merge, rerank, or weigh them against each other. Use this template:

```markdown
Reviewed `<base>..<target>` in <change review | ticket review | feature acceptance> mode.

## Verification

Status: <passed | failed | blocked | not run>

<each command and its result, manual evidence checked, and the reason if blocked>

## Standards

- Blocking: <finding>
- Not blocking: <finding>

## Spec

- Blocking: <finding>
- Not blocking: <finding>

<or "No spec available.">

Summary: Standards <n> blocking, <n> not blocking. Spec <n> blocking, <n> not blocking. Verification <status>.
```

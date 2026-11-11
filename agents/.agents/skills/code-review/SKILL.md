---
name: code-review
description: Review changes, an implementation ticket, or final feature acceptance along separate Standards and Spec axes. Use when reviewing a branch or work in progress, checking an implementation against its ticket, or verifying a complete feature.
---

# Code review

Review code on two separate axes:

- **Standards**: does the changed code follow the repo's documented standards?
- **Spec**: does the code do what the requirements in scope ask?

This skill reports findings and verification results. It never edits files, commits, or writes specs or tickets.

Review each axis in a fresh, read-only subagent, both in parallel, when subagents are available. Otherwise review the axes one after the other yourself, keep their findings separate, and note in the report that one agent reviewed both.

## 1. Gather inputs

Use what the caller supplies. Infer the rest, and ask only for what you can't infer. Load `issue-tracker` to read specs and tickets.

| Input | Values | If not supplied |
| --- | --- | --- |
| Ticket | A feature and a ticket number | Review a range instead. |
| Range | A base commit and a target commit | Read them from the ticket. Otherwise ask for the base, and treat the target as the working tree. |
| Commit filter | A pattern that matches the subjects of the commits to review | Review every commit in the range. |
| Context | Anything else to use, such as requirement documents, commands to run, or findings from an earlier review | With a ticket, use its spec, its Verification section, and its comments. Without one, look for ticket references in the commit messages, then ask for requirements. If there are none, skip the Spec axis. |

Context only adds to what a ticket provides. It never replaces the ticket's requirements or commands.

The ticket decides the mode:

- Feature acceptance: an acceptance ticket. Review everything in the range against the whole current spec. Read tickets only for manual evidence.
- Ticket review: an implementation ticket with both `Base commit:` and `Target commit:`. Review its work against the ticket and its spec.
- Change review: anything else. Review a diff against whatever requirements exist, including a ticket that isn't ready for ticket review.

## 2. Pin the range

Resolve both ends with `git rev-parse --verify <ref>^{commit}`. Then check that the base is in the target's history with `git merge-base --is-ancestor <base> <target>`. If either check fails, the range doesn't describe the work, so stop and report it. Never replace a missing ticket commit with `HEAD` or a guess.

Collect the code to review:

- With a commit filter, list the matching commits in the range and read each one with `git show`. Also read the code at the target, because later commits may have changed the same lines.
- Without one, diff the range. When the target is the working tree, include uncommitted changes and relevant untracked files. Leave out unrelated changes and say so in the report.

An empty diff ends a change review with "No changes to review." In the other modes, review the code at the target anyway.

## 3. Verify

1. Verification runs on the working tree, so the working tree must hold the code under review:
   - If the target is the working tree, run there.
   - If the target is a commit, run only when `git status --short` prints nothing and `git diff --name-only <target> HEAD` lists nothing, or only spec and ticket files. Otherwise report verification as `blocked` and list what differs.
2. Run the verification commands yourself, even when someone reports they passed. Record each command and its result.
3. In feature acceptance, read the evidence on manual tickets. Don't perform a manual check yourself.

Report one verification status:

| Status | When |
| --- | --- |
| `passed` | Every command ran and succeeded, and every required piece of manual evidence exists. |
| `failed` | A command ran and failed. |
| `blocked` | There are no commands, a command couldn't run, manual evidence is missing, or the working tree doesn't hold the target's code. |
| `not run` | Change review only, when there are no commands. |

## 4. Find standards sources

Read these, in order, when present:

1. `CODING_STANDARDS.md`
2. `CONTRIBUTING.md`
3. Other documented coding rules in the repo.

## 5. Review the two axes

Fill in [STANDARDS_PROMPT.md](references/STANDARDS_PROMPT.md) and [SPEC_PROMPT.md](references/SPEC_PROMPT.md), and send one to each subagent. If you review the axes yourself, follow the same prompts.

## 6. Mark each finding

Each axis labels its findings with a category. Mark them:

| Category | Blocking |
| --- | --- |
| hard violation | yes |
| missing, wrong, or unrequested behavior | yes |
| unresolved earlier finding | yes |
| failed verification | yes |
| smell | no |
| ticket problem | no |

Blocked verification isn't a finding. It shows only in the verification status. A fixed earlier finding isn't a finding either. Report it on a `Fixed:` line.

## 7. Report

Present the two axis reports as written, lightly cleaned. Don't merge, rerank, or weigh them against each other. Use this template:

```markdown
Reviewed `<base>..<target>` in <change review | ticket review | feature acceptance> mode.

## Verification

Status: <passed | failed | blocked | not run>

<each command and its result, manual evidence checked, and the reason if blocked>

## Standards

- Blocking (<category>): <finding>
- Not blocking (<category>): <finding>
- Fixed: <earlier not-blocking finding>

## Spec

- Blocking (<category>): <finding>
- Not blocking (<category>): <finding>
- Fixed: <earlier not-blocking finding>

<or "No spec available.">

Summary: Standards <n> blocking, <n> not blocking. Spec <n> blocking, <n> not blocking. Verification <status>.
```

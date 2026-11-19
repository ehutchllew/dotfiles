---
name: code-review
description: Review changes, an implementation ticket, or final feature acceptance along separate Standards and Spec axes. Use when reviewing a branch or work in progress, checking an implementation against its ticket, or verifying a complete feature.
---

# Code review

Review code on two separate axes:

- **Standards**: does the changed code follow the repo's documented standards?
- **Spec**: does the code do what the requirements in scope ask?

This skill reports findings and verification results. It never edits files, commits, or writes specs or tickets.

Perform both axes yourself, using shared inputs and verification results. Don't spawn subagents.

## 1. Gather inputs

Use what the caller supplies. Infer the rest, and ask only for what you can't infer. Load `issue-tracker` to read specs and tickets. With a ticket, read it and its spec. For feature acceptance, also read the manual tickets needed for evidence.

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
3. In feature acceptance, check that the evidence on manual tickets covers the required scenarios for the code under review. Reuse earlier evidence if the relevant code and scenarios haven't changed. Don't perform a manual check yourself.

Report one verification status:

| Status | When |
| --- | --- |
| `passed` | Every command ran and succeeded, and the required manual evidence applies to the target. |
| `failed` | A command ran and failed. |
| `blocked` | There are no commands, a command couldn't run, required manual evidence is missing or stale, or the working tree doesn't hold the target's code. |
| `not run` | Change review only, when there are no commands. |

## 4. Find standards sources

Read these, in order, when present:

1. `CODING_STANDARDS.md`
2. `CONTRIBUTING.md`
3. Other documented coding rules in the repo.

Also read the [smell baseline](references/smell-baseline.md).

## 5. Review the two axes

Collect earlier blocking findings and unresolved not-blocking findings from Context and ticket comments. Ignore findings recorded before the ticket's last `revise`. Recheck blocking findings even if the latest change didn't touch their code.

Report earlier not-blocking findings that are now gone as `Fixed:`. Don't repeat those still present.

### Standards

Review only the changed code and earlier blocking Standards findings. Documented repo rules override the smell baseline. Skip anything tooling enforces. Cite the file, hunk, and rule for each finding.

### Spec

Check the requirements for the current mode:

- Change review: the available requirements.
- Ticket review: the ticket. Its covered stories don't make it responsible for other tickets' work.
- Feature acceptance: the whole current spec. Check that all stories work together. Read code and tests even when they predate the range. Read other tickets only for manual evidence, and don't treat ticket statuses as proof.

Read the code at the target when needed to judge behavior. Quote the requirement and cite the code for each finding. Unused abstractions and parameters belong to Standards, not unrequested behavior.

## 6. Mark each finding

Label each finding with a category:

| Category | Blocking | Meaning |
| --- | --- | --- |
| hard violation | yes | Code breaks a documented rule. |
| missing | yes | A requirement in scope is missing or partly done. |
| wrong | yes | Behavior is incorrect or contradicts the requirements. |
| unrequested | yes | Behavior nobody asked for. |
| unresolved earlier finding | yes | An earlier blocking finding is still present. |
| failed verification | yes | A failed command shows unmet behavior. Name the behavior. |
| smell | no | Code matches a smell in the baseline. |
| ticket problem | no | Feature acceptance only. A ticket is inconsistent with the spec or code. |

Blocked verification isn't a finding. It shows only in the verification status.

If a material behavior could reasonably satisfy the spec in more than one way and you need a decision to finish the review, report the question under Clarifications. Don't call an unspecified preference a defect.

## 7. Report

Keep Standards and Spec separate. Check that findings are supported, remove duplicates, and explain priorities or conflicts in the summary. Keep each axis report under 400 words. Use this template:

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

<Clarifications, if a product decision is needed.>

Summary: Standards <n> blocking, <n> not blocking. Spec <n> blocking, <n> not blocking. Verification <status>.

<Key priorities or conflicts, if any.>
```

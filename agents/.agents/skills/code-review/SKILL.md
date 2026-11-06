---
name: code-review
description: Review changes, an implementation ticket, or final feature acceptance along separate Standards and Spec axes. Use when reviewing a branch or work in progress, checking an implementation against its ticket, or verifying a complete feature.
---

# Code review

Keep two assessments separate:

- **Standards**: does the changed code conform to documented standards?
- **Spec**: does the current implementation satisfy the requirements in scope?

Use fresh, parallel subagents for the two axes when supported. If nested subagents are unavailable, review the axes sequentially yourself and keep the findings separate. Do not describe that fallback as independent-context review.

## 1. Identify the mode and inputs

Load `issue-tracker` when reading issues or specs, requesting storage operations for simple fetches. Explicit paths supplied by the caller take precedence over discovery; do not ask for inputs already provided.

| Mode | Inputs and scope |
| --- | --- |
| Change review | A supplied fixed point, plus an issue or spec when available. Review the diff and its scoped requirements. |
| Ticket review | An implementation ticket and its spec. Read `Base commit:` from the ticket and review the entire current ticket implementation, including previous blocking findings. |
| Feature acceptance | The current final acceptance ticket, full spec, and every associated ticket. Read the feature baseline from the acceptance ticket. Check every current story and integration requirement against the current implementation. |

Use the caller's requested mode. Otherwise infer feature acceptance from `Kind: acceptance`, ticket review from an implementation ticket with `Base commit:`, and change review for an ordinary fixed-point comparison.

For ticket review or feature acceptance, request the contract from `issue-tracker` before beginning. Reading the actual feature spec does not require its authoring template.

For discovery in a change review, find the issue from a user-provided path or number, then issue references in commit messages. Use `issue-tracker` to fetch it and its parent spec. Otherwise look for a user-provided spec or a matching spec under `docs/`, `specs/`, or `.scratch/`. If neither exists, ask for them; when the user confirms there are none, skip the Spec axis and report "No spec available."

Read the selected ticket's full comments, not only its latest report. In feature acceptance, also read every associated ticket and require the current acceptance ticket's `Spec revision:` to match the spec's `Revision:`. A previous pass or a set of checked boxes is not evidence that the revised feature works.

## 2. Pin the fixed point

For ticket review and feature acceptance, the selected ticket's `Base commit:` is authoritative. Never replace a missing base with `HEAD`, reconstruct it from commit subjects, or use a different ticket's base. Stop for explicit recovery if the field is missing or resolves to a different commit than a caller-supplied fixed point.

For a change review, use the fixed point the user supplied. If none was supplied, ask for it.

Confirm the reference resolves with `git rev-parse --verify <fixed-point>^{commit}` before starting any review. A persisted ticket base must also be an ancestor of `HEAD`, checked with `git merge-base --is-ancestor <base> HEAD`. If not, stop for recovery rather than silently comparing against another merge-base.

Capture `git diff <fixed-point>...HEAD` and `git log <fixed-point>..HEAD --oneline` once. The ticket baseline remains unchanged across repairs and restarts, so the review includes earlier implementation attempts. Keep ticket findings scoped to the selected ticket even if unrelated changes also fall within the range. Feature acceptance deliberately covers the whole feature diff.

Check `git status --short`. Ralph handoffs require a clean checkout before review; unexpected changes need recovery, not an expanded review scope. For standalone change reviews, include relevant uncommitted changes using `git diff $(git merge-base <fixed-point> HEAD)` and list relevant untracked files for the reviewers to read. Leave unrelated changes out and say so in the report.

An empty diff ends an ordinary change review with "No changes to review." It does not prevent ticket or feature acceptance against the current implementation. Report an empty Standards diff and continue the requirements and verification checks.

## 3. Independently verify ticket or feature behavior

In ticket review, run the ticket's `## Verification` commands, or the spec's Testing decisions commands if the ticket has none. In feature acceptance, run the full feature-level commands from the spec and the acceptance ticket. Read the commands and required evidence in associated human-work tickets too.

Run the commands yourself even when an implementer reported success. Record the exact commands and results. Do not assume a targeted ticket test proves cross-ticket integration. Confirm any required manual checks have explicit human evidence; do not claim to have performed a manual check yourself.

Missing commands, unavailable human evidence, or an environment that prevents verification block completion. A completed command that demonstrates unmet behavior is a finding. Do not change source, tests, snapshots, or the requirements to make checks pass. Unexpected repository changes from verification need recovery; never discard them. Pass the results to the two review axes so they need not repeat the same commands.

For ordinary change reviews, report any checks you ran or that were unavailable without inventing a verification policy.

## 4. Identify standards sources

Find the repo's documented coding rules. Include `CODING_STANDARDS.md` and `CONTRIBUTING.md` whenever present.

The Standards axis always also reads [references/smell-baseline.md](references/smell-baseline.md). Its baseline smells are judgement calls, never hard violations, and documented repo rules override them.

## 5. Review the two axes

Start fresh read-only subagents together in the foreground when possible. Otherwise use the sequential fallback described above. End each subagent prompt with:

> Perform this review directly. Do not load the `code-review` skill, spawn sub-agents, or modify any files.

Give both reviewers the selected mode, fixed point, diff command, commit list, relevant paths, previous blocking findings, and independent verification results. Pass paths rather than copied documents when those paths are available.

### Standards prompt

Include every standards-source path and the absolute path to this skill's `references/smell-baseline.md`, with instructions to read them.

Ask for documented-rule violations and baseline smells, citing the file, hunk, and rule. Distinguish hard violations from judgement calls. Skip anything tooling enforces. In ticket mode, exclude changes belonging only to other tickets. In feature acceptance, assess the full feature diff. Recheck earlier blocking Standards findings even when the latest repair did not touch their code. Keep the report under 400 words.

### Spec prompt

Include the issue and spec paths when available. For feature acceptance, also include all associated ticket paths.

Ask for:

- Missing or partial requirements in scope.
- Behavior that was not requested. Do not include unused abstractions or parameters, which belong to Standards.
- Incorrect behavior or contradictions with the requirements.
- Unresolved earlier blocking findings and failed verification that demonstrates unmet behavior.

The implementation ticket sets the scope in ticket mode. `Covers:` connects it to user stories but does not make one slice responsible for other tickets' deliverables. Read the current implementation beyond changed lines when needed to establish whether the ticket works. A clean repair diff alone cannot establish that an earlier omission was fixed.

In feature acceptance, the complete current spec sets the scope, not the last implementation ticket or just the acceptance ticket's checklist. Check that all current stories work together, approved scope changes are reflected, no unfinished implementation or unresolved blocker remains, and required human evidence exists. Inspect current code and tests even when relevant behavior predates the latest diff. Do not treat ticket statuses as proof of correctness.

Quote the issue or spec requirement for each finding and cite the relevant implementation. Keep the report under 400 words. Skip this axis only in a change review with neither issue nor spec.

## 6. Aggregate

Present the two reports verbatim or lightly cleaned. Do not merge or rerank their findings. Use this template:

```markdown
Reviewed `<fixed-point>...HEAD` in <change review | ticket review | feature acceptance> mode.

## Verification

<commands, results, required human evidence, and any blocker; or "Not run.">

## Standards

<Standards report>

## Spec

<Spec report, or "No spec available.">

**Summary:** Standards: <n> findings, worst: <issue or "none">. Spec: <n> findings, worst: <issue or "none">. Verification: <passed | failed | blocked | not run>.
```

Do not pick a single winner across axes. Following standards cannot compensate for implementing the wrong thing, and satisfying requirements cannot compensate for a hard standards violation.

In Ralph, return these assessments to the reviewer that loaded this skill. That reviewer saves its ticket outcome using `issue-tracker`. The two axis subagents remain read-only. A pass requires successful independent verification, no hard Standards violation, and no missing, wrong, or unrequested behavior in scope. Blocked verification is not a pass.

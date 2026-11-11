# Spec prompt

Fill in the `<...>` fields and send the prompt as written.

```text
Review the Spec axis of a code review. Report whether the code does what its requirements ask. Another reviewer checks code standards, so don't judge those.

Mode: <change review | ticket review | feature acceptance>
Ticket: <feature and ticket number, or "none">
Context: <what the caller gave beyond the ticket, or "none">
Code to review: <the diff command, or the commits to read with git show>
Target: <target commit, or "working tree">
Verification results: <status, and each command and its result>

1. Read the requirements:
   - If there is a ticket, load issue-tracker and read the ticket, its spec, and its comments. In feature acceptance, also read the manual tickets for their evidence.
   - Read any requirement documents in the Context. They add to the ticket's requirements and never replace them.
2. Collect earlier findings from the Context and the ticket's comments: blocking findings, and not-blocking findings that aren't resolved yet. Ignore findings recorded before the ticket's last `revise`.
3. Read the code to review. Read the code at the target too when you need it to tell whether the work does what's asked.
4. Report each finding with one of these categories:
   - missing: a requirement in scope isn't done, or is only partly done.
   - unrequested: behavior nobody asked for. Unused abstractions and parameters belong to the Standards review.
   - wrong: behavior that is incorrect or contradicts the requirements.
   - unresolved earlier finding: an earlier blocking finding is still present.
   - failed verification: a failed command shows unmet behavior. Say which behavior.
   - ticket problem: feature acceptance only. A ticket is inconsistent with the spec or the code.

List each earlier not-blocking finding that is gone as fixed. Don't report one that's still present again.

Scope:
- Change review: the requirements you read in step 1.
- Ticket review: the ticket. Its covered stories don't make it responsible for other tickets' work.
- Feature acceptance: the whole current spec. Check that all stories work together. Read code and tests even when they predate the range. Read tickets only for manual evidence, and don't treat ticket statuses as proof.

Quote the requirement and cite the code for each finding. Keep the report under 400 words.

Perform this review directly. Do not load the code-review skill, spawn subagents, or modify any files.
```

# Standards prompt

Fill in the `<...>` fields and send the prompt as written.

```text
Review the Standards axis of a code review. Report whether the changed code follows the repo's documented standards. Another reviewer checks whether it meets its requirements, so don't judge that.

Mode: <change review | ticket review | feature acceptance>
Ticket: <feature and ticket number, or "none">
Context: <what the caller gave beyond the ticket, or "none">
Code to review: <the diff command, or the commits to read with git show>
Target: <target commit, or "working tree">
Standards sources: <paths, or "none">
Smell baseline: <absolute path to the code-review skill's references/smell-baseline.md>
Verification results: <status, and each command and its result>

1. Read every standards source and the smell baseline.
2. Collect earlier Standards findings from the Context and, if there is a ticket, from its comments: blocking findings, and not-blocking findings that aren't resolved yet. Load issue-tracker to read the ticket. Ignore findings recorded before the ticket's last `revise`.
3. Read the code to review.
4. Report each finding with one of these categories:
   - hard violation: the code breaks a documented rule.
   - smell: the code matches a smell in the baseline. Documented rules override the baseline.
   - unresolved earlier finding: an earlier blocking Standards finding is still present, even if the latest change didn't touch that code.

List each earlier not-blocking finding that the code no longer has as fixed. Don't report one that's still present again.

Review only the code listed above. Skip anything tooling enforces. Cite the file, the hunk, and the rule for each finding. Keep the report under 400 words.

Perform this review directly. Do not load the code-review skill, spawn subagents, or modify any files.
```

# Reviewer prompt

Fill in the `<...>` fields and send the prompt as written.

```text
You are reviewing one ticket. Report what you find. Another agent records the outcome on the ticket.

Feature: <feature slug>
Ticket: <ticket number>
Commit filter: <`^<NN>: ` for an implementation ticket, or "none" for an acceptance ticket>

1. Load issue-tracker and use it to read the ticket. Note its acceptance criteria.
2. Load code-review and run it with this ticket and commit filter. It has everything it needs, so don't ask for more inputs.
3. Match unmet criteria and newly fixed findings to their entries on the ticket.

Rules:
- Don't edit files, commit, push, or switch branches. Use issue-tracker only to read.
- Don't load the ralph skill.

Return the code-review results as one JSON object. Use criterion IDs from `ticket get` and quote fixed findings exactly as the ticket records them. Keep the Standards and Spec results separate.

{"verification_status":"passed", "verification":["npm test passed"], "evidence":[], "unmet_criteria":[], "fixed_findings":[], "standards":{"blocking":[], "not_blocking":[]}, "spec":{"blocking":[], "not_blocking":[]}, "questions":[], "summary":""}

Put Clarifications in `questions`. Keep categories in findings and give the reason when verification is blocked.
```

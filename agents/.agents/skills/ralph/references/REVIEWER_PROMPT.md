# Reviewer prompt

Fill in the `<...>` fields and send the prompt as written.

```text
You are reviewing one ticket. Report what you find. Another agent records the outcome on the ticket.

Feature: <feature slug>
Ticket: <ticket number>
Commit filter: <`^<NN>: ` for an implementation ticket, or "none" for an acceptance ticket>

1. Load issue-tracker and fetch the ticket. Note its acceptance criteria.
2. Load code-review and run it with this ticket and commit filter. It has everything it needs, so don't ask for more inputs.
3. List each acceptance criterion that the review shows isn't met.
4. List each unresolved `Not blocking:` finding in the ticket's comments that the review reports as fixed.

Rules:
- Don't edit files, commit, push, or switch branches.
- Don't load the ralph skill.

Reply in exactly this format, with nothing before or after it. Quote each criterion and finding exactly as the ticket writes it.

Unmet criteria: <each unmet criterion, one per line, or "none">
Fixed findings: <each fixed not-blocking finding, one per line, or "none">

<the code-review report as written>
```

# Reviewer Prompt

Fill in the `<...>` fields and send the below prompt as written. Pass paths, not copied document contents or commit SHAs.

```text
You are reviewing one ticket in a Ralph loop. You independently verify the work; Ralph does not run its checks for you.

Spec: <spec path>
Ticket: <ticket path>
Glossary: <glossary path or "none">

1. Load issue-tracker and fetch the ticket.
2. Read the spec and glossary when present.
3. Read the entire ticket. Stop and notify Ralph if any of the following is true in the ticket file:
    - Its `Status` field isn't `needs-review`
    - Its `Blocked by` tickets aren't labled as `done`
    - Its `Kind` field is `acceptance` and its `Spec revision` doesn't match the spec files revision
    - Its `Kind` field is `acceptance` and its `Base commit` is empty and there are no previous `acceptance` tickets with populated `Base commit`s
    - Its `Attempt` is greater than `Attempt limit`
    - An already-started ticket has no base 
    - Check that it holds only one deliverable: one behaviour that can be checked end to end on its own
3. Load code-review. Use ticket review mode for Kind: implementation, or feature acceptance mode for Kind: acceptance. 
    3a. In feature acceptance mode, read the full spec and all feature tickets. Review the current implementation against every current story, including cross-ticket behavior. Use the acceptance ticket's feature baseline for the Standards diff, not the last implementation ticket's base.
4. Independently run the verification required by code-review. Recheck unresolved blocking findings and the entire current ticket or feature behavior, not just the newest repair diff. If nested subagents are unavailable, review Standards and Spec sequentially and keep their findings separate.
5. Pass the ticket only when verification passed, no hard Standards violation remains in scope, and no scoped requirement is missing, wrong, or unrequested. Otherwise send it back. If verification or review cannot be completed, record a blocked review instead of claiming a pass.
6. Save the outcome and persist the relevant data in the ticket. List blocking findings and command results. Put code smells and other judgement calls as comments with the "Not Blocking: " prefix. For failed feature acceptance, describe the needed follow-up work but do not create or edit other tickets.

Rules:
- Don't edit code, commit, push, or switch branches.
- Don't load the ralph skill.

Reply with a short summary and the path to the ticket.
```

# Implementer prompt

Fill in the `<...>` fields and send the below prompt as written. Pass paths, not copied document contents or commit SHAs.

```text
You are implementing one ticket in a Ralph loop. Ensure you adhere to the rules and conventions declared by the ticket. A fresh subagent will review your work after you finish.

Spec: <spec path>
Ticket: <ticket path>
Glossary: <glossary path or "none">

1. Load issue-tracker and request the ticket implementation contract, including metadata, baselines, attempts, lifecycle transitions, edit permissions, and persistence.
2. Fetch the ticket and read the spec, glossary when present, and the entire ticket, including all comments.
3. Validate its metadata and attempt history under that contract. Stop and notify Ralph if any of the following is true:
    - Its `Kind` is not `implementation` or its `Mode` is not `afk`.
    - Its `Status` is neither `open` nor `in-progress`.
    - Any blocker is not `done`.
    - Execution metadata is invalid under the contract, including a missing base on an already-started ticket or no budget for a required new attempt.
    - The ticket contains more than one independently verifiable deliverable.
4. Follow the contract to claim the ticket and start or resume its attempt. Persist the initialization checkpoint before editing source. Implement only its deliverable and address unresolved blocking findings from all previous reviews.
5. Run the ticket's Verification commands, or the spec's Testing decisions commands if the ticket has none. If neither names commands, stop rather than inventing a verification policy.
6. Tick only criteria you actually met, record commands and results, and follow the contract's review-handoff lifecycle. Commit the source changes and persist the ticket before handoff. For tracked local tickets, commit code and the ticket update together. Prefix commit subjects with the ticket number, such as "03: add order totals".

Rules:
- Don't push, rewrite history, or switch branches.
- If you cannot finish, update the ticket status and record why and what must happen before resuming. Persist safe partial work without hiding failures. If you cannot make a safe committed handoff, leave the changes for deliberate recovery and report them. Never discard work to obtain a clean checkout.
- Don't load the ralph skill.

Reply with a short summary and the path to the ticket.
```

# Implementer prompt

Fill in the `<...>` fields and send the prompt as written.

```text
You are implementing one ticket. Another agent reviews your work after you finish.

Feature: <feature slug>
Ticket: <ticket number>

1. Load issue-tracker and use it to read the ticket and its spec, including the ticket's current comments. Read GLOSSARY.md, relevant ADRs, and CODING_STANDARDS.md if they exist.
2. Check that you can start. If you can't, change nothing and reply `precondition failed`. You can't start when:
   - the ticket is unclear, or contradicts the spec, and you'd have to guess what it wants, or
   - its Verification section names no commands.
3. Find this ticket's earlier commits with `git log --grep='^<NN>: ' BASE..HEAD`, where BASE is the ticket's `Base commit:`. Build on them. Fix every blocking finding in this ticket's comments that is still unresolved. Act only on this ticket's comments, never on another ticket's.
4. Implement only what the ticket asks. Leave its Out of scope items alone.
5. Run every command in the ticket's Verification section.
6. Commit your work. Start every commit subject with `<NN>: `, such as `<NN>: add order totals`. Leave no uncommitted changes, and never discard work.

Reply `ready for review` only when you met every acceptance criterion and every verification command passes. If you can't get there, commit what you have and reply `stuck`.

Rules:
- Never edit tickets or the spec. Use issue-tracker only to read.
- Never commit secrets, such as API keys, tokens, passwords, or private keys, in code or commit messages.
- Don't push, rewrite history, or switch branches.
- Don't load the ralph skill.

Return one JSON object. Use criterion IDs from `ticket get`, not the criterion text.

{"outcome":"ready for review", "criteria_met":["C01"], "verification":["npm test passed"], "notes":[], "reason":""}

Use `stuck` or `precondition failed` and give a reason when you can't hand off. List each verification command and its result, including failures or skips. Use `notes` for evidence the reviewer can't reproduce.
```

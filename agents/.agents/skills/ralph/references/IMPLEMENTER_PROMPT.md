# Implementer prompt

Fill in the `<...>` fields and send the prompt as written.

```text
You are implementing one ticket. Another agent reviews your work after you finish.

Feature: <feature slug>
Ticket: <ticket number>

1. Load issue-tracker. Fetch the ticket, including every comment, and the feature's spec. Read GLOSSARY.md, relevant ADRs, and CODING_STANDARDS.md if they exist.
2. Check that you can start. If you can't, change nothing and reply `precondition failed`. You can't start when:
   - the ticket is unclear, or contradicts the spec, and you'd have to guess what it wants, or
   - its Verification section names no commands.
3. Find this ticket's earlier commits with `git log --grep='^<NN>: ' BASE..HEAD`, where BASE is the ticket's `Base commit:`. Build on them. Fix every blocking finding in this ticket's comments that is still unresolved. Act only on this ticket's comments, never on another ticket's.
4. Implement only what the ticket asks. Leave its Out of scope items alone.
5. Run every command in the ticket's Verification section.
6. Commit your work. Start every commit subject with `<NN>: `, such as `<NN>: add order totals`. Leave no uncommitted changes, and never discard work.

Reply `ready for review` only when you met every acceptance criterion and every verification command passes. If you can't get there, commit what you have and reply `stuck`.

Rules:
- Never edit tickets or the spec.
- Never commit secrets, such as API keys, tokens, passwords, or private keys, in code or commit messages.
- Don't push, rewrite history, or switch branches.
- Don't load the ralph skill.

Reply in exactly this format, with nothing before or after it. Quote each criterion exactly as the ticket writes it.

Outcome: <ready for review | stuck | precondition failed>
Criteria met: <each acceptance criterion you met, one per line, or "none">
Verification: <each command and its result, one per line>
Reason: <why you stopped, for stuck or precondition failed, or "none">
```

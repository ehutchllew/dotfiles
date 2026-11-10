---
name: to-standards
description: Create or update the repo's CODING_STANDARDS.md from its configs, docs, and code. Use when the user asks for coding standards, or when a repo has none.
disable-model-invocation: true
metadata:
    opencode/autoinvoke: false
---

# To standards

Write the repo's coding rules to `CODING_STANDARDS.md` at the repo root. The file holds only rules that tooling doesn't enforce. Each rule is mandatory, so leave out soft preferences.

This skill reads code but changes only `CODING_STANDARDS.md`.

## Process

1. Read what exists:
   - `CODING_STANDARDS.md`, if present. Update it instead of starting over.
   - Lint, formatter, type checker, and editor configs.
   - `CONTRIBUTING.md`, ADRs, and other docs that state coding rules.
   - A few representative source files from each main area, and their tests.
2. Draft the rules:
   - Leave out anything a config already enforces.
   - Write only rules the code already follows or the user states. Don't invent preferences.
   - When the code follows two patterns for the same thing, don't pick one. List both as a question for the user.
   - Write one rule per line. Add a short example only when the rule is unclear without one.
   - Group rules under topic headings.
   - Put every rule about tests in one `## Testing` section, and always include it.
3. If the repo shows no testing rules, propose these for the Testing section:
   - Test external behavior, not implementation details.
   - Test at public entry points and module boundaries, checking their inputs and outputs.
   - Don't test private helpers, trivial code, or third-party libraries.
   - Add a focused lower-level test only where it gives clearer failures.
4. Show the user the draft and your questions. For an update, show only the changes and any new questions. Repeat until the user approves.
5. Write the file.

## Format

```markdown
# Coding standards

Tooling enforces <the tools and what they cover>. The rules below cover the rest.

## <Topic>

- <rule>

## Testing

- <rule>
```

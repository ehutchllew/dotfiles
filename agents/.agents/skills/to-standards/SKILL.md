---
name: to-standards
description: Create or update the repo's CODING_STANDARDS.md from its configs, docs, and code. Use when the user asks for coding standards, or when a repo has none.
disable-model-invocation: true
metadata:
    opencode/autoinvoke: false
---

# To standards

Write the repo's coding and testing rules to `CODING_STANDARDS.md` at the repo root. The file holds only rules that tooling doesn't enforce, plus the commands that run the tooling. Each rule is mandatory, so leave out soft preferences.

This skill reads code but changes only `CODING_STANDARDS.md`. Replace secrets, such as API keys, tokens, passwords, and private keys, with `[REDACTED]` in anything you write to it.

## Process

1. Read what exists:
   - `CODING_STANDARDS.md`, if present. Update it instead of starting over.
   - Lint, formatter, type checker, and editor configs.
   - `CONTRIBUTING.md`, ADRs, and other docs that state coding or testing rules.
   - A few representative source files from each main area, and their tests.
2. If the conversation has a summary with decisions, assumptions, out-of-scope items, or open risks, build from it:
   - Decisions are rules the user stated.
   - Out-of-scope items are rules to leave out.
   - If an assumption would add or change a rule, confirm it with the user first.
3. Find the checks:
   - List each tool the repo uses to enforce code rules, such as a linter, formatter, or type checker. Include tools built into the language, such as `gofmt` and `go vet`, even when no config file names them.
   - For each tool, find the exact command that checks the whole repo and fails when the check fails. If the repo doesn't show one, propose one.
4. Draft the rules:
   - Leave out anything a listed check already enforces.
   - Write only rules the code or tests already follow, or the user states. Don't invent preferences.
   - When the code or tests follow two patterns for the same thing, don't pick one. List both as a question for the user.
   - Write one rule per line. Add a short example only when the rule is unclear without one.
   - Write each rule so a reviewer can tell whether a given line breaks it. Name what it applies to, such as exported fields or every identifier, when the code doesn't make that obvious.
   - Group rules under topic headings.
   - Put every rule about tests in one `## Testing` section, and always include it.
5. If neither the docs nor the tests show any testing rules, propose these for the Testing section. They aim for the fewest tests that would catch a regression a caller would notice:
   - Test each behavior through the narrowest stable interface that owns it, such as a package API, an HTTP endpoint, or a CLI command. Supply inputs and assert on returned values, errors, or state changes. Use a broader test only when the behavior depends on components working together.
   - Don't repeat the same cases at another layer unless the broader test checks something different.
   - Cover normal use, meaningful limits and failures, and known regressions. Add rarer cases only when the spec requires them or a failure would be costly. Don't add a case just because it's possible.
   - Don't test private helpers, trivial code such as getters and constants, or third-party libraries.
   - Don't assert on details the behavior doesn't depend on, such as field order, whitespace, or log wording.
   - Use parameterized or table-driven tests when cases run the same steps and differ only in inputs or expected results. Keep cases with different setup or behavior in separate tests.
6. Show the user the draft and your questions. For an update, show only the changes and any new questions. Repeat until the user approves.
7. Write the file. Commit only this file with the subject `standards: update`. Skip the commit if git ignores the file.

## Format

```markdown
# Coding standards

The checks below enforce what tooling covers. The rules below them cover the rest.

## Checks

- `<command>`: <what it enforces>

## <Topic>

- <rule>

## Testing

- <rule>
```

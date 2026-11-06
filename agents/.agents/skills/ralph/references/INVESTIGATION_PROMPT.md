# Recovery investigation prompt

Fill in the `<...>` fields and send the below prompt as written. Pass paths, not copied document contents or commit SHAs.

```text
You are investigating an incomplete or inconsistent handoff in a Ralph loop. This is read-only recovery diagnosis, not implementation or formal review. Normal scheduling is paused.

Spec: <spec path>
Ticket: <ticket path>
Feature tickets: <ticket directory or explicit ticket paths>
Reason: <missing response or observed handoff problem>

1. Load issue-tracker and request the ticket recovery investigation rules, including baselines, attempts, lifecycle transitions, and persistence.
2. Read the spec, selected ticket and its full comments, and related tickets as needed.
3. Inspect the saved artifacts, working tree status, and relevant git history or diffs using read-only operations. A dirty checkout or missing checkpoint is allowed for this investigation only.
4. Identify what was saved, what remains incomplete or inconsistent, and whether the saved state satisfies the normal handoff contract. Separate observed facts from uncertainties. Commit messages and a missing response are not proof that implementation or verification succeeded.
5. Recommend the next recovery action and identify any decision or missing evidence that requires the user. Report a missing baseline as a blocker; do not infer a replacement. Your report does not authorize another attempt or establish a passed review.

Rules:
- Do not edit any file, including tickets, or run commands that change the checkout or git history. Do not commit, push, or switch branches.
- Do not run tests or other verification commands, repair metadata, or start, finish, or consume an attempt.
- Do not retry denied actions or bypass permission decisions.
- Do not load code-review or ralph, or spawn another subagent.

Reply with the ticket path, observed ticket and git state, supporting file references, uncertainties, and the recommended recovery action. Do not claim that work passed review.
```

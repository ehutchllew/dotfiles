# 01: Show counts

Kind: implementation
Phase: implement
Status: open
Attempt: 1
Covers: US 1
Base commit: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
Target commit: bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb

## What to build

Show the item count in the order summary.

## Acceptance criteria

- [ ] An empty order shows zero.
- [ ] The count equals the sum of quantities.

## Verification

Run `npm test`.

## Out of scope

- Order totals.

## Comments

send back (attempt 1):
- Unmet criteria:
  - "An empty order shows zero."
  - "The count equals the sum of quantities."
- Blocking (wrong): output differs:
  ```text
  ok  	pkg	0.1s
  ```
- Verification:
  - ```text
    ok  	pkg	0.1s
    ```
  - `go test` output:
    ```text
    ok  	pkg	0.1s
    ```
- Not blocking: Rename `n`.
- Not blocking: Split the long function.
- Note:
  ```text
  ok  	pkg	0.1s
  ```

stop (attempt 1):
- Reason:
  Blocked:
  ```text
  ok  	pkg	0.1s
  ```
- Evidence:
  - ```text
    ok  	pkg	0.1s
    ```

comment:
- Not blocking resolved: Rename `n`. Ignored.
- Not blocking resolved: Split the long function. Moved to ticket 02.

resume (attempt 1):
- Resolution: Fixed the database.

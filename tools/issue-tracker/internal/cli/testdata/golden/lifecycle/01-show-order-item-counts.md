# 01: Show order item counts

Kind: implementation
Phase: review
Status: done
Attempt: 2
Covers: US 1
Base commit: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
Target commit: cccccccccccccccccccccccccccccccccccccccc

## What to build

Show the total item quantity in the order summary.

## Acceptance criteria

- [x] An empty order shows an item count of zero.
- [x] The item count equals the sum of the order's item quantities.

## Verification

Run `npm test`.

## Out of scope

- Order totals.

## Comments

hand off (attempt 1):
- Verification:
  - `npm test` passed.

send back (attempt 1):
- Unmet criteria: "An empty order shows an item count of zero."
- Blocking (missing): an empty order shows no item count line.
- Verification:
  - `npm test` passed.
- Not blocking: the local `n` doesn't say it holds the item count.

hand off (attempt 2):
- Verification:
  - `npm test` passed.

pass (attempt 2):
- Verification:
  - `npm test` passed.
- Not blocking resolved: the local `n` doesn't say it holds the item count. Fixed.

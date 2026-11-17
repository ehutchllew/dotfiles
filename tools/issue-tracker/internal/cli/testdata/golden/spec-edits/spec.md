# Checkout counts

Revision: 3
Attempt limit: 5

## Problem statement

Shoppers can't see how many items they ordered.

## Solution

Show item counts in the summary.

## User stories

US 1: As a shopper, I want the item count, so that I know what I ordered

US 2: Withdrawn in revision 3.

US 3: As a shopper, I want totals, so that I can pay

## Implementation decisions

- Add a count to the `summary` module.

## Testing decisions

- Run `npm test`.
- A manual check on a phone.

## Out of scope

- Order totals.

## Change history

- Revision 1: Initial spec.
- Revision 2: US 3 added. US 1 reworded.
- Revision 3: US 2 withdrawn.

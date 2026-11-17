# Checkout counts

Revision: 1
Attempt limit: 3

## Problem statement

Shoppers can't see how many items they ordered.

## Solution

Show item counts in the order summary, on desktop and on phones.

## User stories

US 1: As a shopper, I want to see the item count, so that I know what I ordered

US 2: As a shopper, I want to see the count on my phone, so that I can check on the go

## Implementation decisions

- Add a count to the `summary` module.

## Testing decisions

- Run `npm test`.
- A manual check on a phone.

## Out of scope

- Order totals.

## Change history

- Revision 1: Initial spec.

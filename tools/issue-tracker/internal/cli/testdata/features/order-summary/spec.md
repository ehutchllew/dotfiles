# Order summary

Revision: 1
Attempt limit: 3

## Problem statement

Shoppers can't tell how many items are in an order before they pay.

## Solution

The order summary shows the total item count.

## User stories

US 1: As a shopper, I want to see the item count in the order summary, so that I know what I'm paying for

## Implementation decisions

- The `ordersummary` module adds an item count. From a prototype:

```ts
type Summary = { items: number }
```

## Testing decisions

- Unit tests for the `ordersummary` module.
- Run `npm test`.
- A manual check on a phone.

## Out of scope

- Order totals.

## Change history

- Revision 1: Initial spec.

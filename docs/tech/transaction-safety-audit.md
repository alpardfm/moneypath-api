# Transaction Safety Audit

This document reviews the financial transaction safety guarantees currently implemented.

## Covered Flows

- create mutation
- update mutation
- debt-aware mutation create
- debt-aware mutation update

## Current Guarantees

- wallet changes and debt changes happen inside a single PostgreSQL transaction
- wallet rows are locked with `FOR UPDATE` before balance-changing writes
- debt rows are locked with `FOR UPDATE` before debt-changing writes
- mutation update uses rollback-and-reapply logic in one transaction
- outgoing wallet deductions are guarded with `balance >= amount`
- outgoing debt payments are guarded with `remaining_amount >= amount`
- mutation delete remains disallowed to preserve immutable financial history

## Important Edge Cases

- editing a mutation from one wallet to another locks both wallets before reapply
- reversing a `borrow_new` mutation is rejected if the generated debt is already referenced by another mutation
- debt reactivation clears `deleted_at` and sets `is_active = true` when a mutation increases remaining debt again

## Residual Notes

- wallet balance is still stored and mutated directly for runtime efficiency, but only through mutation flows
- `wallet`, `debt`, and `mutation` list queries are read-only and do not affect consistency guarantees
- future phases may add stronger audit fields such as actor id or request id persistence if needed

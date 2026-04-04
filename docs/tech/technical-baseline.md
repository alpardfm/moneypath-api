# Technical Baseline

This document captures the current technical baseline that matches the latest product logic.

## Core Business Modules

- `auth`: register, login, token validation, password change
- `profile`: current user profile read and update
- `wallet`: wallet master data and activation rules
- `debt`: debt master data and remaining debt state
- `mutation`: source of truth for all balance-changing events
- `dashboard`: derived overview for daily use
- `summary`: derived aggregates for reporting periods

## Shared Technical Components

- `internal/app`: startup lifecycle and dependency wiring
- `internal/config`: configuration loading
- `internal/platform/database`: PostgreSQL connection management
- `internal/platform/logger`: structured logging
- `internal/http/response`: response envelope helpers
- `internal/http/middleware`: request logging and future auth middleware

## Current API Baseline

- transport style: REST API
- router: `chi`
- database: PostgreSQL
- data access: repository pattern per module
- business logic: service/use-case layer per module
- list endpoints: paginated with shared `meta`
- mutation history: supports filter and sort query params

## Naming Direction

- table names: plural `snake_case`
- JSON fields: `snake_case`
- mutation types: `masuk`, `keluar`
- debt paid status: `lunas`
- ownership key: `user_id`

## Schema Direction

The MVP schema is centered on:

- `users`
- `wallets`
- `debts`
- `mutations`

Supporting fields should emphasize:

- ownership boundaries
- balance safety
- debt remaining amount
- immutable financial history

See [phase-1-schema.md](/Users/alpardfm/Documents/Coding/Learn/moneypath-api/docs/tech/phase-1-schema.md) for the ERD, soft delete strategy, and Phase 1 schema constraints.

## Notes

- Wallet balance must always be derived from controlled mutation writes.
- Debt remaining amount must only change through debt-aware mutation flows.
- Dashboard and summary remain derived modules, not source-of-truth modules.
- Wallets and debts may be hidden via soft delete while preserving historical references.
- See [api-hardening.md](/Users/alpardfm/Documents/Coding/Learn/moneypath-api/docs/tech/api-hardening.md) for pagination, validation, and response consistency rules.
- See [transaction-safety-audit.md](/Users/alpardfm/Documents/Coding/Learn/moneypath-api/docs/tech/transaction-safety-audit.md) for current transaction guarantees.

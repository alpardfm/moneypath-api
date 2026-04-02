# Personal Finance API

A backend project for tracking personal cash flow, wallets, debts, budgets, and financial summaries.

Built as a practical system to manage income, expenses, obligations, and financial discipline in a structured way.

## Goals

- Record income and expenses clearly
- Separate money by wallet/account
- Track debts and installment progress
- Monitor monthly cash flow
- Build a foundation for budgeting and financial insights

## Core Features

- Wallet management
- Income & expense transactions
- Categories
- Debt tracking
- Monthly summary
- Budget overview

## Planned Features

- Recurring transactions
- Installment schedule tracking
- Budget alerts
- Financial health summary
- Cash flow projection
- Export report

## Tech Stack

- Go
- PostgreSQL
- REST API
- Docker
- `chi` router
- `pgx` PostgreSQL driver
- `golang-migrate` compatible SQL migrations
- Optional: Swagger / Makefile / CI

## Project Scope

This project is focused on backend fundamentals:
- clean API design
- good domain structure
- predictable business logic
- maintainable codebase

Frontend is not the priority for now.

## Main Entities

- Users
- Wallets
- Categories
- Transactions
- Debts
- Budget Plans
- Monthly Summaries

## Example Use Cases

- Add salary as income
- Record daily spending
- Track installment payments
- Monitor monthly obligations
- Compare spending against budget
- View summary by month and category

## Future Direction

This project can grow into:
- a personal financial operating system
- a budgeting assistant backend
- a simulation tool for debt payoff and buffer growth

## Status

In progress. Built step by step, improved when needed, shipped when ready.

## Recommended Project Structure

See [docs/project-structure.md](/Users/alpardfm/Documents/Coding/Learn/moneypath-api/docs/project-structure.md) for the initial folder layout used as the base for this project.

## Phase 1 Setup

Current foundation choices:

- Router: `chi`
- Config: environment variables via `internal/config`
- Database: PostgreSQL via `pgxpool`
- Migrations: SQL files in `migrations/` compatible with `golang-migrate`

## Environment

Copy `.env.example` and adjust the values for your local setup:

```env
APP_ENV=development
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/moneypath?sslmode=disable
```

## Run the API

```bash
go run ./cmd/api
```

The API will start on `http://localhost:8080` by default.

## Health Check

```bash
curl http://localhost:8080/health
```

Expected response when the database is reachable:

```json
{
  "status": "ok",
  "data": {
    "service": "moneypath-api",
    "database": "up"
  }
}
```

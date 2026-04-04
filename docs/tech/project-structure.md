# Technical Project Structure

This document describes the codebase layout for the current Moneypath MVP direction.

## Structure

```text
moneypath-api/
├── cmd/
│   └── api/
│       └── main.go
├── docs/
│   ├── logic/
│   │   ├── PROMPT.md
│   │   └── TODO.md
│   └── tech/
│       ├── project-structure.md
│       └── technical-baseline.md
├── internal/
│   ├── app/
│   ├── config/
│   ├── http/
│   │   ├── handler/
│   │   ├── middleware/
│   │   └── response/
│   ├── module/
│   │   ├── auth/
│   │   ├── dashboard/
│   │   ├── debt/
│   │   ├── mutation/
│   │   ├── profile/
│   │   ├── summary/
│   │   └── wallet/
│   └── platform/
│       ├── database/
│       └── logger/
├── migrations/
├── scripts/
├── .env.example
├── go.mod
└── README.md
```

## Responsibility Split

- `docs/logic`: product direction, business rules, implementation roadmap.
- `docs/tech`: technical decisions that implement the product direction.
- `internal/app`: application bootstrap and dependency wiring.
- `internal/config`: environment-driven runtime config.
- `internal/http`: shared HTTP transport concerns.
- `internal/module`: business features grouped by module.
- `internal/platform`: infrastructure adapters such as database and logger.

## Module Strategy

The project uses a module-first layout for business code:

- `auth`
- `profile`
- `wallet`
- `debt`
- `mutation`
- `dashboard`
- `summary`

Each module can grow with its own handler, service, repository, and entity files without forcing unrelated features into one shared domain folder.

## Why This Fits The Current Product

The current MVP is centered on:

- authenticated ownership
- wallet balance trust
- debt tracking
- mutation-driven financial state

Because of that, business code should be organized around those workflows instead of older generic buckets like global `domain`, `repository`, and `service` folders.

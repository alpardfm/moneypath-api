# Project Structure

This project uses a simple layered structure so the codebase stays easy to grow.

## Recommended Layout

```text
moneypath-api/
├── cmd/
│   └── api/
│       └── main.go
├── docs/
│   └── project-structure.md
├── internal/
│   ├── app/
│   ├── config/
│   ├── domain/
│   ├── http/
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── response/
│   │   └── router.go
│   ├── platform/
│   │   ├── database/
│   │   └── logger/
│   ├── repository/
│   └── service/
├── migrations/
├── scripts/
├── README.md
└── todo.md
```

## Folder Purpose

- `cmd/api`: application entrypoint.
- `internal/app`: bootstrap and dependency wiring.
- `internal/config`: environment and application configuration.
- `internal/domain`: core business entities and rules.
- `internal/http`: HTTP router, handlers, middleware, and response helpers.
- `internal/platform`: external infrastructure such as database and logger setup.
- `internal/repository`: persistence layer implementation.
- `internal/service`: business use cases and orchestration logic.
- `migrations`: SQL schema migrations.
- `scripts`: helper scripts for local development.
- `.env.example`: baseline environment variables for local setup.

## Suggested Growth Pattern

As the project grows, the `internal/domain`, `internal/repository`, and `internal/service` folders can be split by feature:

- `wallet`
- `category`
- `transaction`
- `debt`
- `budget`
- `summary`

That keeps each feature close to its business logic while still preserving clear layers.

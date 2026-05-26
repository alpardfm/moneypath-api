# Moneypath API

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat-square&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)
[![CI](https://github.com/alpardfm/moneypath-api/actions/workflows/ci.yml/badge.svg)](https://github.com/alpardfm/moneypath-api/actions/workflows/ci.yml)

Personal finance REST API for tracking cash flow, wallets, debts, and budgets. Built with Go, PostgreSQL, and Clean Architecture principles.

> 🚀 **Live:** Deployed and used daily as a real personal finance tool.

---

## Features

- **Authentication** — JWT-based register/login with secure password hashing
- **Wallet Management** — Multiple wallets with balance integrity enforcement
- **Mutation Tracking** — All financial events (income/expense) as immutable records
- **Debt Management** — Track debts with payment history via mutations
- **Dashboard & Summary** — Aggregated financial overview
- **Categories** — Organize transactions by category
- **Recurring Rules** — Scheduled recurring transactions
- **Health Score** — Financial health assessment
- **Export** — Data export capabilities

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    HTTP Layer                         │
│  Router → Middleware (Auth, CORS) → Handler          │
├─────────────────────────────────────────────────────┤
│                  Module Layer                         │
│  auth │ wallet │ mutation │ debt │ dashboard │ ...   │
│  Each module: handler → service → repository         │
├─────────────────────────────────────────────────────┤
│                 Platform Layer                        │
│           database │ logger │ config                  │
├─────────────────────────────────────────────────────┤
│                  PostgreSQL                           │
└─────────────────────────────────────────────────────┘
```

**Design Principles:**
- Module-first layout — each feature is self-contained
- Mutation is the source of truth — all balance changes go through mutations
- Wallet balance integrity — no direct edits, no negative balances
- User data isolation — strict ownership enforcement

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25 |
| Router | chi/v5 |
| Database | PostgreSQL 16 |
| Migrations | golang-migrate |
| Auth | JWT (golang-jwt/v5) |
| Docs | Swagger (swaggo) |
| Linter | golangci-lint |
| Container | Docker + docker-compose |
| CI/CD | GitHub Actions → VPS deploy |

---

## Quick Start

### Prerequisites
- Go 1.25+
- Docker & Docker Compose
- Make

### 1. Clone & Setup
```bash
git clone https://github.com/alpardfm/moneypath-api.git
cd moneypath-api
cp .env.example .env
```

### 2. Start Database
```bash
make compose-up
```

### 3. Run Migrations
```bash
make migrate-up
```

### 4. Run Server
```bash
make run
```

The API will be available at `http://localhost:8080`.

### Using Docker (Full Stack)
```bash
docker compose up -d
```

---

## API Documentation

Swagger UI is available at `/swagger/` when the server is running.

A Postman collection is also available at [`docs/tech/moneypath-api.postman_collection.json`](docs/tech/moneypath-api.postman_collection.json).

---

## Project Structure

```
moneypath-api/
├── cmd/
│   ├── api/          # Application entrypoint
│   └── migrate/      # Migration CLI
├── internal/
│   ├── app/          # Bootstrap & dependency wiring
│   ├── config/       # Environment-driven config
│   ├── http/         # Router, middleware, handlers, response
│   ├── module/       # Business features (auth, wallet, mutation, debt, ...)
│   └── platform/     # Infrastructure (database, logger)
├── migrations/       # SQL migration files
├── scripts/          # Deploy & utility scripts
├── docs/             # Technical & product documentation
├── Dockerfile        # Multi-stage production build
├── docker-compose.yml
└── Makefile
```

---

## Available Commands

```bash
make run              # Run the API server
make test             # Run all tests
make lint             # Run golangci-lint
make fmt              # Format code
make compose-up       # Start PostgreSQL container
make compose-down     # Stop containers
make migrate-up       # Apply migrations
make migrate-down     # Rollback migrations
make prod-up          # Deploy production stack
make smoke-test       # Run smoke tests
```

---

## Business Rules

- **Wallet balance cannot go negative** — outgoing mutations are rejected if insufficient
- **All balance changes go through mutations** — no direct wallet edits
- **Debt state is mutation-driven** — payments reduce remaining debt via mutations
- **User data isolation** — strict ownership enforcement on all resources
- **Mutation is immutable** — can be edited but never deleted

---

## Deployment

The project includes a production deployment pipeline:

1. Push to `master` triggers GitHub Actions
2. Code is synced to VPS via rsync
3. Docker containers are rebuilt and restarted
4. Migrations run automatically on startup

See [`docs/tech/deployment.md`](docs/tech/deployment.md) for details.

---

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

# Deployment Guide

This document describes the current production-like deployment flow for `moneypath-api`.

## Files

- [Dockerfile](/Users/alpardfm/Documents/Coding/Learn/moneypath-api/Dockerfile)
- [docker-compose.prod.yml](/Users/alpardfm/Documents/Coding/Learn/moneypath-api/docker-compose.prod.yml)
- [.env.production.example](/Users/alpardfm/Documents/Coding/Learn/moneypath-api/.env.production.example)
- [docker-entrypoint.sh](/Users/alpardfm/Documents/Coding/Learn/moneypath-api/scripts/docker-entrypoint.sh)
- [smoke-test.sh](/Users/alpardfm/Documents/Coding/Learn/moneypath-api/scripts/smoke-test.sh)

## Environment Preparation

Copy the production example and replace secrets:

```bash
cp .env.production.example .env.production
```

Required variables:

- `APP_ENV`
- `PORT`
- `DATABASE_URL`
- `JWT_SECRET`
- `AUTO_MIGRATE`
- `MIGRATIONS_PATH`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `API_PORT`
- `POSTGRES_PORT`

`DATABASE_URL` must use the same username, password, and database name as the PostgreSQL variables in the same env file.

## Production-like Local Deployment

Bring up the stack:

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build
```

Stop the stack:

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml down
```

## Migration Execution

The API container entrypoint runs migrations before starting the API when:

- `AUTO_MIGRATE=true`

Migration source defaults to:

- `file:///app/migrations`

## Smoke Test

After the stack is healthy, run:

```bash
BASE_URL=http://localhost:18080 ./scripts/smoke-test.sh
```

The smoke test verifies:

- health
- register
- login
- wallet create and list
- debt create
- incoming mutation
- outgoing debt payment mutation
- dashboard
- summary

## Notes

- This is a production-like deployment flow, not yet a cloud-specific deployment target.
- For a real public deployment, keep `JWT_SECRET` and database credentials in secret storage, not committed files.

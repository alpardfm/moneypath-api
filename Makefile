APP_NAME := moneypath-api
PROD_ENV_FILE ?= .env.production
PROD_COMPOSE := docker compose --env-file $(PROD_ENV_FILE) -f docker-compose.prod.yml

.PHONY: run test fmt lint compose-up compose-down migrate-up migrate-down migrate-version prod-up prod-down prod-logs smoke-test

run:
	go run ./cmd/api

test:
	go test ./...

fmt:
	gofmt -w $$(find . -type f -name '*.go' -not -path './vendor/*')

lint:
	golangci-lint run ./...

compose-up:
	docker compose up -d postgres

compose-down:
	docker compose down

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-version:
	go run ./cmd/migrate version

prod-up:
	$(PROD_COMPOSE) up -d --build

prod-down:
	$(PROD_COMPOSE) down

prod-logs:
	$(PROD_COMPOSE) logs -f api

smoke-test:
	BASE_URL=http://localhost:18080 ./scripts/smoke-test.sh

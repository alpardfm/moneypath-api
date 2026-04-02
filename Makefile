APP_NAME := moneypath-api

.PHONY: run test fmt lint compose-up compose-down migrate-up migrate-down migrate-version

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

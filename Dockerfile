FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /moneypath-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /moneypath-migrate ./cmd/migrate

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /moneypath-api /usr/local/bin/moneypath-api
COPY --from=builder /moneypath-migrate /usr/local/bin/moneypath-migrate
COPY --from=builder /app/migrations /app/migrations
COPY scripts/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["docker-entrypoint.sh"]

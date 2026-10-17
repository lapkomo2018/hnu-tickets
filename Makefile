.PHONY: run test lint fmt hooks swagger migrate-up migrate-down seed docker-up docker-down crypto-migrate crypto-rotate

# Load .env so make targets see the same settings as the API.
-include .env
export

POSTGRES_HOST ?= localhost
POSTGRES_PORT ?= 5432
POSTGRES_SSLMODE ?= disable
DATABASE_URL = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)

run:
	go run ./cmd/api

test:
	go test ./...

GOLANGCI_LINT = go tool -modfile=tools/go.mod golangci-lint

lint:
	$(GOLANGCI_LINT) fmt --diff
	$(GOLANGCI_LINT) run

fmt:
	$(GOLANGCI_LINT) fmt

hooks:
	go tool -modfile=tools/go.mod lefthook install

swagger:
	go generate ./...

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

seed:
	psql "$(DATABASE_URL)" -f scripts/seed.sql

docker-up:
	docker compose up --build

docker-down:
	docker compose down

crypto-migrate:
	go run ./cmd/cryptodata -mode migrate

crypto-rotate:
	go run ./cmd/cryptodata -mode rotate

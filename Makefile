.PHONY: run test lint fmt hooks swagger migrate-up migrate-down seed docker-up docker-down

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

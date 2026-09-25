.PHONY: run test lint swagger migrate-up migrate-down seed docker-up docker-down

run:
	go run ./cmd/api

test:
	go test ./...

lint:
	@test -z "$$(gofmt -l .)"
	go vet ./...

swagger:
	go tool swag init -g cmd/api/main.go -o docs --parseInternal

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

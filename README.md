# HNU Tickets

HNU Tickets is a REST API for managing technical support requests. It supports role-based ticket visibility, specialist assignment, a controlled ticket lifecycle, comments, an immutable audit trail, and administrative analytics.

## Architecture

The service is a modular monolith. Each feature owns its domain, application logic, HTTP adapter, and PostgreSQL adapter. Business rules do not depend on HTTP or SQL. PostgreSQL operations that update a ticket and its history run in one transaction.

Modules:

- Auth and Users
- Categories
- Tickets
- Comments
- History
- Analytics

More detail is available in [docs/architecture.md](docs/architecture.md).

## Requirements

- Go 1.27 or newer
- PostgreSQL 17 or newer
- Docker and Docker Compose for the container workflow
- `golang-migrate` and `psql` for local migration commands

## Quick start with Docker

```sh
docker compose up --build
```

The API is available at `http://localhost:8080`. Useful endpoints:

- Health: `GET /health`
- Readiness: `GET /ready`
- Swagger UI: `GET /swagger/`
- OpenAPI: `GET /openapi.yaml`

Load demo data after the services start:

```sh
docker compose exec -T postgres psql -U tickets -d tickets < scripts/seed.sql
```

## Local setup

Copy `.env.example` to `.env`, export its variables, and run:

```sh
make migrate-up
make seed
make run
```

Required environment variables:

| Variable | Purpose | Example |
| --- | --- | --- |
| `APP_ENV` | Runtime environment | `development` |
| `APP_PORT` | HTTP port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://tickets:tickets@localhost:5432/tickets?sslmode=disable` |
| `JWT_SECRET` | HMAC signing secret | a long random value |
| `JWT_TTL` | Access token lifetime | `24h` |
| `CORS_ALLOWED_ORIGINS` | Comma-separated browser origins | `http://localhost:5173` |

The application MUST NOT start without `DATABASE_URL` and `JWT_SECRET`.

## Database and migrations

Migrations are stored in `migrations/` and run automatically in Docker Compose. Local commands:

```sh
make migrate-up
make migrate-down
make seed
```

The schema and ER diagram are documented in [docs/database.md](docs/database.md).

## Demo accounts

All seeded accounts use the password `password123`.

| Email | Role |
| --- | --- |
| `admin@example.com` | ADMIN |
| `specialist@example.com` | SPECIALIST |
| `specialist2@example.com` | SPECIALIST |
| `user@example.com` | USER |
| `user2@example.com` | USER |

These accounts are for development only.

## Testing

```sh
make test
make lint
go build ./...
```

## API

All application endpoints use the `/api/v1` prefix. Send an access token as `Authorization: Bearer <token>`. Dates use RFC3339 and IDs use UUID.

See [docs/api.md](docs/api.md) for permissions and examples, or open the Swagger UI at `/swagger/`.

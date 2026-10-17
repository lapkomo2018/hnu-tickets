# HNU Tickets

HNU Tickets is a REST API for managing technical support requests. It supports role-based ticket visibility, specialist assignment, a controlled ticket lifecycle, comments, an API-level append-only audit trail, and administrative analytics.

## Documentation

Start with the [documentation index](docs/README.md): [specification](docs/specification.md), [business logic](docs/business_logic.md), [security policy](docs/security_policy.md), [privacy policy](docs/privacy_policy.md), [threat model](docs/threat_model.md), and [security requirements](docs/security.md).

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
cp .env.example .env
docker compose up --build
```

Compose takes every setting from `.env`; there are no credentials in `docker-compose.yml`.

The API is available at `http://localhost:8080`. Useful endpoints:

- Health: `GET /health`
- Readiness: `GET /ready`
- Swagger UI: `GET /docs/`
- OpenAPI specification: `GET /docs/doc.json`

Load demo data after the services start:

```sh
docker compose exec -T postgres psql -U tickets -d tickets < scripts/seed.sql
```

## Local setup

Copy `.env.example` to `.env` and run the commands below. Set `POSTGRES_HOST=localhost` first: the example
points at the Compose network. `make` loads `.env` itself; export the
variables only when starting the API another way, for example from an IDE.

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
| `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | Database credentials; Docker Compose uses the same values for the database container | `tickets` |
| `POSTGRES_HOST` | Optional. Database host, default `localhost`. Use `postgres` when the API runs in Docker Compose | `postgres` |
| `POSTGRES_PORT` | Optional. Database port, default `5432`; Compose also publishes PostgreSQL on it | `5432` |
| `POSTGRES_SSLMODE` | Optional. libpq `sslmode`, default `disable` | `require` |
| `JWT_SECRET` | HMAC signing secret | a long random value |
| `JWT_TTL` | Access token lifetime | `24h` |
| `CORS_ALLOWED_ORIGINS` | Comma-separated browser origins | `http://localhost:5173` |
| `ADMIN_EMAIL` | Optional. Email of the administrator created on startup | `admin@example.com` |
| `ADMIN_PASSWORD` | Required with `ADMIN_EMAIL`, at least 8 characters | a strong password |
| `ADMIN_NAME` | Optional. Display name of that administrator, default `Administrator` | `Administrator` |

The application MUST NOT start without `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, and `JWT_SECRET`. The API builds
the connection string from the `POSTGRES_*` variables and escapes the credentials.

## First administrator

Registration always creates `USER` accounts, and only an administrator can change roles
(`PATCH /api/v1/users/{id}/role`). To get the first administrator without seed data, set `ADMIN_EMAIL`
and `ADMIN_PASSWORD`: on startup the API creates that account with the `ADMIN` role if no user with this
email exists. An existing user is never modified, so the password is not reset and another account is
never promoted. Promote further users through the role endpoint.

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

`make lint` runs [golangci-lint](https://golangci-lint.run) (configured in `.golangci.yml`) and fails on
unformatted code or unsorted imports; `make fmt` fixes formatting. Linters and git hooks are pinned in
`tools/go.mod`, so no global install is needed.

Install the git hooks once per clone:

```sh
make hooks
```

The pre-commit hook formats staged Go files, regenerates the OpenAPI specification and stages it, lints,
checks `go mod tidy`, and runs the tests; the pre-push hook reruns the tests with the race detector.

## API

All application endpoints use the `/api/v1` prefix. Send an access token as `Authorization: Bearer <token>`. Dates use RFC3339 and IDs use UUID.

See [docs/api.md](docs/api.md) for permissions and examples, or open the Swagger UI at `/docs/`.

The specification is generated from the handler annotations and the DTO types with
[swaggo/swag](https://github.com/swaggo/swag). The generator itself is part of this module
(`cmd/swagger` drives the swag API), so no external CLI is needed. Regenerate after changing routes,
request bodies, or responses:

```sh
make swagger      # or: go generate ./...
```

This rewrites `docs/docs.go`, `docs/swagger.json`, and `docs/swagger.yaml`, which are committed and
embedded into the binary. `go test ./...` fails when a route is missing from the specification.

With `APP_ENV=development` the server runs that same generator on startup and serves the fresh
document, so annotations only need a restart to show up in the Swagger UI. The generator reads Go
sources, which the container image does not ship: any other environment serves the specification
compiled into the binary.


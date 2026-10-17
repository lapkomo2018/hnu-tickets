# Project policy — HNU Tickets

| Field | Value |
| --- | --- |
| Version | 1.0.0 |
| Status | Proposed for adoption through this pull request |
| Updated | 2026-10-17 |
| Owner | Repository maintainer, M. O. Lapko |
| Scope | Server code, tests, migrations, configuration and documentation in this repository |

This policy defines rules for changes to HNU Tickets. It adapts the course example to Go and the existing feature-based modular monolith. It does not assert that every historical implementation already satisfies every rule. Known gaps are documented in [Security](security.md).

## Change control

Policy changes must be explicit, reviewed by the maintainer and accompanied by a version increment and an explanation of affected rules. Subsequent policy changes use a separate `[POLICY]` commit. Initial adoption belongs to the `[INIT]` documentation baseline for laboratory No. 1. A feature change must not silently weaken this policy; a necessary exception must name its scope, reason and follow-up in the pull request and relevant documentation.

## Technology and structure

- Use the Go version declared in [`go.mod`](../go.mod), PostgreSQL and the existing `chi`, `pgx`, JWT and bcrypt libraries. Versions are pinned in module files; update them through a reviewed change.
- Keep the current `cmd/`, `internal/`, `migrations/`, `scripts/` and `docs/` structure. Tests live beside Go code as `*_test.go`.
- Features own their domain rules, application services, HTTP adapters and repository interfaces/implementations. The composition root is `cmd/api/main.go`.
- Domain rules must not depend on HTTP or SQL. Handlers decode and validate transport data; services authorize and coordinate operations; PostgreSQL adapters implement storage contracts.
- Cross-feature access uses narrow interfaces, such as `comments.AccessChecker`. Shared platform packages contain technical concerns rather than ticket business rules.
- Do not introduce an event bus, a new architecture hierarchy or a shared abstraction without a concrete requirement. Ticket history is a relational audit record, not event sourcing.

See [Architecture](architecture.md) for the actual component model and acknowledged coupling.

## Code and error handling

- Use `gofmt` and the repository's configured import formatting and `golangci-lint`; run `make fmt` and `make lint`.
- Use lower-case package names, `snake_case.go` filenames and Go naming conventions for exported and private identifiers. Preserve meaningful types such as `Status`, `Priority` and `Role`.
- Document exported APIs and non-obvious invariants. Avoid comments that merely restate a statement.
- Return Go errors; translate expected database conditions into feature errors. Preserve unexpected causes with operation context and `%w`.
- Map feature errors in their HTTP adapters. Do not return SQL errors, stack traces, credentials or raw internal errors to clients.
- Use `context.Context` for cancellable operations. Optional assignees and lifecycle timestamps may be pointers; no PHP-style null-object or inheritance scheme is required.

## Business rules and data consistency

- Maintain one authoritative [business-logic description](business_logic.md) and keep changes consistent with [FR-01–FR-12](specification.md#functional-requirements).
- Enforce authorization on the server for each operation, including reads of individual tickets, comments and history. UI restrictions alone are insufficient.
- Derive the actor from verified authentication. Never accept the actor's identity or role from a request body.
- Save ticket mutations and their history entries in one transaction. A failed history write must roll back the ticket mutation.
- Lock a mutable ticket before checking and applying a workflow change. Retain foreign keys, uniqueness and enum/check constraints in migrations.
- Use parameterized SQL. Dynamic ordering and grouping identifiers must come from server-controlled allowlists.
- Apply schema changes through numbered migrations. Review destructive migrations and recovery implications before executing them against valuable data.

## Information protection

- Store passwords only as salted bcrypt hashes; never log passwords, access tokens, signing keys or database credentials.
- Keep live secrets outside Git. `.env.example` contains examples only; `.env` and private SSH keys must not be committed. Use separate secrets for development and deployment.
- Enforce least privilege for users and deployment identities. Separate migration privileges from runtime database privileges before production deployment.
- Use HTTPS for network deployment and verified TLS for remote database connections. Development defaults do not establish a protected production deployment.
- Use synthetic data for tests, screenshots and demo seeds. Do not publish actual support messages or personal information in the repository.
- Treat backups and ticket history as confidential data. Access must be restricted; restoration must be checked on an isolated database.

Detailed requirements and open work are in [Security](security.md).

## Testing and verification

- Unit tests cover domain invariants, role decisions and invalid inputs as well as successful operations. Use fakes or mocks for external dependencies in unit tests.
- Use `httptest` for HTTP behavior. Database transaction and constraint guarantees require separate integration checks against an isolated PostgreSQL database; unit tests do not prove them.
- Add regression coverage for a changed business or security rule. Test unauthorized access using an unrelated USER and an unassigned SPECIALIST.
- Run `make lint`, `make test`, `go build ./...` and, for API changes, `make swagger` with a generated-file diff check. Run `go test -race ./...` where supported for concurrency-sensitive changes.
- Documentation-only changes require link, consistency and diff review. Do not claim a code test or security scenario passed unless it was executed.
- Existing CI performs linting, OpenAPI freshness checking, tests and build. A dependency-vulnerability check such as `govulncheck` is additional planned verification, not an existing CI guarantee.

## Git workflow

Use a task branch from `main` and submit a pull request describing the problem, final changes, verification and known limitations. Do not commit directly to `main` for routine work. Merge after review and successful required checks; branch protection enforcement depends on repository settings.

Use descriptive commit messages. The initial course documentation commit starts with `[INIT]`; subsequent documentation changes may use `[DOCS]`, fixes `[FIX]` and explicit policy revisions `[POLICY]`. This project does not require a separate `develop` branch.

Local hooks are installed with `make hooks`. See the root [README](../README.md#testing) for the current hook behavior. Hooks supplement CI and review; they do not replace them.

## Documentation maintenance

Update requirements, business rules, architecture, API and security notes in the same change when their behavior changes. Keep generated OpenAPI files generated; do not hand-edit them. Distinguish intended requirements, source-observed behavior and executed test results. Record new implementation gaps and close them only with implementation evidence and relevant checks.

# Architecture

## Modular monolith

HNU Tickets runs as one Go process and one PostgreSQL database. Code is grouped by feature instead of horizontal controller/service/repository folders. This keeps the deployment simple while preserving module boundaries.

```text
HTTP request
  -> authentication middleware
  -> feature handler
  -> application service and policy
  -> repository port
  -> PostgreSQL adapter
```

## Layers

- Domain code defines tickets, roles, priorities, statuses, and lifecycle rules.
- Application services authorize the actor and coordinate repositories.
- HTTP handlers decode and validate transport data. The error middleware maps application errors to status codes.
- PostgreSQL adapters use parameterized SQL through `pgx`.
- The composition root in `cmd/api/main.go` creates and connects dependencies.

## Error boundaries

Each feature owns its expected errors:

```text
PostgreSQL error -> feature error -> HTTP status and JSON error
```

PostgreSQL adapters translate expected database conditions into feature errors and wrap unexpected storage failures with operation context. Application services return those errors directly and create feature errors for validation and business rules. Each feature's HTTP adapter assigns statuses and public error codes for errors owned by that feature. The common error middleware writes the response, logs unexpected failures, and hides their internal details.

## Transaction boundaries

Ticket creation, editing, assignment, and status changes write the ticket and audit history in the same transaction. Repositories obtain the current transaction from the request context. Mutable ticket rows are selected with `FOR UPDATE` before workflow changes.

## Authentication and authorization

JWT authentication establishes a principal with a user UUID and role. Authorization remains in application services:

- USER sees and comments on tickets they created.
- SPECIALIST sees and comments on tickets assigned to them.
- ADMIN sees all tickets and manages users, categories, assignments, and analytics.

The token contains only `sub`, `role`, issue time, and expiry.

## Ticket state machine

```text
NEW -> ASSIGNED -> IN_PROGRESS -> RESOLVED -> CLOSED
                         |             |
                         v             v
                WAITING_FOR_USER   IN_PROGRESS
                         |
                         v
                   IN_PROGRESS
```

`NEW`, `ASSIGNED`, and `IN_PROGRESS` can transition to `CANCELLED` where allowed by policy. `CLOSED` and `CANCELLED` are terminal.

History is an append-only audit log. It is not event sourcing, and analytics reads the current relational model directly.

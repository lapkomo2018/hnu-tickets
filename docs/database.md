# Database

PostgreSQL stores five business tables and a token-revocation table. UUID primary keys are generated with `pgcrypto`. Status, priority, and role values are protected by check constraints.

```mermaid
erDiagram
    USERS ||--o{ TICKETS : creates
    USERS ||--o{ TICKETS : assigned
    USERS ||--o{ COMMENTS : writes
    USERS ||--o{ TICKET_HISTORY : changes
    CATEGORIES ||--o{ TICKETS : classifies
    TICKETS ||--o{ COMMENTS : contains
    TICKETS ||--o{ TICKET_HISTORY : records

    USERS {
        uuid id PK
        varchar name
        varchar email UK
        text password_hash
        varchar role
    }
    CATEGORIES {
        uuid id PK
        varchar name UK
        text description
    }
    TICKETS {
        uuid id PK
        varchar title
        text description
        varchar status
        varchar priority
        uuid category_id FK
        uuid created_by FK
        uuid assigned_to FK
        timestamptz resolved_at
        timestamptz closed_at
    }
    COMMENTS {
        uuid id PK
        uuid ticket_id FK
        uuid author_id FK
        text content
    }
    TICKET_HISTORY {
        uuid id PK
        uuid ticket_id FK
        uuid changed_by FK
        varchar action
        text old_value
        text new_value
    }
```

Indexes cover ticket status, priority, category, creator, assignee, creation time, and the ticket foreign keys used by comments and history.

Deleting a category referenced by a ticket returns a conflict. Tickets are never physically deleted through the API; cancellation is a status transition. Comments and history use cascading foreign keys only to preserve database consistency if a ticket is removed administratively outside the API.

## Token revocation

`revoked_tokens` stores a UUID token ID and expiry only. Every protected request checks it; logout inserts idempotently and prunes expired entries. Migration `000006` creates the table and expiry index. Raw bearer tokens are not stored.

## Protection boundaries

The schema constrains allowed status/role/priority values but does not enforce the complete lifecycle or require an assignee for ASSIGNED. Those are application invariants; the generic status-route gap is tracked as G-02 in [Security](security.md#implementation-gaps).

History has no API update/delete operation, but it is an ordinary database table. Cascading deletion and privileged SQL access mean it is not a tamper-proof ledger. Runtime/migration privilege separation and verified backup restoration remain required deployment work (G-04/G-08).


## Encrypted text

Migration `000007` adds the singleton `data_crypto_state` marker. `tickets.description`, `comments.content`, `ticket_history.old_value` and `ticket_history.new_value` retain their TEXT types but store AES-256-GCM envelopes after `cryptodata -mode migrate`. NULL history values stay NULL. UUIDs are allocated before insertion to bind ciphertext to its row. Data conversion and rotation are offline transactions; migration SQL alone does not encrypt data. See [Cryptography](cryptography.md).

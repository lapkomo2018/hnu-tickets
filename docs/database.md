# Database

PostgreSQL stores five core tables. UUID primary keys are generated with `pgcrypto`. Status, priority, and role values are protected by check constraints.

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

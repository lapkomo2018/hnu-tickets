# Project documentation

HNU Tickets is the server part of the **Technical Support Ticket Management Information System with Task Distribution among Specialists and Performance Analytics**. One deployment serves one organization; the web client is maintained separately.

| Document | Purpose |
| --- | --- |
| [Specification](specification.md) | Scope, FR-01–FR-12, constraints and acceptance criteria |
| [Project policy](project_policy.md) | Architecture, coding, testing and change-management rules |
| [Business logic](business_logic.md) | Entities, invariants, roles, lifecycle and analytics |
| [Architecture](architecture.md) | Components, dependencies, transactions and deployment |
| [Security policy](security_policy.md) | Mandatory development, operational and vulnerability-management controls |
| [Privacy policy](privacy_policy.md) | Data categories, purpose, access, retention and removal |
| [Threat model](threat_model.md) | DFD, trust boundaries, attacker capabilities and STRIDE scenarios |
| [Security requirements](security.md) | SEC-R controls, acceptance scenarios and implementation gaps |
| [Database](database.md) | Relational model, constraints and indexes |
| [API guide](api.md) | API conventions, permissions and workflow |
| [OpenAPI](swagger.yaml) | Generated endpoint and DTO contract |

Functional requirements retain FR-01–FR-12 across the specification and business rules. Security requirements use SEC-R identifiers, threats use TM identifiers and verification scenarios use SEC identifiers. Keep these links consistent when behavior changes.

Policies specify required controls. Implementation gaps identify where additional work is needed; acceptance scenarios define how to verify it.

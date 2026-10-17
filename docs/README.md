# Project documentation

HNU Tickets is the server part of the **Technical Support Ticket Management Information System with Task Distribution among Specialists and Performance Analytics**. One deployment serves one organization. The web client is a separate project.

| Document | Purpose |
| --- | --- |
| [Specification](specification.md) | Scope, FR-01–FR-12, constraints and acceptance criteria |
| [Project policy](project_policy.md) | Development, architecture, testing, Git and data-handling rules |
| [Business logic](business_logic.md) | Module responsibilities, entities, invariants, role and lifecycle policies |
| [Architecture](architecture.md) | Components, dependencies, transactions and deployment boundaries |
| [Security](security.md) | Assets, threats, security requirements, implementation gaps and verification scenarios |
| [Database](database.md) | Relational model, constraints and indexes |
| [API guide](api.md) | API conventions, permissions and example workflow |
| [OpenAPI](swagger.yaml) | Generated endpoint and DTO contract; regenerated with `make swagger` |

## Requirements baseline

The functional baseline is the existing Software Construction technical specification. FR identifiers retain its numbering; the documents describe the same product. Requirements express intended behavior. Statements about current implementation are based on source review at commit `d917d21e413380db4947556fc3661657a5636dab` on 2026-10-17. An implementation note is not a successful runtime test. Known differences are tracked in [Security](security.md#implementation-gaps) and the specification.

The policy and business-logic examples supplied with the Data and Information Security course have been adapted to the existing Go modular monolith. PHP class conventions, a separate CQRS hierarchy and an event bus are not part of this implementation.

## Data and Information Security — laboratory work No. 1

| Documentation task | Repository artifact |
| --- | --- |
| Organize `docs/` | This index and the documents above |
| Define project policy | `project_policy.md` |
| Describe business logic | `business_logic.md` |
| Specify functions, constraints and acceptance | `specification.md` |
| Describe the architecture | `architecture.md`, `database.md` |
| Specify protection of information | `security.md` |
| Record the initial course documentation | A commit starting with `[INIT]` on `docs/security-lab-1` |

This course reuses the existing `lapkomo2018/hnu-tickets` repository. The methodological guide requests a new repository with a prescribed name; reuse is a deliberate deviation and is not presented as fulfillment of that item. A separate laboratory report and evidence of the student's local Git/SSH setup are outside this repository documentation change. No SSH setup or laboratory execution is claimed by these files.

## Source materials

- [Data and Information Security, laboratory No. 1 methodological guide](https://docs.google.com/document/d/1iJtRZeSMFSKFcpaQBQ9KOlaom3Acgdsv/edit).
- [Project-policy example](https://drive.google.com/file/d/1BRme_6QQjqVPtwIT8GmrSNNKmtlF9Jzw/view).
- [Business-logic example](https://drive.google.com/file/d/1R8fGb1U_6BTMf5woqgaeaoC61g0_-B74/view).
- [Software Construction, laboratory No. 1 technical specification](https://docs.google.com/document/d/17-yJyBPrhCOHj3Ee_OQHcGjJoC46sjPGtDXMujjBZ0U/edit).

The course guide references ISO/IEC/IEEE 29148:2018 for requirements, ISO/IEC/IEEE 42010:2022 for architecture descriptions and NIST SP 800-218 for secure development. These documents use explicit requirements, architecture concerns and verification criteria; no formal certification or full standards-conformance assessment is asserted.

# API guide

Base URL: `/api/v1`

Successful resources are returned directly as JSON. Errors always use:

```json
{
  "error": {
    "code": "INVALID_STATUS_TRANSITION",
    "message": "invalid status transition: cannot transition ticket from NEW to CLOSED"
  }
}
```

## Permissions

This table describes the intended role policy. See [business logic](business_logic.md) for the full lifecycle matrix and [security gaps](security.md#implementation-gaps) for source-observed differences, including the remaining assignment-invariant gap. Authentication and logout behavior is described in [Authentication](authentication.md).

| Area | USER | SPECIALIST | ADMIN |
| --- | --- | --- | --- |
| Register, login, own profile | Yes | Yes | Yes |
| List categories | Yes | Yes | Yes |
| Manage categories | No | No | Yes |
| Manage users and roles | No | No | Yes |
| Create tickets | Yes | No | No |
| Read tickets | Own | Assigned | All |
| Edit ticket fields | Own while NEW | No | Yes |
| Assign specialists | No | No | Yes |
| Work ticket lifecycle | Limited | Assigned | All valid transitions |
| Read and add comments | Own | Assigned | All |
| Read history | Own | Assigned | All |
| Analytics | No | No | Yes |

## Typical workflow

1. Register or log in with `POST /auth/register` or `POST /auth/login`.
2. Create a ticket with `POST /tickets` as USER.
3. Assign it with `PATCH /tickets/{id}/assign` as ADMIN.
4. Move it through `IN_PROGRESS`, `WAITING_FOR_USER`, and `RESOLVED` with `PATCH /tickets/{id}/status`.
5. Close the resolved ticket as its creator or an administrator.

Ticket listing supports `page`, `pageSize`, `q`, `status`, `priority`, `categoryId`, `assignedTo`, `createdFrom`, `createdTo`, `sort`, and `order`. Analytics supports RFC3339 `from` and `to` period filters.

The complete schema, request bodies, and responses are in the generated `swagger.yaml`, served as Swagger UI at `/docs/`.


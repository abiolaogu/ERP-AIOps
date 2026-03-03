# API Overview

## API Principles

- RESTful resources with explicit Tenant scoping
- Versioned path strategy (`/v1`) for backward compatibility
- Standard error response envelope
- Idempotency for write operations with financial effects

## Base URL by Environment

- dev: `https://api.dev.product.example.com/v1`
- staging: `https://api.staging.product.example.com/v1`
- prod: `https://api.product.example.com/v1`

## Core Resource Areas

- Authentication: session token lifecycle
- Tenants and Orgs: provisioning and metadata management
- Users and Roles: lifecycle and authorization assignment
- Audit Logs: compliance and troubleshooting timeline
- Billing: subscription and charge operations
- Notifications: channel delivery and status tracking

## Error Model

```json
{
  "error": {
    "code": "ROLE_ASSIGNMENT_DENIED",
    "message": "Admin does not have permission to assign requested role",
    "trace_id": "trc_01HY..."
  }
}
```

## Pagination and Filtering

- `page`, `page_size` for list endpoints
- `from`, `to`, `action`, `actor_user_id` filters for audit logs

## Idempotency

- `Idempotency-Key` required for `POST /billing/charges`
- server stores key with request hash and response snapshot

## Related Documents

- [OpenAPI Contract](openapi.yaml)
- [AuthN and AuthZ](AuthN-AuthZ.md)


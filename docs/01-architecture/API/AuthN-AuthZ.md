# Authentication and Authorization

## Authentication (AuthN)

- Protocol: OAuth2/OIDC-compatible login with JWT bearer tokens
- Token types:
  - Access token: short-lived, used for API calls
  - Refresh token: rotated, revocable token for session renewal

### JWT Claims

- `sub`: User identifier
- `tenant_id`: Tenant scope
- `org_id`: Org scope when applicable
- `roles`: role identifiers for RBAC
- `env`: Environment assertion for telemetry and policy checks

## Authorization (AuthZ)

### RBAC Model

- Roles are scoped to Tenant and optional Org.
- Permissions are atomic operations, e.g. `user.assign_role`, `billing.create_charge`.
- Role templates:
  - `tenant_admin`
  - `org_admin`
  - `security_analyst`
  - `finance_manager`
  - `integration_engineer`

### Enforcement Flow

1. Validate token and claims.
2. Resolve route action metadata.
3. Evaluate role-policy matrix in context of Tenant and Org.
4. Allow or deny request and emit authorization audit event.

## Security Controls

- MFA required for Admin access in staging and prod.
- Token revocation list checked on high-risk operations.
- Session invalidation after role downgrade or suspicious behavior.

## Sample Policy Matrix

| Action | tenant_admin | org_admin | security_analyst | finance_manager |
|---|---:|---:|---:|---:|
| tenant.create | ✅ | ❌ | ❌ | ❌ |
| user.assign_role | ✅ | ✅ | ❌ | ❌ |
| audit.read | ✅ | ✅ | ✅ | ❌ |
| billing.charge.create | ✅ | ❌ | ❌ | ✅ |

## Audit Requirements

- Record allow/deny decisions for privileged actions.
- Include actor, action, target, decision reason, and Environment.


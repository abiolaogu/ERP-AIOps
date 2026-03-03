# Low Level Design (LLD)

## Module Design

### Identity and Access Module

- Validates JWT signature, issuer, audience, and expiration.
- Resolves role claims for Tenant and Org scope.
- Enforces route-level policy checks.

### Tenant Management Module

- Creates Tenant, default Org, baseline policies.
- Maintains tenant status lifecycle (active, suspended, archived).

### Billing Module

- Maintains subscription and usage records.
- Creates idempotent charge commands with external payment reference.

### Audit Module

- Writes immutable events with actor/action/target metadata.
- Supports indexed querying by Tenant, Org, actor, and action.

## Module Interaction (Pseudo UML)

```mermaid
classDiagram
  class AuthMiddleware {
    +validateToken(jwt)
    +resolveContext()
    +authorize(action, resource)
  }

  class TenantService {
    +createTenant(request)
    +getTenant(tenantId)
  }

  class UserService {
    +createUser(tenantId, request)
    +assignRole(tenantId, userId, role)
  }

  class BillingService {
    +createCharge(tenantId, request)
    +getBillingSummary(tenantId)
  }

  class AuditService {
    +recordEvent(event)
    +searchEvents(filter)
  }

  AuthMiddleware --> TenantService
  AuthMiddleware --> UserService
  AuthMiddleware --> BillingService
  TenantService --> AuditService
  UserService --> AuditService
  BillingService --> AuditService
```

## Sequence Diagrams

### Login

```mermaid
sequenceDiagram
  participant User
  participant Portal
  participant API
  participant Identity
  participant Audit

  User->>Portal: Submit credentials
  Portal->>API: POST /auth/login
  API->>Identity: Verify credentials
  Identity-->>API: Token pair with claims
  API->>Audit: record login success event
  API-->>Portal: Access and refresh tokens
  Portal-->>User: Session started
```

### Create Tenant

```mermaid
sequenceDiagram
  participant Admin
  participant AdminPortal
  participant API
  participant TenantSvc
  participant UserSvc
  participant Audit

  Admin->>AdminPortal: Create Tenant request
  AdminPortal->>API: POST /tenants
  API->>TenantSvc: createTenant()
  TenantSvc->>UserSvc: createDefaultOrgAdmin()
  TenantSvc->>Audit: record tenant.created
  API-->>AdminPortal: Tenant created response
```

### Role Assignment

```mermaid
sequenceDiagram
  participant OrgAdmin
  participant AdminPortal
  participant API
  participant UserSvc
  participant Audit
  participant Notify

  OrgAdmin->>AdminPortal: Assign role to User
  AdminPortal->>API: PATCH /tenants/{id}/users/{id}/roles
  API->>UserSvc: validate policy and assign role
  UserSvc->>Audit: record user.role_assigned
  UserSvc->>Notify: send role change notification
  API-->>AdminPortal: role assignment result
```

### Billing Charge

```mermaid
sequenceDiagram
  participant System
  participant API
  participant BillingSvc
  participant PaymentGateway
  participant Audit

  System->>API: POST /tenants/{id}/billing/charges
  API->>BillingSvc: createCharge()
  BillingSvc->>PaymentGateway: request charge authorization
  PaymentGateway-->>BillingSvc: charge outcome + reference
  BillingSvc->>Audit: record billing.charge_created
  API-->>System: charge accepted response
```

### Audit Log Write

```mermaid
sequenceDiagram
  participant Service
  participant AuditSvc
  participant AuditDB
  participant SIEM

  Service->>AuditSvc: emit event envelope
  AuditSvc->>AuditDB: append immutable event
  AuditSvc->>SIEM: forward normalized event
  AuditSvc-->>Service: ack with event_id
```

## Data Access Patterns

- Every query includes Tenant key filter.
- Write operations include actor identity from token claims.
- Sensitive fields are encrypted before persistence.

## Caching Strategy

- Role and policy cache with short TTL (5 min) and explicit invalidation.
- Tenant configuration cache with event-based invalidation.

## Failure Handling

- Retries with exponential backoff for transient Service calls.
- Dead-letter queue for failed notification or audit forwarding operations.
- Idempotency keys for external billing requests.

## LLD-to-Test Mapping

- Policy engine tests for role assignment edge cases
- Transactional tests for tenant creation workflow
- Contract tests for billing and notification integrations


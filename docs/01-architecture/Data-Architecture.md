# Data Architecture

## Multi-Tenant Data Strategy

- Every business table includes `tenant_id` and optional `org_id`.
- Tenant-scoped indexes support high-cardinality query performance.
- Access policies require tenant context from validated token claims.

## ERD

```mermaid
erDiagram
  TENANTS ||--o{ ORGS : owns
  TENANTS ||--o{ USERS : contains
  ORGS ||--o{ USERS : manages
  USERS ||--o{ USER_ROLE_BINDINGS : receives
  ROLES ||--o{ USER_ROLE_BINDINGS : maps
  TENANTS ||--o{ AUDIT_EVENTS : emits
  TENANTS ||--o{ SUBSCRIPTIONS : has
  SUBSCRIPTIONS ||--o{ USAGE_RECORDS : tracks
  SUBSCRIPTIONS ||--o{ CHARGES : bills
  TENANTS ||--o{ NOTIFICATIONS : sends

  TENANTS {
    string id PK
    string name
    string status
    datetime created_at
  }
  ORGS {
    string id PK
    string tenant_id FK
    string name
    datetime created_at
  }
  USERS {
    string id PK
    string tenant_id FK
    string org_id FK
    string email
    string status
    datetime created_at
  }
  ROLES {
    string id PK
    string scope
    string name
  }
  USER_ROLE_BINDINGS {
    string id PK
    string tenant_id FK
    string user_id FK
    string role_id FK
    datetime assigned_at
  }
  AUDIT_EVENTS {
    string event_id PK
    string tenant_id FK
    string action
    string actor_user_id
    datetime occurred_at
  }
  SUBSCRIPTIONS {
    string id PK
    string tenant_id FK
    string plan_code
    string status
  }
  USAGE_RECORDS {
    string id PK
    string subscription_id FK
    string metric_key
    decimal quantity
    datetime recorded_at
  }
  CHARGES {
    string id PK
    string subscription_id FK
    string status
    decimal amount
    string external_reference
  }
  NOTIFICATIONS {
    string id PK
    string tenant_id FK
    string channel
    string status
    datetime sent_at
  }
```

## Data Dictionary

| Entity | Field | Type | Description | PII Class |
|---|---|---|---|---|
| users | email | string | User login and contact | PII-High |
| users | status | string | active/suspended/invited | PII-Low |
| audit_events | actor_user_id | string | User performing action | PII-Moderate |
| charges | amount | decimal | billed amount | Financial-Sensitive |
| notifications | channel | string | email/webhook/in_app | Operational |

## Retention Policies

| Data Set | Retention | Rationale |
|---|---|---|
| Audit events | 7 years | Compliance and investigations |
| Billing charges | 10 years | Finance and tax obligations |
| Notifications logs | 18 months | Operational troubleshooting |
| Access logs | 13 months | Security and incident response |
| Usage metrics | 24 months | Capacity and billing analytics |

## PII Classification

- **PII-High**: email, name, payment metadata references.
- **PII-Moderate**: User IDs linked to actions.
- **PII-Low**: status and role metadata without direct identity.

## Data Quality Rules

- Unique constraints on Tenant + email for Users.
- Non-null tenant_id for all tenant-scoped entities.
- Clock synchronization for all event timestamps.

## Backup and Restore Notes

- Point-in-time recovery snapshots every 15 minutes.
- Daily full backup with encryption and region replication.
- Quarterly restore drill by Environment.


# Software Architecture Document (SAD)

## Purpose

This document defines the architecture of the System that powers the Product across web, Admin, and API channels in a multi-tenant model.

## Architecture Drivers

- Tenant isolation and Org-level administration
- Security and compliance by default
- Low-latency API interactions with auditable workflows
- Operability across dev, staging, and production Environment
- GitHub-native delivery and traceability

## Architectural Style

- Domain-oriented services behind a versioned API
- Event-driven integration for audit, notifications, and billing side effects
- Shared platform services for identity, policy, observability, and deployment
- GitOps deployment model with declarative Environment state

## Core Domains

- Identity and Access Management
- Tenant and Org Management
- User and RBAC Management
- Billing and Usage
- Notification Delivery
- Audit and Compliance Evidence

## HLD Summary

- Frontend channels: User portal and Admin portal
- API gateway: auth enforcement, rate limits, routing
- Domain Services: stateless compute, tenant-aware persistence
- Data layer: transactional store + immutable audit store
- Platform layer: observability, secret management, runtime controls

For full detail: [HLD](HLD.md).

## LLD Summary

- Module contracts, class boundaries, and sequence flows for critical operations
- Tenant and Org context propagation across API and Service boundaries
- Read/write patterns for audit consistency and billing idempotency

For full detail: [LLD](LLD.md).

## Constraints

- Every API operation must include Tenant context.
- Every privileged action must emit audit events.
- Deployment to production must pass policy checks and security scans.

## Key Decisions

1. JWT bearer tokens for API authorization with role claims.
2. Audit event writing as part of transactional boundary for critical operations.
3. Feature flags for progressive release by Tenant and Environment.
4. GitOps-based deployment promotion with Fleet.

## Risks and Countermeasures

| Risk | Countermeasure |
|---|---|
| Cross-tenant data leakage | Row-level scoping, tenant guards, query policy tests |
| Audit event loss | Retry with dead-letter queues and reconciliation jobs |
| Release drift | Fleet reconciliation and manifest policy checks |
| Runtime misconfiguration | Standardized Environment contracts and preflight checks |

## Traceability Map

- BRD goals -> PRD requirements -> API contracts -> test plan -> runbooks
- ADR records capture deviations and rationale for future decisions

## Related Documents

- [HLD](HLD.md)
- [LLD](LLD.md)
- [Security Architecture](Security-Architecture.md)
- [Data Architecture](Data-Architecture.md)
- [Technical Specifications](Technical-Specifications.md)


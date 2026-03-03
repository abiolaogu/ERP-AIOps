# AIOps Category-King Architecture Addendum

## Mission

This addendum extends the core Software Architecture for ERP-AIOps with category-king criteria: high-throughput UX, strict Tenant and Org isolation, resilient Service behavior, and operator-grade observability in every Environment.

## Context Diagram

```mermaid
flowchart TB
    U["User"]
    A["Admin"]
    P["ERP-AIOps Product UI"]
    API["Unified API Layer"]
    SVC["Domain Services (aiops)" ]
    AUD["Audit + Observability Service"]
    IAM["Identity Service (OIDC/JWT)"]
    NTFY["Notification Service"]

    U --> P
    A --> P
    P --> API
    API --> SVC
    SVC --> AUD
    API --> IAM
    API --> NTFY
```

## Container and Interaction View

```mermaid
flowchart LR
    FE["Next.js App Router UI"] --> Q["TanStack Query Client"]
    Q --> G["GraphQL/REST Gateway"]
    G --> AUTH["AuthN/AuthZ Service"]
    G --> CORE["AIOps Domain Service"]
    CORE --> EVT["Event Plane"]
    CORE --> LOG["Audit Stream"]
    EVT --> RT["Realtime Gateway"]
    RT --> FE
```

## Deployment Topology

```mermaid
flowchart TB
    subgraph DEV["Environment: dev"]
      DEV_UI["Product UI"]
      DEV_API["Gateway + Services"]
    end

    subgraph STG["Environment: staging"]
      STG_UI["Product UI"]
      STG_API["Gateway + Services"]
      STG_OBS["Tracing + Metrics + Logs"]
    end

    subgraph PRD["Environment: prod"]
      PRD_UI["Multi-region UI"]
      PRD_API["Autoscaled Gateway + Services"]
      PRD_DATA["Tenant-isolated data plane"]
      PRD_OBS["SLO + alerting stack"]
    end

    DEV_UI --> DEV_API
    STG_UI --> STG_API --> STG_OBS
    PRD_UI --> PRD_API --> PRD_DATA
    PRD_API --> PRD_OBS
```

## Architecture Control Points

| Control Point | Standard | Verification |
|---|---|---|
| Tenant isolation | tenant_id on all records + scoped query predicates | contract tests + query lint rules |
| Access control | policy-first checks before mutation | auth regression suite + audit sampling |
| Auditability | immutable event envelope for critical actions | event schema validation + replay tests |
| Performance | latency budgets per top workflow | SLO dashboards + load test gates |
| Release safety | progressive deployment + rollback trigger | canary policy + automated rollback drills |

## Critical Sequence: Role Assignment with Audit Evidence

```mermaid
sequenceDiagram
    participant Admin as Admin
    participant UI as Product UI
    participant API as API Gateway
    participant Auth as Auth Service
    participant Domain as AIOps Service
    participant Audit as Audit Service

    Admin->>UI: Assign role to User in Org
    UI->>API: POST /roles/assign
    API->>Auth: Validate Admin scope + policy
    Auth-->>API: Authorized
    API->>Domain: Apply role binding
    Domain->>Audit: Emit ROLE_ASSIGNED event
    Audit-->>Domain: Evidence id returned
    Domain-->>API: Success + evidence id
    API-->>UI: 200 + updated permissions
```

## Performance and Reliability Budgets

| Path | Budget | SLO |
|---|---|---|
| interactive read workflows | p95 < 200ms | 99.9% availability |
| mutation workflows | p95 < 300ms | <0.5% failed mutations |
| realtime invalidation | propagation < 2s | 99.5% under target |
| audit write | durable write < 500ms | zero lost critical events |

## Architecture Decision Triggers

- If p95 exceeds threshold for 3 consecutive releases, enforce performance hardening sprint.
- If policy violation anomalies rise above baseline, block feature releases pending control fix.
- If incident MTTR worsens quarter-over-quarter, require runbook redesign and ownership update.

## Linkage to Core Architecture Set

- [SAD](SAD.md)
- [HLD](HLD.md)
- [LLD](LLD.md)
- [Security Architecture](Security-Architecture.md)
- [Data Architecture](Data-Architecture.md)
- [Technical Specifications](Technical-Specifications.md)

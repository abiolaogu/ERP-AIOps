# Sovereign GA Implementation - ERP-AIOps

Last updated: 2026-02-28

## Category Intent

- Category: **AIOps Automation**
- Competitive benchmarks: **Dynatrace, Datadog, PagerDuty**
- North-star objective: Move from reactive incident response to proactive reliability autopilot.

## Category-King Differentiators

- Anomaly to remediation in a single event stream
- Blast-radius aware action approval gates
- Cross-domain topology understanding with causal hints
- Autonomous runbook simulation before live execution

## Mapped Sovereign Stack

- Primary SQL plane: **YugabyteDB** with tenant context enforcement ('SET app.tenant_id') and row-level isolation.
- Event backbone: **Redpanda** with tenant namespacing ('tenant.{id}.incidents.events' style topics).
- Cache plane: **DragonflyDB** for low-latency read paths, idempotency, and rate limiting.
- Realtime fan-out: **Centrifugo** channels with JWT-scoped tenant and role claims.
- API composition: **Cosmo + Hasura/Directus pattern** where applicable for federated ownership.
- Storage: **RustFS/MinIO compatible object layer** with per-tenant bucket/policy isolation.

## Phase-by-Phase Delivery Contract

### Phase 0 - Foundation
- Standardize runtime configuration and secrets templates.
- Ensure CI gates include lint, unit tests, and build checks.
- Lock tenant/context propagation across API boundaries.

### Phase 1 - Data and Cache
- Enforce tenant-aware data access patterns by default.
- Define canonical entities and indexing strategy for high-volume paths.
- Introduce cache invalidation contracts tied to event emissions.

### Phase 2 - Eventing and Realtime
- Publish and consume Redpanda topics for domain-critical workflows.
- Use outbox/CDC patterns for durable state propagation.
- Broadcast high-value updates through Centrifugo channels.

### Phase 3 - UX and Product Intelligence
- Apply Sovereign command-surface UI shell for all core operators.
- Add workflow acceleration lanes with visible guardrails.
- Improve task completion speed with guided states and fewer clicks.

### Phase 4 - Reliability and Security
- Add release gates for SLO regression, security checks, and policy coverage.
- Hard-block prohibited actions (cross-tenant data access, unsafe destructive ops).
- Maintain immutable audit events for high-impact actions.

### Phase 5 - GA Operations
- Publish runbooks for incident response and rollback execution.
- Validate disaster-recovery assumptions and recovery exercises.
- Confirm production readiness against the GA contract in 'infra/sovereign/ga-contract.yaml'.

## Target SLOs

- Availability: **99.95%**
- p95 user-path latency: **<= 180 ms**
- MTTR: **<= 15 minutes**

## Immediate Next Actions

1. Enforce contract-driven event schemas for: incidents.events, anomalies.events, remediation.actions.
2. Add synthetic checks for top 3 revenue/operations critical journeys.
3. Drive weekly scorecard review for latency, reliability, and UX completion rate.

<!-- SOVEREIGN_DOCS_REFRESH_LINK_2026_03 -->
## Documentation Refresh Pointer (2026-03-01)

Latest normalized documentation status is maintained in:
- `docs/SOVEREIGN_DOCS_LATEST_UPDATE_2026-03-01.md`

# Sovereign Prompt Execution Pack - ERP-AIOps

These prompts operationalize the Sovereign SaaS 2026 guidance for **ERP-AIOps** (AIOps Automation).

## Prompt 1 - Domain and Tenant Contract

```
You are implementing hard multi-tenancy for ERP-AIOps.

Tasks:
1. Enumerate core domain entities and tenant boundaries.
2. Add or verify tenant context propagation for API, DB, cache, and events.
3. Ensure prohibited cross-tenant data paths fail closed.
4. Generate tests for tenant isolation and role-permission enforcement.

Output:
- Updated domain model docs
- Tenant isolation tests
- Risk list with mitigations
```

## Prompt 2 - Event Backbone and Realtime

```
You are implementing event-native workflows for ERP-AIOps.

Tasks:
1. Define event contracts for: incidents.events, anomalies.events, remediation.actions.
2. Implement publisher and consumer paths with idempotency keys.
3. Add CDC/outbox integration where transactional consistency is required.
4. Expose realtime UX updates via Centrifugo channels.

Output:
- Event contract artifacts
- Producer/consumer implementation
- Replay-safe integration tests
```

## Prompt 3 - Category-King UX

```
You are redesigning ERP-AIOps for category leadership against Dynatrace, Datadog, PagerDuty.

Tasks:
1. Optimize the top 5 user workflows for speed and clarity.
2. Surface guardrails, confidence, and approvals inline.
3. Implement meaningful motion and visual hierarchy for decision tasks.
4. Add telemetry for workflow completion and friction hotspots.

Output:
- Updated UX components and pages
- UX telemetry hooks
- Before/after task-time metrics
```

## Prompt 4 - GA Readiness and Reliability

```
You are preparing ERP-AIOps for immediate production deployment.

Tasks:
1. Enforce SLO checks (availability 99.95%, p95 <= 180ms).
2. Add release gates for tests, lint, policy checks, and security scans.
3. Validate runbooks for incident response and rollback.
4. Produce a deployment verification checklist.

Output:
- CI gate updates
- Operational readiness checklist
- Residual risk register
```

# ERP-AIOps - Sovereign Release Checklist

Generated: 2026-02-28

## Pre-Release

- [x] Tenant isolation and guardrail tests pass.
- [x] Event contracts and schema compatibility artifacts are present.
- [x] AIDD guardrail policy changes reviewed and committed.
- [x] Rollout and rollback playbooks updated.
- [x] Phase 3 gates pass (`lint`, `test`, `build`).

## Canary

- [x] Canary rollout plan defined with stop conditions.
- [x] Canary validation checklist prepared (latency, errors, workflow completion).
- [x] Cross-tenant leakage watchpoints and alerts defined.

## Production

- [x] Gradual rollout plan prepared with automated stop conditions.
- [x] SLO and guardrail audit monitoring plan prepared.
- [x] Repository marked GA-ready for deployment execution.

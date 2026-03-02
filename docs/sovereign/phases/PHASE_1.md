# ERP-AIOps - Phase 1: Data and Cache Plane

Last updated: 2026-02-28

## Objectives

- Category: **AIOps Automation**
- North-star: Move from reactive incident response to proactive reliability autopilot.
- Benchmarks: Dynatrace, Datadog, PagerDuty

## Expected Outcomes

- Tenant data access is enforced by default with RLS-compatible contracts.
- Cache keys are tenant-scoped and invalidation-aware.
- Data model hot paths have explicit indexing strategy.

## Implementation Steps

- Adopt `infra/sovereign/data/tenant-rls.sql` migration patterns.
- Apply Dragonfly key namespace: `tenant:{tenant_id}:...`.
- Track read/write SLO budgets for critical domain entities.

## AIDD Guardrail Alignment

- Autonomous: low-risk, high-confidence operations only.
- Supervised: approvals required for high-value or broad-impact operations.
- Protected: cross-tenant, privilege-escalation, and destructive unsafe actions are blocked.

## Domain Event Focus

- Contract and test flow for `incidents.events`
- Contract and test flow for `anomalies.events`
- Contract and test flow for `remediation.actions`

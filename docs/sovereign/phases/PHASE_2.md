# ERP-AIOps - Phase 2: Eventing and Realtime

Last updated: 2026-02-28

## Objectives

- Category: **AIOps Automation**
- North-star: Move from reactive incident response to proactive reliability autopilot.
- Benchmarks: Dynatrace, Datadog, PagerDuty

## Expected Outcomes

- Critical workflows emit durable event contracts through Redpanda.
- Realtime channels are tenant-scoped and role-aware.
- Replay-safe idempotency is required for all consumers.

## Implementation Steps

- Implement topic contracts from `infra/sovereign/events/topic-contracts.yaml`.
- Use outbox/CDC style publish path for transactional integrity.
- Fan out UX updates via channel contracts in `infra/sovereign/realtime`.

## AIDD Guardrail Alignment

- Autonomous: low-risk, high-confidence operations only.
- Supervised: approvals required for high-value or broad-impact operations.
- Protected: cross-tenant, privilege-escalation, and destructive unsafe actions are blocked.

## Domain Event Focus

- Contract and test flow for `incidents.events`
- Contract and test flow for `anomalies.events`
- Contract and test flow for `remediation.actions`

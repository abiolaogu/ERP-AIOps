# ERP-AIOps Next.js 16 + Workik Migration Baseline

Date: 2026-03-01

## Objective

Adopt the Sovereign SaaS 2026 frontend and shared-infrastructure standard:

- Frontend: Next.js 16 + Shadcn UI + Tailwind v4
- Service pattern: Workik-style features/*/services + TanStack Query hooks
- Data path: Cosmo -> Hasura DDN -> YugabyteDB
- Event path: Yugabyte CDC -> Redpanda Connect -> Redpanda -> Centrifugo -> Query invalidation

## Delivered Baseline

- New scaffold path: apps/sovereign-next16
- Shared Workik contract updated in .workik/config.yml
- Example inventory feature with Zod schema, service class, and realtime-aware hook

## Cutover Steps

1. Port domain slices from legacy frontend into apps/sovereign-next16/features.
2. Point all frontend data access to ERP_SHARED_COSMO_URL.
3. Remove legacy Refine/Ant Design dependencies after feature parity is validated.
4. Promote apps/sovereign-next16 to primary web app and archive previous UI stack.

## AIDD Guardrail Notes

- Keep tenant identifiers mandatory in all queries/mutations.
- Enforce mutation validation in service layer before API calls.
- Treat realtime events as cache invalidation signals, not source of truth.

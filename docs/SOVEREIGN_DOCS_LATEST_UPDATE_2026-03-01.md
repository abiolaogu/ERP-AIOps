# Sovereign Documentation Refresh — ERP-AIOps

Date: 2026-03-01
Last Updated UTC: 2026-03-01T19:44:09Z
Status: GA Ready (Validated)

## What Was Refreshed

- Consolidated current-state architecture and deployment posture.
- Normalized shared infrastructure references to:
  - `ERP_SHARED_HASURA_URL`
  - `ERP_SHARED_DB_URL`
  - `ERP_SHARED_IAM_URL`
  - `ERP_SHARED_OTLP_ENDPOINT`
- Confirmed AIDD guardrail alignment for protected/supervised/autonomous operations.
- Linked updated Figma Make Prompt expansion for design + automation handoff.

## Shared Infrastructure Contract

- Hasura GraphQL: `http://localhost:8090/v1/graphql`
- Hasura WS: `ws://localhost:8090/v1/graphql`
- ERP-DBaaS: PostgreSQL shared service (`ERP_SHARED_DB_URL`)
- ERP-IAM: OIDC/JWT shared IAM (`ERP_SHARED_IAM_URL`)
- ERP-Observability: OTLP/metrics/traces (`ERP_SHARED_OTLP_ENDPOINT`)

## GA Validation Snapshot

- Infra consolidation validator: PASS
- GA predeploy validator: PASS
- Phase 6 cutover validator: PASS

## Primary Documentation Set

- GA implementation: `docs/SOVEREIGN_GA_IMPLEMENTATION.md`
- Prompt execution: `docs/SOVEREIGN_PROMPT_EXECUTION.md`
- Readiness scorecard: `docs/SOVEREIGN_GA_READINESS_SCORECARD.md`
- Architecture: `docs/ARCHITECTURE.md`
- Deployment: `docs/DEPLOYMENT.md`
- Security: `docs/SECURITY.md`

## Figma Make Prompt Packs

- Repo-level design prompt file: `docs/design/Figma_Make_Prompts.md`
- Portfolio-level module prompt file: `/Users/AbiolaOgunsakin1/ERP/Documentation/ERP-AIOps/15-ERP-AIOps_Figma-Make-Prompts.md`

## Operations Baseline

- Run full validation: `make ga-all`
- Run infra consolidation validation: `./tools/sovereign/validate-infra-consolidation.sh .`
- Regenerate cutover manifest: `./tools/sovereign/ga-phase6-cutover.sh .`

## Notes

This file is the canonical “latest-level” refresh indicator for ERP-AIOps. It does not replace deep-dive docs; it links and normalizes them.

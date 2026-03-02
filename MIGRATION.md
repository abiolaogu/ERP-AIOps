# Sovereign ERP 2026 Migration - ERP-AIOps

## Removed

- Module-owned infrastructure provisioning for legacy identity, object storage, brokers, caches, and local SQL runtimes.
- Legacy generated display-only components and obsolete wiring.
- Forbidden stack references in active module code and configs.

## Added

- Frontend-only Next.js 16 App Router baseline with TypeScript.
- Shadcn-style UI primitives with Tailwind CSS v4 design tokens.
- Workik feature-slice architecture:
  - `src/features/control-center/services`
  - `src/features/control-center/hooks`
  - `src/features/control-center/ui`
- Realtime provider with Centrifugo invalidation topic convention:
  `${NEXT_PUBLIC_ENV}.${NEXT_PUBLIC_ORG}.${NEXT_PUBLIC_MODULE}.${tenant}.ui.invalidate`

## Infra Centralization

- Module now depends on shared services in `/ERP/shared-infra`.
- This repo remains UI/control-plane only.

## Validation

```bash
npm install --no-audit --no-fund
npm run typecheck
npm run build
```

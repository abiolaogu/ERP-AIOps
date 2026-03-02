# Sovereign ERP 2026 Migration - ERP-AIOps

## Removed

- Legacy identity/object-store/broker/cache/database runtime artifacts from module scope
- Legacy UI framework bindings and generated display-only fragments
- Module-owned infrastructure manifests

## Added

- Next.js 16 frontend-only shell (TypeScript)
- Shadcn + Tailwind v4 design baseline
- Workik feature-slice layout
- Realtime provider with Centrifugo + TanStack Query invalidation
- Auth policy with ERP-IAM OIDC and configured fallback mode

## Validation

```bash
npm install
npm run typecheck
npm run build
```

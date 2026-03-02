# ERP-AIOps

Sovereign ERP 2026 module frontend.

## Local Development

- Primary URL: http://localhost:5179
- Additional URLs: none
- Auth policy: dev-token-fallback

```bash
npm install --no-audit --no-fund
npm run dev
npm run typecheck
npm run build
```

## Architecture

- Next.js 16 App Router + TypeScript
- Shadcn-style primitives + Tailwind CSS v4
- Workik feature slices under `src/features/*`
- TanStack Query default staleTime of 5 minutes
- Realtime invalidation via Centrifugo topic:
  `${NEXT_PUBLIC_ENV}.${NEXT_PUBLIC_ORG}.${NEXT_PUBLIC_MODULE}.
  ${tenant}.ui.invalidate`
- OIDC via ERP-IAM (Authentik) with fallback policy based on module requirements

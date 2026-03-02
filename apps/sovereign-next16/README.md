# Next.js 16 + Shadcn + Workik Template

Reference implementation of the Sovereign SaaS frontend standard.

## Architecture

- `app/`: Next.js 16 App Router routes
- `features/*/services`: business logic layer (Workik pattern)
- `features/*/hooks`: TanStack Query hooks
- `core/providers`: query + realtime providers
- `ui/`: lightweight Shadcn-style primitives

## Realtime Pattern

- Redpanda topic emits `ENTITY_UPDATED`
- Centrifugo broadcasts to `tenant_updates:tenant_<tenant_id>`
- `RealtimeProvider` invalidates TanStack cache keys by entity + tenant

## Startup

```bash
cp .env.example .env.local
npm install
npm run dev
```

# AIOps Integration — ERP-AIOps

## Module Info

- **Module**: ERP-AIOps
- **Frontend URL**: http://localhost:5179
- **Topic Prefix**: `*.*.erp.aiops.*`
- **Consumer Group Pattern**: `cg.[env].[org].aiops.[service]`

## Heartbeat

ERP-AIOps publishes its own health heartbeat to `*.*.erp.aiops.health.heartbeat`.

## Event Format

```json
{
  "envelope": {
    "env": "dev",
    "org": "abiola",
    "platform": "erp",
    "app": "aiops",
    "tenant": "system",
    "entity": "health",
    "event": "heartbeat",
    "event_id": "uuid",
    "occurred_at": "2026-03-02T00:00:00Z",
    "schema_version": "1.0",
    "producer": "svc-erp-aiops"
  }
}
```

## Dev-Token Fallback

When `NEXT_PUBLIC_AUTH_POLICY=dev-token-fallback`, AIOps UI uses prefilled credentials for local development.

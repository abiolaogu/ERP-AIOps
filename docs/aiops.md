# AIOps Integration

This module is monitored by the shared ERP-AIOps platform.

## Signals Emitted

| Signal | Topic | Description |
|--------|-------|-------------|
| Health heartbeat | `[env].[org].erp.aiops.health` | 30s interval liveness/readiness |
| Error events | `[env].[org].erp.aiops.errors` | Application-level errors |
| Metric snapshots | `[env].[org].erp.aiops.metrics` | Key business/operational metrics |

## AIOps Capabilities Available

- Anomaly detection on module metrics
- SLO tracking (availability, latency, error rate)
- Automated incident creation on threshold breach
- Autonomous remediation (pod restart, cache clear) within guardrail bounds
- Escalation via notification channels

## Guardrail Tiers

| Tier | Risk | Actions | Approval |
|------|------|---------|----------|
| Autonomous | ≤ 3 | restart_pod, clear_cache, create_incident | None |
| Supervised | ≤ 7 | scale_horizontally, config_change, rollback | 1 approval |
| Protected | ≤ 10 | failover, schema_change, cross_module | 2 approvals |

## Integration Contract

- Module registry: `ERP-AIOps/configs/modules/registry.yaml`
- Runbooks: `ERP-AIOps/configs/runbooks/`
- Standards: `ERP-AIOps/docs/aiops-standards.md`

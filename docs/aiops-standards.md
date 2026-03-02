# AIOps Standards — Shared Autonomous IT Operations

ERP-AIOps is the shared operations brain for all Sovereign ERP modules. It consumes health signals, detects anomalies, and orchestrates remediation across the platform.

## Signals Consumed

| Signal | Source | Topic | Purpose |
|--------|--------|-------|---------|
| Health heartbeat | All modules | `*.*.erp.aiops.health.heartbeat` | Module liveness |
| Deploy event | CI/CD pipelines | `*.*.erp.<module>.deploy.*` | Change tracking |
| Error budget | Module telemetry | `*.*.erp.<module>.slo.*` | SLO monitoring |
| Alerts | Alertmanager | Webhook | Incident creation |
| Observability data | ERP-Observability | `*.*.erp.observability.*` | Anomaly correlation |

## Actions Available

| Action | Tier | Approval | Description |
|--------|------|----------|-------------|
| `restart_pod` | Autonomous | None | Restart unhealthy pod |
| `clear_cache` | Autonomous | None | Flush DragonflyDB cache |
| `scale_up_by_one` | Autonomous | None | Add one replica |
| `create_incident` | Autonomous | None | Create incident record |
| `send_notification` | Autonomous | None | Alert via channel |
| `scale_horizontally` | Supervised | 1 approval | Scale replicas |
| `config_change` | Supervised | 1 approval | Modify runtime config |
| `rollback_deployment` | Supervised | 1 approval | Revert to prior version |
| `failover_primary` | Protected | 2 approvals | Database failover |
| `cross_module_remediation` | Protected | 2 approvals | Multi-module action |

## Module Integration Contract

Every module MUST:
1. Publish heartbeat events to `*.*.erp.aiops.health.heartbeat` (every 30s)
2. Publish deploy events on release
3. Expose `/healthz` and `/ready` endpoints
4. Use consumer group pattern: `cg.[env].[org].[module].[service]`

## AIOps Topics

- `*.*.erp.aiops.health.heartbeat` — module health heartbeats
- `*.*.erp.aiops.alerts.created` — alert notifications to all modules
- `*.*.erp.aiops.incidents.created` — incident lifecycle events
- `*.*.erp.aiops.remediation.requested` — remediation requests
- `*.*.erp.aiops.remediation.executed` — remediation execution results
- `*.*.erp.aiops.audit.actions` — immutable audit trail

## Dashboard

AIOps Dashboard (http://localhost:5179):
- Module Registry viewer
- Incident list
- Remediation queue
- SLO status overview

# Observability

## Observability Scope

The System uses logs, metrics, and traces to monitor Product health and support incident response across Environments.

## Signal Standards

### Logging

- Structured JSON logs
- Required fields: `timestamp`, `level`, `service`, `tenant_id`, `org_id`, `trace_id`, `environment`
- No raw secrets or sensitive payloads

### Metrics

- RED metrics (Rate, Errors, Duration) for APIs
- queue depth and retry rates for async Services
- billing and notification business metrics

### Tracing

- Distributed trace from gateway to Service to persistence layer
- Cross-service span propagation with trace context headers

## SLI and SLO Definitions

| SLI | SLO |
|---|---|
| API success ratio | >= 99.9% per month |
| API p95 latency | <= 200 ms read, <= 350 ms write |
| Auth success ratio | >= 99.95% |
| Audit event ingestion | >= 99.99% successful writes |
| Incident ack time | <= 10 minutes sev-1 |

## Alerting Strategy

- Error budget burn alerts
- latency threshold alerts by route and Service
- auth failure anomaly alerts
- billing charge failure spike alerts
- audit pipeline lag alerts

## Dashboards

- Executive reliability dashboard
- Service-level operational dashboard
- Tenant health dashboard for support teams
- Security and compliance event dashboard

## Operational Reviews

- Daily on-call review of active alerts
- Weekly reliability review with action items
- Monthly SLO and error budget governance


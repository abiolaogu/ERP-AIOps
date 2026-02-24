# ERP-AIOps

AIOps platform for intelligent operations management. Provides automated incident management, anomaly detection, root cause analysis, event correlation, auto-remediation, cost optimization, security scanning, topology mapping, forecasting, and adaptive thresholds across the ERP ecosystem.

## Architecture

| Component       | Technology          | Port  | Description                          |
|----------------|---------------------|-------|--------------------------------------|
| Gateway        | Go (net/http)       | 8090  | API gateway with CORS, auth, logging |
| Rust API       | Axum                | 8080  | Core REST API for all AIOps entities |
| AI Brain       | Python FastAPI      | 8001  | AI/ML analysis and forecasting       |
| YugabyteDB     | PostgreSQL-compat   | 5433  | Primary data store                   |
| DragonflyDB    | Redis-compat        | 6379  | Caching layer                        |
| Hasura         | GraphQL Engine      | 19109 | GraphQL federation endpoint          |
| Frontend       | React + Refine.dev  | 5179  | AIOps management UI                  |

## Quick Start

```bash
# Start all services
make dev

# Or start individually
make gateway    # Go gateway on :8090
make api        # Rust API on :8080
make ai-brain   # Python AI brain on :8001
make web        # Frontend on :5179
```

## API Routes

### Gateway (port 8090)

| Route                    | Backend      | Description                |
|-------------------------|--------------|----------------------------|
| /v1/aiops/incidents     | rust-api     | Incident management        |
| /v1/aiops/anomalies     | rust-api     | Anomaly detection          |
| /v1/aiops/rules         | rust-api     | AIOps rule engine          |
| /v1/aiops/topology      | rust-api     | Service topology           |
| /v1/aiops/remediation   | rust-api     | Auto-remediation           |
| /v1/aiops/cost          | rust-api     | Cost optimization          |
| /v1/aiops/security      | rust-api     | Security scanning          |
| /v1/aiops/analyze       | ai-brain     | AI analysis                |
| /v1/aiops/forecast      | ai-brain     | Forecasting                |
| /v1/aiops/health        | gateway      | Health check               |

## Multi-Tenancy

All tables use `tenant_id TEXT NOT NULL` for tenant isolation. The gateway extracts `X-Tenant-ID` from request headers and forwards it to backend services.

## Module Dependencies

- **Required:** ERP-Platform (auth, tenant management)
- **Optional:** ERP-Observability (metrics, logs, traces ingestion)

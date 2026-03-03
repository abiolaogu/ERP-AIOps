# Sovereign AIOps -- Technology Architecture

**Confidential | Series A | March 2026**

---

## 1. Architecture Overview

Sovereign AIOps is a cloud-native, event-driven platform built on five architectural pillars: high-throughput data ingestion, real-time event correlation, ML-powered anomaly detection, topology-aware intelligence, and autonomous remediation execution. The platform is designed to process 50 million events per minute at sub-second latency while maintaining the safety guarantees required for autonomous production operations.

```
┌─────────────────────────────────────────────────────────────────────┐
│                        DATA SOURCES                                 │
│  Metrics │ Logs │ Traces │ Events │ Custom │ ERP Modules           │
└─────────┬───────┬────────┬────────┬────────┬───────────────────────┘
          │       │        │        │        │
          ▼       ▼        ▼        ▼        ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    INGESTION LAYER                                   │
│  OpenTelemetry Collectors │ Custom Agents │ API Gateway (Port 5179) │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    REDPANDA EVENT BUS                                │
│  Shared event bus across all ERP modules                            │
│  Topics: metrics, logs, traces, events, incidents, remediations     │
│  Throughput: 50M events/min │ Latency: <10ms p99                    │
└──────┬──────────┬───────────┬──────────┬────────────────────────────┘
       │          │           │          │
       ▼          ▼           ▼          ▼
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────────┐
│ Anomaly  │ │ Event    │ │ Topology │ │ AIDD Engine          │
│ Detection│ │Correlation│ │ Engine   │ │ (Guardrail Framework)│
│ Service  │ │ Service  │ │          │ │                      │
└──────┬───┘ └──────┬───┘ └──────┬───┘ └──────────┬───────────┘
       │            │            │                 │
       ▼            ▼            ▼                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    INTELLIGENCE LAYER                                │
│  Incident Manager │ RCA Engine │ Capacity Planner │ SLO Tracker     │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    EXECUTION LAYER                                   │
│  Runbook Engine │ Rollback Controller │ Health Check Validator       │
│  847 pre-built runbooks │ Custom runbook runtime │ Sandbox mode      │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    PRESENTATION LAYER                                │
│  React + Vite + Refine.dev │ REST API │ GraphQL │ WebSocket         │
│  Dashboards │ Topology Maps │ Incident Timeline │ Executive Reports │
└─────────────────────────────────────────────────────────────────────┘
```

## 2. Core Services

### 2.1 Ingestion Service

**Purpose:** Collect telemetry data from all infrastructure sources with exactly-once semantics and sub-second latency.

| Component | Technology | Purpose |
|---|---|---|
| OTel Collector Fleet | OpenTelemetry Collector | Standard telemetry collection (OTLP, Prometheus, StatsD) |
| Custom Agent | Go binary (12MB) | Lightweight agent for topology discovery and runbook execution |
| API Gateway | Go (net/http) on port 5179 | REST API for custom metrics, events, and webhook ingestion |
| Schema Registry | Redpanda Schema Registry | Avro schema validation for all event types |
| Rate Limiter | Token bucket (per-tenant) | Protect platform from ingestion spikes; graceful backpressure |

**Performance:**
- Ingestion throughput: 50M events/minute (sustained), 200M events/minute (burst)
- End-to-end latency: <100ms from source to Redpanda topic (p99)
- Agent footprint: 12MB binary, <1% CPU, <50MB RAM per host

### 2.2 Redpanda Event Bus

**Purpose:** Central nervous system of the platform. All data flows through Redpanda topics, enabling real-time correlation across data types and ERP modules.

| Topic | Data Type | Retention | Consumers |
|---|---|---|---|
| `aiops.metrics.raw` | Time-series metrics | 7 days | Anomaly Detection, ClickHouse |
| `aiops.logs.raw` | Log events | 3 days | Anomaly Detection, Search |
| `aiops.traces.raw` | Distributed traces | 3 days | Topology Engine, RCA |
| `aiops.events.raw` | Infrastructure events | 30 days | Event Correlation, Incident Manager |
| `aiops.incidents` | Correlated incidents | 90 days | AIDD Engine, Dashboard |
| `aiops.remediations` | Execution records | 365 days | Audit Trail, Analytics |
| `erp.*.events` | ERP module events | 7 days | Cross-domain Correlation |

**Why Redpanda (not Kafka):**
- 10x lower tail latency at high throughput (critical for real-time correlation)
- Single binary deployment (no ZooKeeper/KRaft complexity)
- Shared event bus with other ERP modules eliminates data silos
- 40% lower infrastructure cost at our event volume

### 2.3 Anomaly Detection Service

**Purpose:** Identify deviations from normal behavior across all telemetry signals using ensemble ML models.

**Model Architecture:**

| Model | Type | Use Case | Inference Time |
|---|---|---|---|
| Statistical Baseline | ARIMA + Prophet | Seasonal pattern detection | <5ms |
| Multivariate Detector | Isolation Forest | Multi-metric anomaly scoring | <10ms |
| Deep Anomaly Model | Transformer (custom) | Complex temporal pattern detection | <50ms |
| Log Anomaly Detector | BERT-based classifier | Log pattern deviation detection | <30ms |
| Ensemble Router | Gradient boosted tree | Combines model outputs into final score | <2ms |

**Training Pipeline:**
- Models retrained daily on rolling 30-day window per tenant
- Transfer learning from cross-customer aggregated data (anonymized)
- A/B testing framework for model promotion (shadow mode → canary → production)
- Model performance tracking: precision, recall, F1, detection latency

**Performance Metrics:**
- Precision: 94.7% (false positive rate: 5.3%)
- Recall: 91.2% (missed anomaly rate: 8.8%)
- Detection latency: <200ms from metric ingestion to anomaly alert
- Training time: <45 minutes per tenant (daily retrain)

### 2.4 Event Correlation Engine

**Purpose:** Group related events and anomalies into a single incident, reducing alert noise by 73%.

**Correlation Strategies:**

| Strategy | Mechanism | Example |
|---|---|---|
| Temporal | Events within configurable time window (default: 5 minutes) | CPU spike and OOM kill within 3 minutes |
| Topological | Events on services with known dependency relationships | Database latency → API timeout → frontend error |
| Causal | ML-inferred causal relationships from historical incident data | Deployment event → memory leak → cascading failure |
| Semantic | NLP-based similarity of log messages and error descriptions | Similar stack traces across different services |
| Cross-Domain | ERP module events correlated with infrastructure events | Invoice processing delay → database replica lag |

**Noise Reduction Metrics:**
- Average noise reduction: 73% (range: 62% to 84% across customers)
- Correlation accuracy: 96.2% (validated by customer feedback)
- Average events per incident: 47 events correlated into 1 actionable incident

### 2.5 Topology Engine

**Purpose:** Maintain a real-time dependency map of all infrastructure components, services, and their relationships.

**Discovery Methods:**

| Method | Coverage | Latency |
|---|---|---|
| Agent-based discovery (network connections, process list) | 92% of services | Real-time |
| Trace-based inference (distributed tracing spans) | 87% of service dependencies | Near real-time |
| Config-based (Kubernetes labels, Terraform state, AWS tags) | 78% of infrastructure | On change |
| ML-inferred (traffic pattern analysis) | 65% of undocumented dependencies | Hourly |
| **Combined** | **97% total coverage** | **<5 minute refresh** |

**Topology Graph:**
- Storage: Neo4j graph database
- Nodes: 45,000 (current) -- services, hosts, containers, databases, load balancers, queues
- Edges: 312,000 (current) -- dependencies, network flows, data flows, ownership
- Query latency: <50ms for 3-hop dependency traversal
- Blast radius calculation: <200ms for any node

### 2.6 AIDD Engine

**Purpose:** Govern all autonomous actions through the three-tier guardrail framework.

**Decision Flow:**

```
Incident Detected
    │
    ▼
┌─────────────────┐
│ Classify Incident│ ← Category, severity, blast radius
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Check AIDD Tier │ ← Per-category tier configuration
└────────┬────────┘
         │
    ┌────┴────┬──────────┐
    ▼         ▼          ▼
 Tier 1    Tier 2     Tier 3
 (Monitor)  (Suggest)   (Act)
    │         │          │
    ▼         ▼          ▼
  Log &    Generate   ┌─────────────┐
  Alert    Suggestion │ Pre-Flight  │
           + Notify   │ Checks      │
              │       └──────┬──────┘
              ▼              │
           Human          Pass? ──No──→ Escalate to Tier 2
           Approves?         │
              │           Yes │
           Yes│              ▼
              ▼       ┌─────────────┐
           Execute    │ Execute     │
           Runbook    │ Runbook     │
              │       └──────┬──────┘
              ▼              │
           Post-Flight       ▼
           Validation  Post-Flight
              │       Validation
              ▼              │
           Report     Pass? ──No──→ Auto-Rollback
                             │
                          Yes │
                             ▼
                          Report + Close
```

**Safety Controls (Tier 3):**

| Control | Description |
|---|---|
| Pre-execution health check | Verify target system health before runbook execution |
| Blast radius limit | Maximum number of affected resources per execution (configurable) |
| Change window enforcement | Runbooks only execute during approved maintenance windows |
| Concurrent execution limit | Maximum simultaneous runbook executions per tenant |
| Automatic rollback | If post-execution health check fails, reverse all changes |
| Human override | Any user can halt autonomous execution with one-click kill switch |
| Audit immutability | All execution records written to append-only audit log |

### 2.7 Runbook Execution Engine

**Purpose:** Execute validated remediation actions against customer infrastructure.

**Architecture:**
- Execution runtime: Isolated container per runbook execution (gVisor sandboxed)
- Credential management: HashiCorp Vault integration; customer-managed secrets
- Supported platforms: AWS (IAM role assumption), GCP (service account), Azure (managed identity), Kubernetes (RBAC), SSH (key-based)
- Execution timeout: Configurable per runbook (default: 5 minutes)
- Retry policy: Configurable (default: 1 retry with exponential backoff)

**Pre-built Runbook Categories:**

| Category | Count | Examples |
|---|---|---|
| Kubernetes | 187 | Pod restart, HPA adjustment, node drain, PDB update |
| AWS | 156 | EC2 instance recovery, RDS failover, Lambda throttle, S3 lifecycle |
| Database | 124 | Connection pool reset, slow query kill, replica promotion, vacuum |
| Networking | 98 | DNS flush, load balancer health check, firewall rule, CDN purge |
| Application | 89 | Service restart, config reload, cache flush, feature flag toggle |
| Security | 72 | Certificate rotation, key rotation, access revocation, IP block |
| GCP | 68 | GKE scaling, Cloud SQL failover, Pub/Sub reset |
| Azure | 53 | AKS operations, Cosmos DB failover, App Service restart |
| **Total** | **847** | |

## 3. Data Architecture

### 3.1 Data Stores

| Store | Technology | Purpose | Retention |
|---|---|---|---|
| Event Bus | Redpanda | Real-time event streaming | 7-365 days by topic |
| Time-Series | ClickHouse | Metric storage and analytics | 13 months |
| Graph | Neo4j | Topology and dependency mapping | Current state + 90-day history |
| Operational | PostgreSQL (RDS) | User data, config, incidents, runbooks | Indefinite |
| Search | OpenSearch | Log search and full-text query | 30 days (configurable) |
| ML Models | S3 + MLflow | Model artifacts and experiment tracking | All versions |
| Audit Trail | S3 (append-only) | Immutable execution and access logs | 7 years |
| Cache | Redis | Session data, API cache, rate limiting | Ephemeral |

### 3.2 Data Flow

1. **Ingestion**: Source → OTel Collector/Agent → Redpanda topic (raw)
2. **Processing**: Redpanda → Anomaly Detection + Correlation + Topology → Redpanda topic (processed)
3. **Storage**: Redpanda → ClickHouse (metrics), OpenSearch (logs), Neo4j (topology), PostgreSQL (incidents)
4. **Intelligence**: Processed data → ML models → Anomaly scores, predictions, RCA graphs
5. **Execution**: AIDD decision → Runbook Engine → Customer infrastructure → Health check → Report
6. **Presentation**: PostgreSQL/ClickHouse → API (port 5179) → React frontend

## 4. Infrastructure

### 4.1 Deployment Architecture

| Component | Infrastructure | Scaling |
|---|---|---|
| Application Services | AWS EKS (Kubernetes) | Horizontal pod autoscaling |
| Redpanda Cluster | Dedicated EC2 (i3en.xlarge) | Vertical + horizontal |
| ClickHouse Cluster | Dedicated EC2 (r6g.2xlarge) | Sharding by tenant |
| PostgreSQL | AWS RDS (Multi-AZ) | Read replicas |
| Neo4j | Dedicated EC2 (r6g.xlarge) | Causal cluster (3 nodes) |
| ML Inference | AWS EKS with GPU nodes (g5.xlarge) | Queue-based autoscaling |
| ML Training | AWS SageMaker (spot instances) | Batch scheduling |

### 4.2 High Availability

| Metric | Target | Current |
|---|---|---|
| Platform uptime SLA | 99.95% | 99.97% (trailing 12 months) |
| Data durability | 99.999999999% (11 nines) | S3-backed |
| Recovery Point Objective (RPO) | <1 minute | Redpanda replication factor 3 |
| Recovery Time Objective (RTO) | <15 minutes | Automated failover with health checks |
| Multi-region DR | Active-passive (us-east-1 primary, eu-west-1 DR) | Tested quarterly |

### 4.3 Security Architecture

| Layer | Controls |
|---|---|
| Network | VPC isolation, private subnets, NACLs, security groups, WAF |
| Transport | TLS 1.3 everywhere, mutual TLS between services |
| Authentication | OAuth 2.0 + OIDC, SAML 2.0 SSO, MFA enforced |
| Authorization | RBAC with least-privilege, per-resource permissions |
| Data at Rest | AES-256 encryption, per-tenant KMS keys |
| Data in Transit | TLS 1.3, certificate pinning for agent communication |
| Secrets | HashiCorp Vault, automatic rotation, no secrets in code |
| Audit | Immutable append-only logs, 7-year retention, tamper detection |
| Compliance | SOC 2 Type I (complete), Type II (Q1 2027), ISO 27001 (Q2 2027) |

## 5. Technology Stack Summary

| Layer | Technology |
|---|---|
| Backend Services | Go 1.22 (primary), Python 3.12 (ML pipelines) |
| Frontend | React 18, Vite, Refine.dev, Ant Design, D3.js (topology visualization) |
| Event Streaming | Redpanda (shared across ERP modules) |
| Time-Series DB | ClickHouse |
| Graph DB | Neo4j |
| Operational DB | PostgreSQL 16 (via AWS RDS) |
| Search | OpenSearch 2.x |
| ML Framework | PyTorch 2.x, scikit-learn, Prophet, MLflow |
| ML Serving | TorchServe on GPU nodes, ONNX Runtime for lightweight models |
| Container Orchestration | Kubernetes (AWS EKS) |
| CI/CD | GitHub Actions, ArgoCD (GitOps) |
| Infrastructure as Code | Terraform, Helm |
| Observability | OpenTelemetry, Grafana, our own platform (dogfooding) |
| Secrets Management | HashiCorp Vault |
| API Gateway | Custom Go API server on port 5179 |

## 6. Scalability Considerations

### 6.1 Current Scale

| Metric | Value |
|---|---|
| Monitored Resources | 45,000 |
| Events Ingested/Day | 2.1 billion |
| Active ML Models | 320 (40 per customer) |
| Topology Nodes | 45,000 |
| Topology Edges | 312,000 |
| Concurrent Users | 180 |

### 6.2 Target Scale (2028)

| Metric | Value | Growth Factor |
|---|---|---|
| Monitored Resources | 620,000 | 14x |
| Events Ingested/Day | 28 billion | 13x |
| Active ML Models | 2,200 | 7x |
| Topology Nodes | 620,000 | 14x |
| Topology Edges | 4,340,000 | 14x |
| Concurrent Users | 2,500 | 14x |

### 6.3 Scaling Strategy

- **Redpanda**: Horizontal scaling via partition count increase; dedicated broker nodes per high-volume tenant
- **ClickHouse**: Sharding by tenant_id with distributed query routing
- **ML Inference**: Queue-based autoscaling with GPU node pools; model distillation for high-frequency models
- **Topology Engine**: Graph partitioning by tenant; in-memory caching for hot paths
- **API Layer**: Horizontal pod autoscaling based on request rate and latency

## 7. Technical Moat Assessment

| Moat | Time to Replicate | Defensibility |
|---|---|---|
| AIDD Guardrail Framework | 18-24 months | Patent pending; 18 months production hardening |
| Redpanda ERP Integration | 12-18 months | Requires shared event bus architecture from day one |
| ML Models (2.8B events trained) | 12-18 months | Data network effects; improves with each customer |
| 847 Pre-built Runbooks | 12 months | Community marketplace creates flywheel |
| Topology Auto-Discovery (97%) | 6-12 months | Multiple discovery methods combined |

---

*Confidential. Sovereign AIOps, Inc. All rights reserved.*

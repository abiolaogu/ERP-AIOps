# Technical Architecture Overview -- Sovereign AIOps

**Confidential -- Series A Investment Memorandum**

---

## 1. Architecture Summary

Sovereign AIOps is built as a cloud-native, event-driven platform with streaming ML inference. The architecture prioritizes three qualities: **integration breadth** (works with any monitoring tool), **real-time intelligence** (sub-30-second anomaly detection), and **safe automation** (guardrailed remediation with rollback).

---

## 2. System Architecture

```
                    External Monitoring Tools
    [Prometheus] [Datadog] [CloudWatch] [PagerDuty] [Custom Webhooks]
         |           |          |           |              |
         v           v          v           v              v
    +----------------------------------------------------------+
    |              Event Ingestion Gateway (Go)                 |
    |    Source Adapters -> Normalizer -> Enricher -> Router    |
    |              Throughput: 100K events/sec                  |
    +----------------------------------------------------------+
                              |
                    [Apache Kafka Cluster]
                    /         |          \
                   v          v           v
    +----------------+ +----------------+ +------------------+
    |   Anomaly      | |  Correlation   | |  Event Storage   |
    |   Detection    | |  Engine        | |  (PostgreSQL +   |
    |   Engine       | |  (Temporal +   | |   TimescaleDB +  |
    |   (IF+LSTM+    | |   Topological  | |   ClickHouse)    |
    |    Drain)      | |   + Bayesian)  | |                  |
    +----------------+ +----------------+ +------------------+
           |                   |
           v                   v
    +----------------+ +------------------+
    |   Remediation  | |  Notification    |
    |   Executor     | |  Router          |
    |   (Runbook DSL | |  (Slack, Teams,  |
    |    + Rollback) | |   PagerDuty)     |
    +----------------+ +------------------+
           |
    [Kubernetes API] -- auto-discovery + remediation actions

    +----------------------------------------------------------+
    |  Supporting Services                                      |
    |  [Topology Service] [SLO Calculator] [Capacity Planner]  |
    |  [Change Risk Analyzer] [Post-Mortem Generator]          |
    +----------------------------------------------------------+
           |
    +----------------------------------------------------------+
    |  API Layer: Hasura GraphQL Engine                         |
    |  Frontend: React + Vite + Refine.dev                     |
    +----------------------------------------------------------+
```

---

## 3. Technology Stack

| Layer | Technology | Justification |
|---|---|---|
| **Backend services** | Go 1.22+ | Performance, concurrency, Kubernetes ecosystem compatibility |
| **ML training** | Python (scikit-learn, PyTorch, Prophet) | ML ecosystem maturity |
| **ML inference** | ONNX Runtime (Go bindings) | Production performance without Python runtime overhead |
| **Event streaming** | Apache Kafka | Proven at millions of events/sec, exactly-once semantics |
| **Primary database** | PostgreSQL 16 | Reliability, extensions ecosystem, transactional integrity |
| **Time-series** | TimescaleDB (PostgreSQL extension) | Time-series queries on same PostgreSQL instance |
| **Analytics** | ClickHouse | Columnar storage for high-cardinality analytics |
| **Graph queries** | Apache AGE (PostgreSQL extension) | Topology graph traversal without separate graph database |
| **Vector search** | pgvector (PostgreSQL extension) | Semantic search for runbook matching, incident similarity |
| **Cache** | Redis Cluster | Real-time state, sliding windows, rate limiting |
| **API gateway** | Hasura GraphQL Engine | Auto-generated API from schema, real-time subscriptions |
| **Frontend** | React + Vite + Refine.dev + Ant Design | Enterprise-ready UI framework with data management |
| **Container orchestration** | Kubernetes | Industry standard, deployment target for customers too |

---

## 4. Key Technical Innovations

### 4.1 Streaming ML Inference in Go

Unlike competitors who batch-process events for ML analysis (introducing minutes of latency), Sovereign AIOps performs streaming inference directly in the event processing pipeline:

- Models exported from Python training to ONNX format
- ONNX Runtime Go bindings enable inference without Python runtime
- Sub-5-second latency for Isolation Forest, sub-30-second for LSTM
- No GPU required for inference (CPU-optimized models)

### 4.2 PostgreSQL-Everything Architecture

We consolidate four database needs into PostgreSQL with extensions:
- **Relational data:** Native PostgreSQL
- **Time-series:** TimescaleDB extension (hypertables, compression, retention)
- **Graph queries:** Apache AGE extension (Cypher query language)
- **Vector search:** pgvector extension (HNSW indexing for embeddings)

**Benefit:** Single operational burden, single backup strategy, consistent query interface.

### 4.3 Integration-First Ingestion

The Event Ingestion Gateway uses an adapter pattern:
- Each monitoring source has a dedicated adapter (Prometheus, Datadog, CloudWatch, etc.)
- Adapters translate source-specific formats to our Common Event Format (CEF)
- New integrations require only a new adapter -- no core pipeline changes
- Customers can build custom adapters via webhook + field mapping configuration

---

## 5. Scalability

| Dimension | Current Capacity | Design Limit | Scaling Method |
|---|---|---|---|
| Event ingestion | 100K events/sec | 1M+ events/sec | Kafka partitions + consumer groups |
| Anomaly detection | 50K inferences/sec | 500K/sec | Horizontal pod autoscaling |
| Event storage | 10TB | 100TB+ | TimescaleDB compression + partitioning |
| Analytics queries | 1B events scanned | 100B+ events | ClickHouse sharding |
| Topology nodes | 10K nodes | 100K+ nodes | AGE index optimization |
| Concurrent tenants | 100 | 10,000+ | Tenant-partitioned Kafka topics |

---

## 6. Security Architecture

- **Multi-tenancy:** All data paths enforce `tenant_id` isolation at database and API layers
- **Encryption:** AES-256 at rest (PostgreSQL TDE), TLS 1.3 in transit, mTLS for service-to-service
- **Authentication:** JWT via OAuth 2.0/OIDC (ERP-Auth module), SAML for enterprise SSO
- **Authorization:** RBAC (Viewer, Operator, Admin, Super Admin) enforced at Hasura permission layer
- **Audit:** Immutable audit log with cryptographic hash chain for all remediation actions
- **Secrets:** HashiCorp Vault for runbook credentials, auto-rotation every 24 hours
- **PII:** Auto-redaction of PII patterns in log data before storage
- **Compliance:** SOC2 Type II in progress, architecture supports FedRAMP

---

## 7. Deployment Models

| Model | Target Customer | Infrastructure |
|---|---|---|
| **SaaS (multi-tenant)** | SMB, mid-market | Sovereign-hosted on AWS/GCP |
| **Dedicated instance** | Enterprise | Single-tenant cloud deployment |
| **Private cloud** | Regulated industries | Customer's cloud account (Terraform/Helm) |
| **On-premises** | Government, financial services | Air-gapped Kubernetes deployment |

---

*This document is confidential and intended for potential investors only.*

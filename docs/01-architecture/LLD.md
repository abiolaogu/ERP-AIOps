# Low-Level Design (LLD) -- Sovereign AIOps Platform

**Module:** ERP-AIOps | **Port:** 5179 | **Version:** 2.0 | **Date:** 2026-03-03

---

## 1. Anomaly Detection Algorithms

### 1.1 Time Series Anomaly Detection -- Isolation Forest

**Purpose:** Detect point anomalies in individual metric values (CPU spikes, latency outliers, error rate jumps).

**Algorithm Details:**
- **Type:** Unsupervised ensemble method based on random partitioning
- **Intuition:** Anomalies are few and different; they require fewer random splits to isolate
- **Complexity:** Training O(n log n), Inference O(log n) per sample

**Implementation:**

```
Input: metric_value, service_id, metric_name, timestamp
Output: anomaly_score (0.0-1.0), is_anomaly (bool), explanation

Algorithm:
1. Retrieve baseline model for (service_id, metric_name) from model registry
2. If no model exists, enqueue for training (use rule-based fallback: >3 sigma)
3. Feature extraction:
   - raw_value: the current metric value
   - z_score: (value - rolling_mean_1h) / rolling_stddev_1h
   - rate_of_change: (value - value_5m_ago) / 5m
   - hour_of_day: cyclical encoding (sin/cos of hour)
   - day_of_week: cyclical encoding (sin/cos of day)
   - is_deployment_window: binary flag from change events
4. Run Isolation Forest inference:
   - anomaly_score = average_path_length across 100 trees
   - Normalize to 0.0-1.0 range
5. Dynamic threshold adjustment:
   - base_threshold = 0.65 (tuned per tenant via feedback)
   - If is_deployment_window: threshold += 0.15 (more tolerant during deploys)
   - If hour_of_day in [2:00-6:00] AND metric == "traffic": threshold += 0.10
6. is_anomaly = anomaly_score > adjusted_threshold
7. If is_anomaly: generate SHAP explanation (top 3 contributing features)
```

**Model Training Pipeline:**
```
Schedule: Daily at 02:00 UTC per tenant
Input: 90 days of metric data per (service, metric) pair
Steps:
1. Query TimescaleDB for historical metrics
2. Feature engineering (same as inference features + statistical aggregates)
3. Remove known incident windows (labeled data from incidents table)
4. Train Isolation Forest: n_estimators=100, max_samples=256, contamination=auto
5. Validate: holdout 20% of data, measure precision/recall against labeled anomalies
6. If precision > 0.85 AND recall > 0.70: deploy model to registry
7. If validation fails: retain existing model, alert ML team
```

### 1.2 Time Series Anomaly Detection -- LSTM

**Purpose:** Detect temporal pattern anomalies (gradual degradation, missing seasonality, unusual trends).

**Architecture:**
```
Input sequence: [x(t-59), x(t-58), ..., x(t)]  (60 time steps, 1-minute resolution)
Layer 1: LSTM(128 units, return_sequences=True)
Layer 2: Dropout(0.2)
Layer 3: LSTM(64 units, return_sequences=False)
Layer 4: Dense(32, activation='relu')
Layer 5: Dense(1, activation='linear')  -- predicts x(t+1)

Loss: MAE between predicted x(t+1) and actual x(t+1)
Anomaly: prediction_error = |predicted - actual|
Score: prediction_error / rolling_mean_error_30d
Threshold: score > 3.0 (adaptive per metric)
```

**Concept Drift Handling:**
- LSTM retrained weekly with sliding 90-day window
- Online learning: exponential moving average of prediction errors adjusts threshold
- Deployment events trigger temporary threshold relaxation (15-minute window)

### 1.3 Log Anomaly Detection -- Drain Algorithm

**Purpose:** Detect novel log patterns that deviate from established templates.

```
Drain Log Parser Configuration:
  depth: 4 (parse tree depth)
  similarity_threshold: 0.4
  max_children: 100

Process:
1. Receive log line: "Connection to db-primary timeout after 30s"
2. Tokenize: ["Connection", "to", "db-primary", "timeout", "after", "30s"]
3. Traverse Drain parse tree:
   a. Level 1: Route by log length (6 tokens)
   b. Level 2: Route by first token ("Connection")
   c. Level 3-4: Match against existing templates using token similarity
4. If match found (similarity > 0.4):
   - Assign to existing cluster, update template with wildcards
   - Template: "Connection to <*> timeout after <*>"
   - This is a KNOWN pattern -> no anomaly
5. If no match found:
   - Create new cluster -> this is a NOVEL pattern -> anomaly
   - anomaly_score based on rarity (1.0 for brand new cluster)
   - Frequency decay: score = 1.0 * exp(-count / 50)
6. Clusters with >1000 occurrences in 24h are considered baseline
7. Clusters inactive for 30 days are archived
```

---

## 2. Correlation Algorithm

### 2.1 Temporal Windowing

```
Configuration:
  window_size: 300s (5 minutes, configurable per tenant)
  slide_interval: 30s
  min_events_to_correlate: 2
  max_events_per_window: 500

Algorithm:
1. Maintain sliding window buffer per (tenant_id, environment)
2. On new event arrival:
   a. Add to current window
   b. Evict expired events (timestamp < now - window_size)
3. Attempt correlation within window:
   a. Group by shared attributes: {service, environment, region}
   b. For each group with >= min_events:
      - Calculate temporal density: events_in_group / window_size
      - If density > threshold: mark as temporally correlated
   c. Cross-service correlation:
      - If events from multiple services share time window:
        - Query topology service for dependency relationship
        - If dependency exists: mark as topologically correlated
```

### 2.2 Topological Correlation (Graph Traversal)

```
Algorithm: Upstream Root Cause Discovery
Input: Set of anomalous services S = {s1, s2, ..., sn}
Output: Root cause candidate with probability score

1. For each service si in S:
   a. BFS upstream in dependency graph (max depth = 5)
   b. Collect upstream services: upstream(si) = {u1, u2, ..., uk}
2. Find common upstream ancestors:
   candidates = intersection(upstream(s1), upstream(s2), ..., upstream(sn))
3. For each candidate c in candidates:
   a. Compute centrality: how many affected services depend on c?
   b. Compute temporal priority: did c's anomaly precede downstream anomalies?
   c. Compute anomaly severity: what is c's anomaly confidence score?
4. Rank candidates:
   score(c) = w1*centrality(c) + w2*temporal_priority(c) + w3*severity(c)
   where w1=0.4, w2=0.35, w3=0.25
5. Return top candidate as likely root cause
```

### 2.3 Bayesian Root Cause Inference

```
P(root_cause = service_X | symptoms) =
  P(symptoms | root_cause = service_X) * P(root_cause = service_X) / P(symptoms)

Features (symptoms):
  - Which services are anomalous
  - What types of anomalies (latency, errors, saturation)
  - Time of day, day of week
  - Recent deployment events
  - Current topology state

Training: Incremental update from historical incidents with confirmed root causes
Cold start: Use topological correlation until 50+ labeled incidents available

Output: Ranked list of (service, probability, reasoning)
```

---

## 3. Remediation Execution Engine

### 3.1 Runbook DSL Specification

```yaml
apiVersion: aiops.sovereign.io/v1
kind: Runbook
metadata:
  name: restart-unhealthy-pods
  description: Restart pods in CrashLoopBackOff state
  severity: [warning, critical]
  services: ["*"]
  mode: suggest-and-approve

trigger:
  conditions:
    - type: anomaly
      metric: pod_restart_count
      operator: gt
      value: 5
      window: 10m
    - type: status
      kubernetes:
        pod_status: CrashLoopBackOff

guardrails:
  max_affected_pods: 3
  max_affected_percentage: 25%
  required_healthy_replicas: 2
  cooldown: 30m
  blackout_windows:
    - cron: "0 9-17 * * 1-5"
      timezone: America/New_York

steps:
  - name: verify-issue
    action: kubernetes.get_pods
    params:
      namespace: "{{ incident.namespace }}"
      label_selector: "app={{ incident.service }}"
    register: pods

  - name: count-unhealthy
    action: filter
    params:
      input: "{{ pods }}"
      condition: "status.phase != 'Running' OR restart_count > 5"
    register: unhealthy_pods

  - name: check-guardrails
    action: assert
    params:
      - "{{ unhealthy_pods | length }} <= {{ guardrails.max_affected_pods }}"
      - "{{ unhealthy_pods | length }} / {{ pods | length }} <= 0.25"
    on_failure: abort

  - name: delete-unhealthy-pods
    action: kubernetes.delete_pods
    params:
      pods: "{{ unhealthy_pods }}"
      grace_period: 30
    rollback:
      action: kubernetes.scale_deployment
      params:
        replicas: "{{ pods | length }}"

  - name: wait-for-healthy
    action: wait
    params:
      condition: "all pods Running"
      timeout: 300s
      poll_interval: 10s

  - name: verify-health
    action: health_check
    params:
      endpoint: "{{ incident.service }}/healthz"
      expected_status: 200
      retries: 3
    on_failure: rollback

post_execution:
  - action: update_incident
    params:
      status: resolved
      resolution: "Auto-remediated: restarted {{ unhealthy_pods | length }} pods"
  - action: notify
    params:
      channel: "{{ incident.team_slack_channel }}"
      message: "Incident auto-resolved via runbook."
```

### 3.2 Step Executor State Machine

```
States: PENDING -> RUNNING -> VERIFYING -> COMPLETED | FAILED | ROLLED_BACK

PENDING -> RUNNING: Approval gate passed (or mode is autonomous)
RUNNING -> VERIFYING: All steps completed successfully
VERIFYING -> COMPLETED: Health checks pass within monitoring window
VERIFYING -> FAILED: Health checks fail, triggers rollback
FAILED -> ROLLED_BACK: Rollback steps executed in reverse order

Error Handling:
  - Step timeout: Mark step as FAILED, trigger rollback
  - Network error: Retry 3x with exponential backoff (1s, 2s, 4s)
  - Guardrail violation: Abort immediately, no rollback needed
  - Concurrent execution: Distributed lock per (service, runbook)
```

### 3.3 Audit Log Schema

```sql
CREATE TABLE remediation_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    incident_id UUID NOT NULL REFERENCES incidents(id),
    runbook_id UUID NOT NULL REFERENCES runbooks(id),
    execution_id UUID NOT NULL,
    step_name TEXT NOT NULL,
    step_index INTEGER NOT NULL,
    action TEXT NOT NULL,
    params JSONB NOT NULL,
    status TEXT NOT NULL CHECK (status IN (
        'started','completed','failed','rolled_back','skipped'
    )),
    output JSONB,
    error_message TEXT,
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    executed_by TEXT NOT NULL,
    approved_by TEXT,
    rollback_of UUID REFERENCES remediation_audit_log(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Immutable: no UPDATE or DELETE permissions granted
```

---

## 4. SLO Calculation Engine

### 4.1 Core Formulas

```
Error Budget Calculation:
  SLO Target: 99.9%
  Error Budget = 1 - 0.999 = 0.001 = 0.1%
  In 30-day window: 0.001 * 30 * 24 * 60 = 43.2 minutes of allowed downtime

Error Budget Consumption:
  actual_bad_minutes = count(minutes where SLI < threshold)
  consumption_rate = actual_bad_minutes / total_budget_minutes
  remaining = 1 - consumption_rate

Burn Rate:
  burn_rate = error_rate_in_window / error_budget_rate
  error_budget_rate = (1 - SLO_target) / window_hours

Multi-Window Burn Rate Alerting:
  | Window  | Burn Rate | Budget Exhaustion | Action |
  |---------|-----------|-------------------|--------|
  | 1 hour  | 14.4x     | 5 days            | PAGE   |
  | 6 hours | 6x        | 12 days           | TICKET |
  | 3 days  | 1x        | 30 days           | REVIEW |
```

### 4.2 SLI Types & Measurement

```
Availability SLI:
  SLI = requests_with_status < 500 / total_requests

Latency SLI (p99):
  SLI = requests_with_latency < threshold / total_requests

Error Rate SLI:
  SLI = 1 - (error_requests / total_requests)

Throughput SLI:
  SLI = min(actual_rps / expected_rps, 1.0)
```

---

## 5. Capacity Forecasting

### 5.1 Model Architecture

```
Dual-Model Ensemble:

Model 1: Facebook Prophet (Seasonality + Trend)
  - Daily, weekly, monthly seasonality components
  - Holiday and known-event handling
  - Uncertainty intervals via MCMC

Model 2: XGBoost (Event-Driven Spikes)
  - Features: deployments, traffic events, business events
  - Non-seasonal demand changes

Ensemble:
  forecast = alpha * prophet + (1-alpha) * xgboost
  alpha dynamically weighted by recent prediction accuracy

Output per resource per service:
  {
    "resource": "cpu",
    "service": "payment-api",
    "forecast_72h": [
      {"timestamp": "...", "predicted": 0.72, "lower": 0.65, "upper": 0.79}
    ],
    "exhaustion_warning": {
      "predicted_breach": "2026-03-05T14:00Z",
      "confidence": 0.91,
      "recommended_action": "Scale from 4 to 6 replicas"
    }
  }
```

---

## 6. Event Processing Pipeline -- Detailed Flow

```
1. HTTP/Webhook received at Ingestion Gateway (Go, net/http)
2. Source adapter extracts fields (one adapter per monitoring source)
3. Normalize to Common Event Format (CEF):
   {
     "id": "uuid-v7",
     "tenant_id": "tenant_abc",
     "source": "prometheus",
     "severity": "warning",
     "service": "payment-api",
     "environment": "production",
     "timestamp": "2026-03-03T10:15:30Z",
     "message": "High latency detected",
     "labels": {"pod": "payment-api-7b9c4", "namespace": "production"},
     "metrics": {"latency_p99_ms": 1250, "error_rate": 0.032}
   }
4. Publish to Kafka topic: events.{type}.{tenant_id}
5. Anomaly Detection consumer: ML inference
6. If anomalous: publish to anomalies.detected
7. Correlation Engine: temporal window + topology lookup + Bayesian scoring
8. If correlated: create/update incident in PostgreSQL
9. Incident triggers: runbook matching -> approval gate -> execution
10. Notification Router sends alerts per escalation policy
11. All state changes written to audit log
```

---

## 7. API Contracts

### 7.1 Key GraphQL Mutations

```graphql
mutation CreateRunbook($input: RunbookInput!) {
  insert_runbooks_one(object: $input) {
    id name mode trigger_conditions steps guardrails
  }
}

mutation ExecuteRemediation($incidentId: UUID!, $runbookId: UUID!, $mode: RemediationMode!) {
  execute_remediation(incident_id: $incidentId, runbook_id: $runbookId, mode: $mode) {
    execution_id status steps_planned approval_required
  }
}

mutation DefineSLO($input: SLOInput!) {
  insert_slo_definitions_one(object: $input) {
    id service_id sli_type target window_days burn_rate_alerts
  }
}
```

### 7.2 Key GraphQL Queries

```graphql
query GetIncidentTimeline($incidentId: UUID!) {
  incidents_by_pk(id: $incidentId) {
    id severity status root_cause_service root_cause_confidence
    correlated_events(order_by: {timestamp: asc}) {
      id source service message severity timestamp
    }
    remediation_executions(order_by: {started_at: desc}) {
      id runbook { name } status
      steps { name status output started_at completed_at }
    }
    timeline_entries(order_by: {timestamp: asc}) {
      timestamp type description actor
    }
  }
}

query GetServiceTopology($tenantId: String!, $env: String!) {
  topology_nodes(where: {tenant_id: {_eq: $tenantId}, environment: {_eq: $env}}) {
    id service_name type health_status
    edges_out {
      target_node { service_name type health_status }
      edge_type latency_p99
    }
  }
}
```

---

## 8. Database Indexes & Performance

```sql
-- Hot path: event ingestion lookups
CREATE INDEX idx_events_tenant_ts ON events(tenant_id, event_timestamp DESC);
CREATE INDEX idx_events_fingerprint ON events(fingerprint);
CREATE INDEX idx_events_service_ts ON events(tenant_id, service_id, event_timestamp DESC);

-- Anomaly detection
CREATE INDEX idx_anomalies_tenant_detected ON anomalies(tenant_id, detected_at DESC);
CREATE INDEX idx_anomalies_service ON anomalies(tenant_id, service_id, detected_at DESC);

-- Incident lookup
CREATE INDEX idx_incidents_tenant_status ON incidents(tenant_id, status, created_at DESC);
CREATE INDEX idx_incidents_tenant_severity ON incidents(tenant_id, severity, created_at DESC);

-- Topology traversal
CREATE INDEX idx_topo_edges_source ON topology_edges(source_node_id);
CREATE INDEX idx_topo_edges_target ON topology_edges(target_node_id);
CREATE INDEX idx_topo_nodes_tenant_env ON topology_nodes(tenant_id, environment);

-- SLO queries
CREATE INDEX idx_slo_measurements_slo_time ON slo_measurements(slo_id, measurement_time DESC);
```

---

*Document Control: Technical design changes require Architecture Review Board approval and LLD version increment.*

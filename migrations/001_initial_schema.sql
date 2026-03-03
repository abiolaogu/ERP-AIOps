-- ERP-AIOps: Autonomous IT Operations Database Schema
-- Migration: 001_initial_schema.sql
-- Tables: 16 (7 core + 9 cross-module integration)

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- CORE AIOPS TABLES (7)
-- ============================================================

-- 1. incidents: IT incidents tracked by AIOps
CREATE TABLE incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    severity TEXT NOT NULL CHECK (severity IN ('critical', 'high', 'medium', 'low', 'info')),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'investigating', 'mitigating', 'resolved', 'closed')),
    source_module TEXT NOT NULL,
    source_event_id TEXT,
    assigned_to TEXT,
    escalation_policy_id UUID,
    root_cause TEXT,
    resolution TEXT,
    acknowledged_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 2. anomalies: Detected anomalies from metric/log analysis
CREATE TABLE anomalies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    anomaly_type TEXT NOT NULL CHECK (anomaly_type IN ('metric', 'log', 'trace', 'behavioral', 'capacity')),
    source_module TEXT NOT NULL,
    metric_name TEXT,
    expected_value DOUBLE PRECISION,
    actual_value DOUBLE PRECISION,
    deviation_score DOUBLE PRECISION NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('critical', 'high', 'medium', 'low')),
    status TEXT NOT NULL DEFAULT 'detected' CHECK (status IN ('detected', 'confirmed', 'investigating', 'resolved', 'false_positive')),
    correlated_incident_id UUID REFERENCES incidents(id),
    detection_model TEXT NOT NULL,
    confidence DOUBLE PRECISION NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    raw_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 3. topology_nodes: Service dependency graph nodes
CREATE TABLE topology_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    node_type TEXT NOT NULL CHECK (node_type IN ('service', 'database', 'cache', 'queue', 'gateway', 'external')),
    module_name TEXT NOT NULL,
    service_name TEXT NOT NULL,
    namespace TEXT NOT NULL,
    version TEXT,
    health_status TEXT NOT NULL DEFAULT 'unknown' CHECK (health_status IN ('healthy', 'degraded', 'critical', 'unknown')),
    endpoint_url TEXT,
    metadata JSONB DEFAULT '{}',
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, module_name, service_name)
);

-- 4. topology_edges: Service dependency graph edges
CREATE TABLE topology_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    source_node_id UUID NOT NULL REFERENCES topology_nodes(id) ON DELETE CASCADE,
    target_node_id UUID NOT NULL REFERENCES topology_nodes(id) ON DELETE CASCADE,
    edge_type TEXT NOT NULL CHECK (edge_type IN ('http', 'grpc', 'tcp', 'kafka', 'database', 'cache')),
    protocol TEXT,
    avg_latency_ms DOUBLE PRECISION,
    error_rate DOUBLE PRECISION,
    requests_per_sec DOUBLE PRECISION,
    metadata JSONB DEFAULT '{}',
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, source_node_id, target_node_id, edge_type)
);

-- 5. correlation_rules: Rules for correlating incidents/anomalies
CREATE TABLE correlation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    rule_type TEXT NOT NULL CHECK (rule_type IN ('temporal', 'topological', 'pattern', 'causal', 'statistical')),
    condition_expr JSONB NOT NULL,
    time_window_seconds INTEGER NOT NULL DEFAULT 300,
    min_confidence DOUBLE PRECISION NOT NULL DEFAULT 0.7,
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 50,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 6. automation_playbooks: Predefined remediation workflows
CREATE TABLE automation_playbooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    trigger_conditions JSONB NOT NULL,
    steps JSONB NOT NULL,
    guardrail_tier TEXT NOT NULL CHECK (guardrail_tier IN ('autonomous', 'supervised', 'protected')),
    risk_score INTEGER NOT NULL CHECK (risk_score >= 1 AND risk_score <= 10),
    target_modules TEXT[] NOT NULL DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    max_executions_per_hour INTEGER NOT NULL DEFAULT 5,
    cooldown_seconds INTEGER NOT NULL DEFAULT 300,
    rollback_steps JSONB,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 7. playbook_executions: Execution history of playbooks
CREATE TABLE playbook_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    playbook_id UUID NOT NULL REFERENCES automation_playbooks(id),
    incident_id UUID REFERENCES incidents(id),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'rolled_back', 'cancelled')),
    triggered_by TEXT NOT NULL CHECK (triggered_by IN ('auto', 'manual', 'escalation')),
    target_module TEXT NOT NULL,
    steps_completed JSONB DEFAULT '[]',
    steps_total INTEGER NOT NULL,
    current_step INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    execution_duration_ms BIGINT,
    approvals JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- CROSS-MODULE INTEGRATION TABLES (9)
-- ============================================================

-- 8. module_health_status: Real-time health of all 24 modules
CREATE TABLE module_health_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    module_name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('healthy', 'degraded', 'critical', 'unknown', 'maintenance')),
    gateway_healthy BOOLEAN NOT NULL DEFAULT true,
    hasura_healthy BOOLEAN NOT NULL DEFAULT true,
    database_healthy BOOLEAN NOT NULL DEFAULT true,
    latency_p50_ms DOUBLE PRECISION,
    latency_p95_ms DOUBLE PRECISION,
    latency_p99_ms DOUBLE PRECISION,
    error_rate DOUBLE PRECISION DEFAULT 0,
    request_rate DOUBLE PRECISION DEFAULT 0,
    pod_count INTEGER DEFAULT 0,
    pod_ready_count INTEGER DEFAULT 0,
    cpu_usage_pct DOUBLE PRECISION,
    memory_usage_pct DOUBLE PRECISION,
    last_heartbeat_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_incident_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, module_name)
);

-- 9. operational_metrics: Aggregated metrics from all modules
CREATE TABLE operational_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    module_name TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    metric_type TEXT NOT NULL CHECK (metric_type IN ('counter', 'gauge', 'histogram', 'summary')),
    value DOUBLE PRECISION NOT NULL,
    unit TEXT,
    dimensions JSONB DEFAULT '{}',
    aggregation_window TEXT NOT NULL DEFAULT '1m',
    sample_count INTEGER DEFAULT 1,
    min_value DOUBLE PRECISION,
    max_value DOUBLE PRECISION,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 10. guardrail_evaluations: AIDD guardrail decision audit trail
CREATE TABLE guardrail_evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    action_type TEXT NOT NULL,
    target_module TEXT NOT NULL,
    guardrail_tier TEXT NOT NULL CHECK (guardrail_tier IN ('autonomous', 'supervised', 'protected')),
    risk_score INTEGER NOT NULL CHECK (risk_score >= 1 AND risk_score <= 10),
    result TEXT NOT NULL CHECK (result IN ('approved', 'pending_approval', 'denied', 'timeout', 'escalated')),
    requested_by TEXT NOT NULL,
    approvals_required INTEGER NOT NULL DEFAULT 0,
    approvals_received INTEGER NOT NULL DEFAULT 0,
    approval_chain JSONB DEFAULT '[]',
    denial_reason TEXT,
    evaluation_duration_ms INTEGER,
    context JSONB DEFAULT '{}',
    resolved_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 11. slo_tracking: SLO definitions and current status per module
CREATE TABLE slo_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    module_name TEXT NOT NULL,
    slo_name TEXT NOT NULL,
    slo_type TEXT NOT NULL CHECK (slo_type IN ('availability', 'latency', 'error_rate', 'throughput', 'saturation', 'custom')),
    target_value DOUBLE PRECISION NOT NULL,
    current_value DOUBLE PRECISION,
    unit TEXT NOT NULL,
    comparison TEXT NOT NULL CHECK (comparison IN ('gte', 'lte', 'eq')),
    status TEXT NOT NULL DEFAULT 'met' CHECK (status IN ('met', 'at_risk', 'breached', 'unknown')),
    error_budget_total DOUBLE PRECISION NOT NULL DEFAULT 100,
    error_budget_remaining DOUBLE PRECISION NOT NULL DEFAULT 100,
    error_budget_burn_rate DOUBLE PRECISION DEFAULT 0,
    evaluation_window TEXT NOT NULL DEFAULT '30d',
    prometheus_query TEXT,
    last_evaluated_at TIMESTAMPTZ,
    breached_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, module_name, slo_name)
);

-- 12. aiops_audit_log: Immutable audit log for all AIOps decisions
CREATE TABLE aiops_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    action TEXT NOT NULL,
    action_category TEXT NOT NULL CHECK (action_category IN ('detection', 'correlation', 'remediation', 'escalation', 'guardrail', 'notification', 'configuration', 'suppression')),
    actor TEXT NOT NULL,
    actor_type TEXT NOT NULL CHECK (actor_type IN ('system', 'user', 'automation', 'api')),
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    target_module TEXT,
    previous_state JSONB,
    new_state JSONB,
    result TEXT NOT NULL CHECK (result IN ('success', 'failure', 'partial', 'skipped')),
    error_message TEXT,
    ip_address TEXT,
    user_agent TEXT,
    correlation_id TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    -- NOTE: No updated_at — this table is append-only / immutable
);

-- 13. notification_channels: Alert destinations
CREATE TABLE notification_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    channel_type TEXT NOT NULL CHECK (channel_type IN ('webhook', 'email', 'slack', 'pagerduty', 'opsgenie', 'teams', 'discord', 'sms')),
    config JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    severity_filter TEXT[] NOT NULL DEFAULT '{critical,high,medium,low}',
    module_filter TEXT[] DEFAULT NULL,
    rate_limit_per_hour INTEGER NOT NULL DEFAULT 100,
    last_sent_at TIMESTAMPTZ,
    failure_count INTEGER NOT NULL DEFAULT 0,
    last_failure_at TIMESTAMPTZ,
    last_failure_reason TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 14. escalation_policies: Escalation chains with step delays
CREATE TABLE escalation_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    steps JSONB NOT NULL,
    repeat_enabled BOOLEAN NOT NULL DEFAULT false,
    repeat_limit INTEGER NOT NULL DEFAULT 3,
    default_for_severities TEXT[] DEFAULT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 15. maintenance_windows: Scheduled suppression windows
CREATE TABLE maintenance_windows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    target_modules TEXT[] NOT NULL,
    suppress_alerts BOOLEAN NOT NULL DEFAULT true,
    suppress_remediation BOOLEAN NOT NULL DEFAULT true,
    suppress_notifications BOOLEAN NOT NULL DEFAULT false,
    schedule_type TEXT NOT NULL CHECK (schedule_type IN ('one_time', 'recurring')),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    recurrence_rule TEXT,
    status TEXT NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'active', 'completed', 'cancelled')),
    created_by TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 16. runbooks: Automated/manual runbooks with trigger conditions
CREATE TABLE runbooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    runbook_type TEXT NOT NULL CHECK (runbook_type IN ('automated', 'semi_automated', 'manual')),
    trigger_conditions JSONB NOT NULL,
    steps JSONB NOT NULL,
    guardrail_tier TEXT NOT NULL CHECK (guardrail_tier IN ('autonomous', 'supervised', 'protected')),
    risk_score INTEGER NOT NULL CHECK (risk_score >= 1 AND risk_score <= 10),
    target_modules TEXT[] NOT NULL DEFAULT '{}',
    estimated_duration_seconds INTEGER,
    success_criteria JSONB,
    rollback_steps JSONB,
    enabled BOOLEAN NOT NULL DEFAULT true,
    max_concurrent_executions INTEGER NOT NULL DEFAULT 1,
    last_executed_at TIMESTAMPTZ,
    execution_count INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    failure_count INTEGER NOT NULL DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- ROW-LEVEL SECURITY
-- ============================================================

ALTER TABLE incidents ENABLE ROW LEVEL SECURITY;
ALTER TABLE anomalies ENABLE ROW LEVEL SECURITY;
ALTER TABLE topology_nodes ENABLE ROW LEVEL SECURITY;
ALTER TABLE topology_edges ENABLE ROW LEVEL SECURITY;
ALTER TABLE correlation_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE automation_playbooks ENABLE ROW LEVEL SECURITY;
ALTER TABLE playbook_executions ENABLE ROW LEVEL SECURITY;
ALTER TABLE module_health_status ENABLE ROW LEVEL SECURITY;
ALTER TABLE operational_metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE guardrail_evaluations ENABLE ROW LEVEL SECURITY;
ALTER TABLE slo_tracking ENABLE ROW LEVEL SECURITY;
ALTER TABLE aiops_audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_channels ENABLE ROW LEVEL SECURITY;
ALTER TABLE escalation_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE maintenance_windows ENABLE ROW LEVEL SECURITY;
ALTER TABLE runbooks ENABLE ROW LEVEL SECURITY;

-- RLS policies: tenant isolation
CREATE POLICY tenant_isolation ON incidents USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON anomalies USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON topology_nodes USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON topology_edges USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON correlation_rules USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON automation_playbooks USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON playbook_executions USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON module_health_status USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON operational_metrics USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON guardrail_evaluations USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON slo_tracking USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON aiops_audit_log USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON notification_channels USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON escalation_policies USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON maintenance_windows USING (tenant_id = current_setting('app.tenant_id'));
CREATE POLICY tenant_isolation ON runbooks USING (tenant_id = current_setting('app.tenant_id'));

-- ============================================================
-- INDEXES
-- ============================================================

-- incidents
CREATE INDEX idx_incidents_tenant ON incidents (tenant_id);
CREATE INDEX idx_incidents_status ON incidents (tenant_id, status);
CREATE INDEX idx_incidents_severity ON incidents (tenant_id, severity);
CREATE INDEX idx_incidents_source ON incidents (tenant_id, source_module);
CREATE INDEX idx_incidents_created ON incidents (tenant_id, created_at DESC);

-- anomalies
CREATE INDEX idx_anomalies_tenant ON anomalies (tenant_id);
CREATE INDEX idx_anomalies_status ON anomalies (tenant_id, status);
CREATE INDEX idx_anomalies_source ON anomalies (tenant_id, source_module);
CREATE INDEX idx_anomalies_severity ON anomalies (tenant_id, severity);
CREATE INDEX idx_anomalies_incident ON anomalies (correlated_incident_id);

-- topology
CREATE INDEX idx_topology_nodes_tenant ON topology_nodes (tenant_id);
CREATE INDEX idx_topology_nodes_module ON topology_nodes (tenant_id, module_name);
CREATE INDEX idx_topology_nodes_health ON topology_nodes (tenant_id, health_status);
CREATE INDEX idx_topology_edges_tenant ON topology_edges (tenant_id);
CREATE INDEX idx_topology_edges_source ON topology_edges (source_node_id);
CREATE INDEX idx_topology_edges_target ON topology_edges (target_node_id);

-- correlation_rules
CREATE INDEX idx_correlation_rules_tenant ON correlation_rules (tenant_id);
CREATE INDEX idx_correlation_rules_enabled ON correlation_rules (tenant_id, enabled);

-- automation_playbooks
CREATE INDEX idx_playbooks_tenant ON automation_playbooks (tenant_id);
CREATE INDEX idx_playbooks_tier ON automation_playbooks (tenant_id, guardrail_tier);
CREATE INDEX idx_playbooks_enabled ON automation_playbooks (tenant_id, enabled);

-- playbook_executions
CREATE INDEX idx_executions_tenant ON playbook_executions (tenant_id);
CREATE INDEX idx_executions_playbook ON playbook_executions (playbook_id);
CREATE INDEX idx_executions_incident ON playbook_executions (incident_id);
CREATE INDEX idx_executions_status ON playbook_executions (tenant_id, status);
CREATE INDEX idx_executions_created ON playbook_executions (tenant_id, created_at DESC);

-- module_health_status
CREATE INDEX idx_health_tenant ON module_health_status (tenant_id);
CREATE INDEX idx_health_module ON module_health_status (tenant_id, module_name);
CREATE INDEX idx_health_status ON module_health_status (tenant_id, status);
CREATE INDEX idx_health_heartbeat ON module_health_status (last_heartbeat_at);

-- operational_metrics
CREATE INDEX idx_metrics_tenant ON operational_metrics (tenant_id);
CREATE INDEX idx_metrics_module ON operational_metrics (tenant_id, module_name);
CREATE INDEX idx_metrics_name ON operational_metrics (tenant_id, metric_name);
CREATE INDEX idx_metrics_recorded ON operational_metrics (tenant_id, recorded_at DESC);
CREATE INDEX idx_metrics_module_name ON operational_metrics (tenant_id, module_name, metric_name, recorded_at DESC);

-- guardrail_evaluations
CREATE INDEX idx_guardrail_tenant ON guardrail_evaluations (tenant_id);
CREATE INDEX idx_guardrail_result ON guardrail_evaluations (tenant_id, result);
CREATE INDEX idx_guardrail_tier ON guardrail_evaluations (tenant_id, guardrail_tier);
CREATE INDEX idx_guardrail_target ON guardrail_evaluations (tenant_id, target_module);
CREATE INDEX idx_guardrail_pending ON guardrail_evaluations (tenant_id, result) WHERE result = 'pending_approval';

-- slo_tracking
CREATE INDEX idx_slo_tenant ON slo_tracking (tenant_id);
CREATE INDEX idx_slo_module ON slo_tracking (tenant_id, module_name);
CREATE INDEX idx_slo_status ON slo_tracking (tenant_id, status);
CREATE INDEX idx_slo_breached ON slo_tracking (tenant_id, status) WHERE status = 'breached';

-- aiops_audit_log
CREATE INDEX idx_audit_tenant ON aiops_audit_log (tenant_id);
CREATE INDEX idx_audit_action ON aiops_audit_log (tenant_id, action);
CREATE INDEX idx_audit_actor ON aiops_audit_log (tenant_id, actor);
CREATE INDEX idx_audit_target ON aiops_audit_log (tenant_id, target_type, target_id);
CREATE INDEX idx_audit_category ON aiops_audit_log (tenant_id, action_category);
CREATE INDEX idx_audit_created ON aiops_audit_log (tenant_id, created_at DESC);
CREATE INDEX idx_audit_correlation ON aiops_audit_log (correlation_id) WHERE correlation_id IS NOT NULL;

-- notification_channels
CREATE INDEX idx_channels_tenant ON notification_channels (tenant_id);
CREATE INDEX idx_channels_type ON notification_channels (tenant_id, channel_type);
CREATE INDEX idx_channels_enabled ON notification_channels (tenant_id, enabled);

-- escalation_policies
CREATE INDEX idx_escalation_tenant ON escalation_policies (tenant_id);
CREATE INDEX idx_escalation_enabled ON escalation_policies (tenant_id, enabled);

-- maintenance_windows
CREATE INDEX idx_maintenance_tenant ON maintenance_windows (tenant_id);
CREATE INDEX idx_maintenance_status ON maintenance_windows (tenant_id, status);
CREATE INDEX idx_maintenance_active ON maintenance_windows (tenant_id, starts_at, ends_at) WHERE status IN ('scheduled', 'active');

-- runbooks
CREATE INDEX idx_runbooks_tenant ON runbooks (tenant_id);
CREATE INDEX idx_runbooks_tier ON runbooks (tenant_id, guardrail_tier);
CREATE INDEX idx_runbooks_enabled ON runbooks (tenant_id, enabled);
CREATE INDEX idx_runbooks_type ON runbooks (tenant_id, runbook_type);

-- ============================================================
-- UPDATED_AT TRIGGERS
-- ============================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_incidents_updated_at BEFORE UPDATE ON incidents FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_anomalies_updated_at BEFORE UPDATE ON anomalies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_topology_nodes_updated_at BEFORE UPDATE ON topology_nodes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_topology_edges_updated_at BEFORE UPDATE ON topology_edges FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_correlation_rules_updated_at BEFORE UPDATE ON correlation_rules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_playbooks_updated_at BEFORE UPDATE ON automation_playbooks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_executions_updated_at BEFORE UPDATE ON playbook_executions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_health_updated_at BEFORE UPDATE ON module_health_status FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_guardrail_updated_at BEFORE UPDATE ON guardrail_evaluations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_slo_updated_at BEFORE UPDATE ON slo_tracking FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_channels_updated_at BEFORE UPDATE ON notification_channels FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_escalation_updated_at BEFORE UPDATE ON escalation_policies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_maintenance_updated_at BEFORE UPDATE ON maintenance_windows FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_runbooks_updated_at BEFORE UPDATE ON runbooks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- NOTE: aiops_audit_log has no update trigger — it is append-only

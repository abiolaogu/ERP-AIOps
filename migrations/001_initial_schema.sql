-- ERP-AIOps Initial Schema
-- Database: YugabyteDB (PostgreSQL-compatible)

-- Incidents table
CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    severity TEXT NOT NULL DEFAULT 'medium',
    status TEXT NOT NULL DEFAULT 'open',
    source TEXT,
    affected_services TEXT[],
    root_cause TEXT,
    correlation_id UUID,
    acknowledged_by TEXT,
    resolved_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    acknowledged_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_incidents_tenant ON incidents(tenant_id);
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
CREATE INDEX IF NOT EXISTS idx_incidents_severity ON incidents(severity);

-- Anomalies table
CREATE TABLE IF NOT EXISTS anomalies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    service TEXT NOT NULL,
    module TEXT,
    anomaly_type TEXT NOT NULL DEFAULT 'spike',
    severity TEXT NOT NULL DEFAULT 'medium',
    expected_value DOUBLE PRECISION,
    actual_value DOUBLE PRECISION,
    deviation_percent DOUBLE PRECISION,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_anomalies_tenant ON anomalies(tenant_id);
CREATE INDEX IF NOT EXISTS idx_anomalies_service ON anomalies(service);

-- AIOps Rules
CREATE TABLE IF NOT EXISTS aiops_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL,
    condition JSONB NOT NULL,
    action JSONB NOT NULL,
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_rules_tenant ON aiops_rules(tenant_id);

-- Topology Nodes
CREATE TABLE IF NOT EXISTS topology_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    module TEXT,
    status TEXT DEFAULT 'healthy',
    metadata JSONB DEFAULT '{}',
    dependencies UUID[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_topology_tenant ON topology_nodes(tenant_id);

-- Remediation Actions
CREATE TABLE IF NOT EXISTS remediation_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    incident_id UUID REFERENCES incidents(id),
    action_type TEXT NOT NULL,
    target_service TEXT NOT NULL,
    parameters JSONB DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'pending',
    result JSONB,
    initiated_by TEXT,
    initiated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_remediation_tenant ON remediation_actions(tenant_id);

-- Cost Reports
CREATE TABLE IF NOT EXISTS cost_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    total_cost DOUBLE PRECISION,
    breakdown JSONB DEFAULT '{}',
    recommendations JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cost_tenant ON cost_reports(tenant_id);

-- Security Findings
CREATE TABLE IF NOT EXISTS security_findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    severity TEXT NOT NULL DEFAULT 'medium',
    category TEXT NOT NULL,
    affected_resource TEXT,
    status TEXT NOT NULL DEFAULT 'open',
    remediation TEXT,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_security_tenant ON security_findings(tenant_id);

-- =====================================================
-- AIOps Autonomous Operations Tables
-- =====================================================

-- Module Health Status: real-time health of all 24 modules
CREATE TABLE IF NOT EXISTS module_health_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    module_name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'healthy',
    latency_ms DOUBLE PRECISION,
    error_rate DOUBLE PRECISION,
    pod_count INTEGER DEFAULT 0,
    last_heartbeat_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, module_name)
);
CREATE INDEX IF NOT EXISTS idx_module_health_tenant ON module_health_status(tenant_id);
CREATE INDEX IF NOT EXISTS idx_module_health_status ON module_health_status(status);
CREATE INDEX IF NOT EXISTS idx_module_health_module ON module_health_status(module_name);

-- Operational Metrics: aggregated metrics from all modules
CREATE TABLE IF NOT EXISTS operational_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    module_name TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    unit TEXT,
    dimensions JSONB DEFAULT '{}',
    collected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_op_metrics_tenant ON operational_metrics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_op_metrics_module ON operational_metrics(module_name);
CREATE INDEX IF NOT EXISTS idx_op_metrics_name ON operational_metrics(metric_name);
CREATE INDEX IF NOT EXISTS idx_op_metrics_collected ON operational_metrics(collected_at DESC);

-- Guardrail Evaluations: AIDD guardrail decision audit trail
CREATE TABLE IF NOT EXISTS guardrail_evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    action_type TEXT NOT NULL,
    target_module TEXT,
    tier TEXT NOT NULL,
    risk_score INTEGER NOT NULL DEFAULT 0,
    result TEXT NOT NULL DEFAULT 'pending',
    approval_chain JSONB DEFAULT '[]',
    requested_by TEXT NOT NULL,
    approved_by TEXT,
    reason TEXT,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_guardrail_eval_tenant ON guardrail_evaluations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_guardrail_eval_tier ON guardrail_evaluations(tier);
CREATE INDEX IF NOT EXISTS idx_guardrail_eval_result ON guardrail_evaluations(result);

-- SLO Tracking: SLO definitions + current status per module
CREATE TABLE IF NOT EXISTS slo_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    module_name TEXT NOT NULL,
    slo_name TEXT NOT NULL,
    slo_type TEXT NOT NULL,
    target DOUBLE PRECISION NOT NULL,
    current_value DOUBLE PRECISION,
    error_budget_total DOUBLE PRECISION,
    error_budget_remaining DOUBLE PRECISION,
    status TEXT NOT NULL DEFAULT 'met',
    window_days INTEGER NOT NULL DEFAULT 30,
    prometheus_query TEXT,
    last_evaluated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, module_name, slo_name)
);
CREATE INDEX IF NOT EXISTS idx_slo_tracking_tenant ON slo_tracking(tenant_id);
CREATE INDEX IF NOT EXISTS idx_slo_tracking_module ON slo_tracking(module_name);
CREATE INDEX IF NOT EXISTS idx_slo_tracking_status ON slo_tracking(status);

-- AIOps Audit Log: immutable audit log for all AIOps decisions
CREATE TABLE IF NOT EXISTS aiops_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    action TEXT NOT NULL,
    actor TEXT NOT NULL,
    target_module TEXT,
    target_resource TEXT,
    decision TEXT NOT NULL,
    tier TEXT,
    risk_score INTEGER,
    details JSONB DEFAULT '{}',
    correlation_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_log_tenant ON aiops_audit_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON aiops_audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_log_created ON aiops_audit_log(created_at DESC);

-- Notification Channels: alert destinations
CREATE TABLE IF NOT EXISTS notification_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    channel_type TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_notification_channels_tenant ON notification_channels(tenant_id);

-- Escalation Policies: escalation chains with step delays and assignees
CREATE TABLE IF NOT EXISTS escalation_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    steps JSONB NOT NULL DEFAULT '[]',
    repeat_count INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_escalation_policies_tenant ON escalation_policies(tenant_id);

-- Maintenance Windows: scheduled suppression windows for alerts/remediation
CREATE TABLE IF NOT EXISTS maintenance_windows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    target_modules TEXT[] DEFAULT '{}',
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    suppress_alerts BOOLEAN DEFAULT true,
    suppress_remediation BOOLEAN DEFAULT true,
    created_by TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'scheduled',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_maintenance_windows_tenant ON maintenance_windows(tenant_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_windows_time ON maintenance_windows(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_maintenance_windows_status ON maintenance_windows(status);

-- Runbooks: automated/manual runbooks with trigger conditions and risk tiers
CREATE TABLE IF NOT EXISTS runbooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    trigger_conditions JSONB DEFAULT '{}',
    steps JSONB NOT NULL DEFAULT '[]',
    risk_tier TEXT NOT NULL DEFAULT 'supervised',
    auto_execute BOOLEAN DEFAULT false,
    target_modules TEXT[] DEFAULT '{}',
    rollback_steps JSONB DEFAULT '[]',
    last_executed_at TIMESTAMPTZ,
    execution_count INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_runbooks_tenant ON runbooks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_runbooks_risk_tier ON runbooks(risk_tier);

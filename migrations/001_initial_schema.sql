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

export interface Incident {
  id: string;
  tenant_id: string;
  title: string;
  description?: string;
  severity: string;
  status: string;
  source?: string;
  affected_services?: string[];
  root_cause?: string;
  correlation_id?: string;
  acknowledged_by?: string;
  resolved_by?: string;
  created_at: string;
  acknowledged_at?: string;
  resolved_at?: string;
  updated_at: string;
}

export interface Anomaly {
  id: string;
  tenant_id: string;
  metric_name: string;
  service: string;
  module?: string;
  anomaly_type: string;
  severity: string;
  expected_value?: number;
  actual_value?: number;
  deviation_percent?: number;
  detected_at: string;
  resolved_at?: string;
  status: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface Rule {
  id: string;
  tenant_id: string;
  name: string;
  description?: string;
  type: string;
  condition: Record<string, unknown>;
  action: Record<string, unknown>;
  enabled: boolean;
  priority: number;
  created_at: string;
  updated_at: string;
}

export interface TopologyNode {
  id: string;
  tenant_id: string;
  name: string;
  type: string;
  module?: string;
  status?: string;
  metadata?: Record<string, unknown>;
  dependencies?: string[];
  created_at: string;
  updated_at: string;
}

export interface RemediationAction {
  id: string;
  tenant_id: string;
  incident_id?: string;
  action_type: string;
  target_service: string;
  parameters?: Record<string, unknown>;
  status: string;
  result?: Record<string, unknown>;
  initiated_by?: string;
  initiated_at: string;
  completed_at?: string;
}

export interface CostReport {
  id: string;
  tenant_id: string;
  period_start: string;
  period_end: string;
  total_cost?: number;
  breakdown?: Record<string, unknown>;
  recommendations?: Record<string, unknown>[];
  created_at: string;
}

export interface SecurityFinding {
  id: string;
  tenant_id: string;
  title: string;
  description?: string;
  severity: string;
  category: string;
  affected_resource?: string;
  status: string;
  remediation?: string;
  detected_at: string;
  resolved_at?: string;
}

// Status color mappings

export const SEVERITY_COLORS: Record<string, string> = {
  critical: "#ef4444",
  high: "#f97316",
  medium: "#f59e0b",
  low: "#3b82f6",
  info: "#94a3b8",
};

export const INCIDENT_STATUS_COLORS: Record<string, string> = {
  open: "#ef4444",
  acknowledged: "#f59e0b",
  investigating: "#3b82f6",
  resolved: "#10b981",
  closed: "#64748b",
};

export const ANOMALY_STATUS_COLORS: Record<string, string> = {
  active: "#ef4444",
  investigating: "#f59e0b",
  resolved: "#10b981",
  dismissed: "#94a3b8",
};

export const REMEDIATION_STATUS_COLORS: Record<string, string> = {
  pending: "#f59e0b",
  running: "#3b82f6",
  completed: "#10b981",
  failed: "#ef4444",
  cancelled: "#94a3b8",
};

export const TOPOLOGY_STATUS_COLORS: Record<string, string> = {
  healthy: "#10b981",
  degraded: "#f59e0b",
  down: "#ef4444",
  unknown: "#94a3b8",
};

export const SECURITY_CATEGORY_COLORS: Record<string, string> = {
  vulnerability: "#ef4444",
  misconfiguration: "#f97316",
  exposed_secret: "#dc2626",
  weak_authentication: "#f59e0b",
  network_exposure: "#3b82f6",
  compliance_violation: "#8b5cf6",
  data_leakage: "#ec4899",
  privilege_escalation: "#ef4444",
};

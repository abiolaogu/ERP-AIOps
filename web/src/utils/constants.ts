export const API_URL =
  import.meta.env.VITE_GRAPHQL_URL || "http://localhost:8090/v1/graphql";
export const IAM_URL =
  import.meta.env.VITE_IAM_URL || "http://localhost:8081";
export const GATEWAY_URL =
  import.meta.env.VITE_GATEWAY_URL || "http://localhost:8090";
export const TENANT_ID = import.meta.env.VITE_TENANT_ID || "default";

export const TOKEN_KEY = "erp_aiops_auth_token";
export const REFRESH_TOKEN_KEY = "erp_aiops_refresh_token";
export const USER_KEY = "erp_aiops_user";

export const DEFAULT_PAGE_SIZE = 20;
export const PAGE_SIZE_OPTIONS = [10, 20, 50, 100];

export const INCIDENT_SEVERITIES = [
  { label: "Critical", value: "critical", color: "#ef4444" },
  { label: "High", value: "high", color: "#f97316" },
  { label: "Medium", value: "medium", color: "#f59e0b" },
  { label: "Low", value: "low", color: "#3b82f6" },
  { label: "Info", value: "info", color: "#94a3b8" },
] as const;

export const INCIDENT_STATUSES = [
  { label: "Open", value: "open", color: "#ef4444" },
  { label: "Acknowledged", value: "acknowledged", color: "#f59e0b" },
  { label: "Investigating", value: "investigating", color: "#3b82f6" },
  { label: "Resolved", value: "resolved", color: "#10b981" },
  { label: "Closed", value: "closed", color: "#64748b" },
] as const;

export const ANOMALY_TYPES = [
  { label: "Spike", value: "spike" },
  { label: "Dip", value: "dip" },
  { label: "Trend Change", value: "trend_change" },
  { label: "Pattern Break", value: "pattern_break" },
  { label: "Seasonal Anomaly", value: "seasonal" },
] as const;

export const ANOMALY_STATUSES = [
  { label: "Active", value: "active", color: "#ef4444" },
  { label: "Investigating", value: "investigating", color: "#f59e0b" },
  { label: "Resolved", value: "resolved", color: "#10b981" },
  { label: "Dismissed", value: "dismissed", color: "#94a3b8" },
] as const;

export const RULE_TYPES = [
  { label: "Alert", value: "alert" },
  { label: "Correlation", value: "correlation" },
  { label: "Auto-Remediation", value: "auto_remediation" },
  { label: "Escalation", value: "escalation" },
  { label: "Suppression", value: "suppression" },
] as const;

export const REMEDIATION_ACTION_TYPES = [
  { label: "Restart Service", value: "restart_service" },
  { label: "Scale Up", value: "scale_up" },
  { label: "Scale Down", value: "scale_down" },
  { label: "Rollback Config", value: "rollback_config" },
  { label: "Clear Cache", value: "clear_cache" },
  { label: "Failover Primary", value: "failover_primary" },
  { label: "Run Playbook", value: "run_playbook" },
  { label: "Notify On-Call", value: "notify_oncall" },
] as const;

export const REMEDIATION_STATUSES = [
  { label: "Pending", value: "pending", color: "#f59e0b" },
  { label: "Running", value: "running", color: "#3b82f6" },
  { label: "Completed", value: "completed", color: "#10b981" },
  { label: "Failed", value: "failed", color: "#ef4444" },
  { label: "Cancelled", value: "cancelled", color: "#94a3b8" },
] as const;

export const TOPOLOGY_NODE_TYPES = [
  { label: "Service", value: "service" },
  { label: "Database", value: "database" },
  { label: "Cache", value: "cache" },
  { label: "Message Queue", value: "message_queue" },
  { label: "Load Balancer", value: "load_balancer" },
  { label: "Gateway", value: "gateway" },
  { label: "External", value: "external" },
] as const;

export const SECURITY_CATEGORIES = [
  { label: "Vulnerability", value: "vulnerability" },
  { label: "Misconfiguration", value: "misconfiguration" },
  { label: "Exposed Secret", value: "exposed_secret" },
  { label: "Weak Authentication", value: "weak_authentication" },
  { label: "Network Exposure", value: "network_exposure" },
  { label: "Compliance Violation", value: "compliance_violation" },
  { label: "Data Leakage", value: "data_leakage" },
] as const;

export const COST_CATEGORIES = [
  { label: "Compute", value: "compute" },
  { label: "Storage", value: "storage" },
  { label: "Network", value: "network" },
  { label: "Database", value: "database" },
  { label: "Cache", value: "cache" },
  { label: "Monitoring", value: "monitoring" },
  { label: "Licensing", value: "licensing" },
] as const;

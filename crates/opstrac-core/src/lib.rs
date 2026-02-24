use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

// ──────────────────────────────────────────────
// Incident
// ──────────────────────────────────────────────

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct Incident {
    pub id: Uuid,
    pub tenant_id: String,
    pub title: String,
    pub description: Option<String>,
    pub severity: String,
    pub status: String,
    pub source: Option<String>,
    pub affected_services: Option<Vec<String>>,
    pub root_cause: Option<String>,
    pub correlation_id: Option<Uuid>,
    pub acknowledged_by: Option<String>,
    pub resolved_by: Option<String>,
    pub created_at: DateTime<Utc>,
    pub acknowledged_at: Option<DateTime<Utc>>,
    pub resolved_at: Option<DateTime<Utc>>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateIncident {
    pub tenant_id: String,
    pub title: String,
    pub description: Option<String>,
    pub severity: Option<String>,
    pub source: Option<String>,
    pub affected_services: Option<Vec<String>>,
}

// ──────────────────────────────────────────────
// Anomaly
// ──────────────────────────────────────────────

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct Anomaly {
    pub id: Uuid,
    pub tenant_id: String,
    pub metric_name: String,
    pub service: String,
    pub module: Option<String>,
    pub anomaly_type: String,
    pub severity: String,
    pub expected_value: Option<f64>,
    pub actual_value: Option<f64>,
    pub deviation_percent: Option<f64>,
    pub detected_at: DateTime<Utc>,
    pub resolved_at: Option<DateTime<Utc>>,
    pub status: String,
    pub metadata: Option<serde_json::Value>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateAnomaly {
    pub tenant_id: String,
    pub metric_name: String,
    pub service: String,
    pub module: Option<String>,
    pub anomaly_type: Option<String>,
    pub severity: Option<String>,
    pub expected_value: Option<f64>,
    pub actual_value: Option<f64>,
    pub deviation_percent: Option<f64>,
    pub metadata: Option<serde_json::Value>,
}

// ──────────────────────────────────────────────
// Rule
// ──────────────────────────────────────────────

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct Rule {
    pub id: Uuid,
    pub tenant_id: String,
    pub name: String,
    pub description: Option<String>,
    #[sqlx(rename = "type")]
    #[serde(rename = "type")]
    pub rule_type: String,
    pub condition: serde_json::Value,
    pub action: serde_json::Value,
    pub enabled: Option<bool>,
    pub priority: Option<i32>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateRule {
    pub tenant_id: String,
    pub name: String,
    pub description: Option<String>,
    #[serde(rename = "type")]
    pub rule_type: String,
    pub condition: serde_json::Value,
    pub action: serde_json::Value,
    pub enabled: Option<bool>,
    pub priority: Option<i32>,
}

// ──────────────────────────────────────────────
// Topology Node
// ──────────────────────────────────────────────

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct TopologyNode {
    pub id: Uuid,
    pub tenant_id: String,
    pub name: String,
    #[sqlx(rename = "type")]
    #[serde(rename = "type")]
    pub node_type: String,
    pub module: Option<String>,
    pub status: Option<String>,
    pub metadata: Option<serde_json::Value>,
    pub dependencies: Option<Vec<Uuid>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateTopologyNode {
    pub tenant_id: String,
    pub name: String,
    #[serde(rename = "type")]
    pub node_type: String,
    pub module: Option<String>,
    pub status: Option<String>,
    pub metadata: Option<serde_json::Value>,
    pub dependencies: Option<Vec<Uuid>>,
}

// ──────────────────────────────────────────────
// Remediation Action
// ──────────────────────────────────────────────

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct RemediationAction {
    pub id: Uuid,
    pub tenant_id: String,
    pub incident_id: Option<Uuid>,
    pub action_type: String,
    pub target_service: String,
    pub parameters: Option<serde_json::Value>,
    pub status: String,
    pub result: Option<serde_json::Value>,
    pub initiated_by: Option<String>,
    pub initiated_at: DateTime<Utc>,
    pub completed_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateRemediationAction {
    pub tenant_id: String,
    pub incident_id: Option<Uuid>,
    pub action_type: String,
    pub target_service: String,
    pub parameters: Option<serde_json::Value>,
    pub initiated_by: Option<String>,
}

// ──────────────────────────────────────────────
// Cost Report
// ──────────────────────────────────────────────

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct CostReport {
    pub id: Uuid,
    pub tenant_id: String,
    pub period_start: DateTime<Utc>,
    pub period_end: DateTime<Utc>,
    pub total_cost: Option<f64>,
    pub breakdown: Option<serde_json::Value>,
    pub recommendations: Option<serde_json::Value>,
    pub created_at: DateTime<Utc>,
}

// ──────────────────────────────────────────────
// Security Finding
// ──────────────────────────────────────────────

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct SecurityFinding {
    pub id: Uuid,
    pub tenant_id: String,
    pub title: String,
    pub description: Option<String>,
    pub severity: String,
    pub category: String,
    pub affected_resource: Option<String>,
    pub status: String,
    pub remediation: Option<String>,
    pub detected_at: DateTime<Utc>,
    pub resolved_at: Option<DateTime<Utc>>,
}

// ──────────────────────────────────────────────
// Errors
// ──────────────────────────────────────────────

#[derive(Debug, thiserror::Error)]
pub enum AIOpsError {
    #[error("Database error: {0}")]
    Database(#[from] sqlx::Error),

    #[error("Not found: {0}")]
    NotFound(String),

    #[error("Validation error: {0}")]
    Validation(String),

    #[error("Tenant ID required")]
    TenantRequired,

    #[error("Internal error: {0}")]
    Internal(String),
}

// ──────────────────────────────────────────────
// Severity and Status enums
// ──────────────────────────────────────────────

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Severity {
    Critical,
    High,
    Medium,
    Low,
    Info,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum IncidentStatus {
    Open,
    Acknowledged,
    Investigating,
    Resolved,
    Closed,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum AnomalyStatus {
    Active,
    Investigating,
    Resolved,
    Dismissed,
}

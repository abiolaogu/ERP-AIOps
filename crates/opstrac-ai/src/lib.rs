//! AI analysis types and client for the AIOps AI brain service.

use serde::{Deserialize, Serialize};
use uuid::Uuid;

/// Request to analyze an incident.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AnalyzeIncidentRequest {
    pub tenant_id: String,
    pub incident_id: Uuid,
    pub title: String,
    pub description: Option<String>,
    pub affected_services: Option<Vec<String>>,
    pub related_anomalies: Option<Vec<Uuid>>,
}

/// Response from AI incident analysis.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AnalyzeIncidentResponse {
    pub incident_id: Uuid,
    pub root_cause: Option<String>,
    pub confidence: f64,
    pub suggested_actions: Vec<SuggestedAction>,
    pub similar_incidents: Vec<Uuid>,
    pub impact_assessment: String,
}

/// A suggested remediation action from AI analysis.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SuggestedAction {
    pub action_type: String,
    pub target_service: String,
    pub description: String,
    pub confidence: f64,
    pub risk_level: String,
    pub parameters: serde_json::Value,
}

/// Request to forecast a metric.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ForecastRequest {
    pub tenant_id: String,
    pub metric_name: String,
    pub service: String,
    pub horizon_hours: u32,
    pub historical_data: Option<Vec<DataPoint>>,
}

/// A single data point for forecasting.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DataPoint {
    pub timestamp: i64,
    pub value: f64,
}

/// Response from metric forecasting.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ForecastResponse {
    pub metric_name: String,
    pub service: String,
    pub predictions: Vec<DataPoint>,
    pub confidence_upper: Vec<DataPoint>,
    pub confidence_lower: Vec<DataPoint>,
    pub anomaly_probability: f64,
}

/// AI Brain client configuration.
#[derive(Debug, Clone)]
pub struct AIBrainConfig {
    pub base_url: String,
    pub timeout_secs: u64,
}

impl Default for AIBrainConfig {
    fn default() -> Self {
        Self {
            base_url: "http://ai-brain:8001".to_string(),
            timeout_secs: 30,
        }
    }
}

//! Autopilot mode for AIOps.
//!
//! Provides fully autonomous operations management where the system
//! detects, diagnoses, and resolves issues without human intervention.

use serde::{Deserialize, Serialize};

/// Autopilot configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AutopilotConfig {
    pub enabled: bool,
    pub max_auto_remediation_severity: String,
    pub require_approval_above: String,
    pub learning_mode: bool,
    pub confidence_threshold: f64,
}

impl Default for AutopilotConfig {
    fn default() -> Self {
        Self {
            enabled: false,
            max_auto_remediation_severity: "medium".to_string(),
            require_approval_above: "high".to_string(),
            learning_mode: true,
            confidence_threshold: 0.85,
        }
    }
}

/// Autopilot decision.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AutopilotDecision {
    pub action: String,
    pub confidence: f64,
    pub reasoning: String,
    pub approved: bool,
}

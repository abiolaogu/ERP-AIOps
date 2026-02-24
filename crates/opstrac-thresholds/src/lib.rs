//! Adaptive thresholds for AIOps metrics.
//!
//! Dynamically adjusts alert thresholds based on historical patterns,
//! seasonality, and trend analysis to reduce false positives.

use serde::{Deserialize, Serialize};

/// An adaptive threshold configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AdaptiveThreshold {
    pub metric_name: String,
    pub service: String,
    pub baseline: f64,
    pub upper_bound: f64,
    pub lower_bound: f64,
    pub sensitivity: f64,
    pub learning_rate: f64,
}

/// Threshold evaluation result.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ThresholdEvaluation {
    pub metric_name: String,
    pub current_value: f64,
    pub threshold: AdaptiveThreshold,
    pub is_anomalous: bool,
    pub deviation: f64,
}

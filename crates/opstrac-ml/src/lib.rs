//! Machine learning pipeline types for AIOps.
//!
//! Provides model training, inference, and lifecycle management
//! for anomaly detection and forecasting models.

use serde::{Deserialize, Serialize};

/// ML model metadata.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModelMetadata {
    pub name: String,
    pub version: String,
    pub model_type: ModelType,
    pub accuracy: f64,
    pub trained_at: String,
    pub features: Vec<String>,
}

/// Supported model types.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ModelType {
    AnomalyDetection,
    Forecasting,
    Classification,
    Clustering,
    RootCauseAnalysis,
}

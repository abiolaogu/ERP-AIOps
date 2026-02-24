//! Cost optimization types for AIOps.
//!
//! Provides infrastructure cost tracking, analysis, and optimization
//! recommendations across the ERP platform.

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

/// A cost breakdown entry.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CostBreakdown {
    pub category: CostCategory,
    pub service: String,
    pub amount: f64,
    pub currency: String,
    pub period_start: DateTime<Utc>,
    pub period_end: DateTime<Utc>,
}

/// Cost category.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum CostCategory {
    Compute,
    Storage,
    Network,
    Database,
    Cache,
    Monitoring,
    Licensing,
    Other,
}

/// A cost optimization recommendation.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CostRecommendation {
    pub title: String,
    pub description: String,
    pub estimated_savings: f64,
    pub effort: EffortLevel,
    pub risk: RiskLevel,
    pub category: CostCategory,
    pub target_service: String,
}

/// Effort level for a recommendation.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum EffortLevel {
    Low,
    Medium,
    High,
}

/// Risk level.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum RiskLevel {
    Low,
    Medium,
    High,
}

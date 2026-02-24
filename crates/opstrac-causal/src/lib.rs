//! Causal analysis engine for AIOps.
//!
//! Implements causal inference to determine root cause relationships
//! between events, metrics, and incidents.

use serde::{Deserialize, Serialize};
use uuid::Uuid;

/// A causal relationship between two events.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CausalLink {
    pub cause: Uuid,
    pub effect: Uuid,
    pub confidence: f64,
    pub mechanism: String,
}

/// Result of causal analysis.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CausalAnalysis {
    pub root_causes: Vec<Uuid>,
    pub causal_chain: Vec<CausalLink>,
    pub confidence: f64,
}

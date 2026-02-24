//! Workflow orchestration for AIOps pipelines.
//!
//! Manages complex multi-step workflows for incident response,
//! remediation, and automated operations.

use serde::{Deserialize, Serialize};

/// A workflow definition.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Workflow {
    pub name: String,
    pub description: String,
    pub steps: Vec<WorkflowStep>,
    pub trigger: WorkflowTrigger,
}

/// A step in a workflow.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkflowStep {
    pub name: String,
    pub action: String,
    pub parameters: serde_json::Value,
    pub on_failure: FailureAction,
}

/// What triggers a workflow.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum WorkflowTrigger {
    Incident,
    Anomaly,
    Schedule,
    Manual,
    Webhook,
}

/// Action to take on step failure.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum FailureAction {
    Stop,
    Continue,
    Retry,
    Rollback,
}

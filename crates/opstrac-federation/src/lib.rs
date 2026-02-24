//! Federation support for AIOps cross-module integration.
//!
//! Enables AIOps to integrate with other ERP modules through
//! Hasura GraphQL federation, providing unified observability
//! and operations management.

use serde::{Deserialize, Serialize};

/// Federation endpoint configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FederationEndpoint {
    pub module: String,
    pub graphql_url: String,
    pub health_url: String,
    pub capabilities: Vec<String>,
}

/// Cross-module event.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FederatedEvent {
    pub source_module: String,
    pub event_type: String,
    pub payload: serde_json::Value,
}

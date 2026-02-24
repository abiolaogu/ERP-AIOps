//! Service topology mapping for AIOps.
//!
//! Discovers, maps, and maintains the service dependency graph
//! across the entire ERP platform for impact analysis and
//! root cause identification.

use serde::{Deserialize, Serialize};
use uuid::Uuid;

/// Type of topology node.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum NodeType {
    Service,
    Database,
    Cache,
    MessageQueue,
    LoadBalancer,
    Gateway,
    External,
}

/// A directed edge in the topology graph.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TopologyEdge {
    pub source: Uuid,
    pub target: Uuid,
    pub edge_type: EdgeType,
    pub latency_ms: Option<f64>,
    pub error_rate: Option<f64>,
}

/// Type of dependency edge.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum EdgeType {
    Http,
    Grpc,
    Database,
    Cache,
    Queue,
    Event,
}

/// Service topology graph.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TopologyGraph {
    pub tenant_id: String,
    pub nodes: Vec<Uuid>,
    pub edges: Vec<TopologyEdge>,
}

impl TopologyGraph {
    /// Get all downstream dependencies of a node.
    pub fn downstream(&self, node_id: &Uuid) -> Vec<&TopologyEdge> {
        self.edges.iter().filter(|e| &e.source == node_id).collect()
    }

    /// Get all upstream dependencies of a node.
    pub fn upstream(&self, node_id: &Uuid) -> Vec<&TopologyEdge> {
        self.edges.iter().filter(|e| &e.target == node_id).collect()
    }
}

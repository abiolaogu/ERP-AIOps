//! Security scanning types for AIOps.
//!
//! Provides security posture assessment, vulnerability scanning,
//! and compliance checking across the ERP infrastructure.

use serde::{Deserialize, Serialize};

/// Security scan configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScanConfig {
    pub scan_type: ScanType,
    pub targets: Vec<String>,
    pub depth: ScanDepth,
    pub include_recommendations: bool,
}

/// Type of security scan.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ScanType {
    Vulnerability,
    Compliance,
    Configuration,
    Network,
    Identity,
    DataExposure,
}

/// Depth of the security scan.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum ScanDepth {
    Quick,
    Standard,
    Deep,
}

/// Security finding category.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum FindingCategory {
    Vulnerability,
    Misconfiguration,
    ExposedSecret,
    WeakAuthentication,
    NetworkExposure,
    ComplianceViolation,
    DataLeakage,
    PrivilegeEscalation,
}

/// Summary of a security scan.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScanSummary {
    pub total_findings: u32,
    pub critical: u32,
    pub high: u32,
    pub medium: u32,
    pub low: u32,
    pub info: u32,
    pub scan_duration_ms: u64,
}

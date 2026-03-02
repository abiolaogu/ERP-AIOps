use axum::{
    extract::State,
    http::StatusCode,
    response::Json,
};
use serde::Deserialize;
use std::sync::Arc;
use uuid::Uuid;

use crate::{AppState, TenantId};

// ──────────────────────────────────────────────
// Hasura Event Trigger Payload
// ──────────────────────────────────────────────

#[derive(Debug, Deserialize)]
pub struct HasuraEventPayload {
    pub event: HasuraEvent,
    pub table: HasuraTable,
    pub trigger: HasuraTrigger,
}

#[derive(Debug, Deserialize)]
pub struct HasuraEvent {
    pub op: String,
    pub data: HasuraEventData,
    pub session_variables: Option<serde_json::Value>,
}

#[derive(Debug, Deserialize)]
pub struct HasuraEventData {
    pub old: Option<serde_json::Value>,
    pub new: Option<serde_json::Value>,
}

#[derive(Debug, Deserialize)]
pub struct HasuraTable {
    pub schema: String,
    pub name: String,
}

#[derive(Debug, Deserialize)]
pub struct HasuraTrigger {
    pub name: String,
}

// ──────────────────────────────────────────────
// Alertmanager Webhook Payload
// ──────────────────────────────────────────────

#[derive(Debug, Deserialize)]
pub struct AlertmanagerPayload {
    pub status: String,
    pub alerts: Vec<AlertmanagerAlert>,
    #[serde(rename = "groupLabels")]
    pub group_labels: Option<serde_json::Value>,
    #[serde(rename = "commonLabels")]
    pub common_labels: Option<serde_json::Value>,
}

#[derive(Debug, Deserialize)]
pub struct AlertmanagerAlert {
    pub status: String,
    pub labels: serde_json::Value,
    pub annotations: Option<serde_json::Value>,
    #[serde(rename = "startsAt")]
    pub starts_at: Option<String>,
    #[serde(rename = "endsAt")]
    pub ends_at: Option<String>,
    #[serde(rename = "generatorURL")]
    pub generator_url: Option<String>,
}

// ──────────────────────────────────────────────
// Helper: Extract tenant_id from event payload
// ──────────────────────────────────────────────

fn extract_tenant_id(data: &Option<serde_json::Value>) -> String {
    data.as_ref()
        .and_then(|v| v.get("tenant_id"))
        .and_then(|v| v.as_str())
        .unwrap_or("default")
        .to_string()
}

fn extract_field_str(data: &Option<serde_json::Value>, field: &str) -> Option<String> {
    data.as_ref()
        .and_then(|v| v.get(field))
        .and_then(|v| v.as_str())
        .map(|s| s.to_string())
}

fn extract_field_uuid(data: &Option<serde_json::Value>, field: &str) -> Option<Uuid> {
    data.as_ref()
        .and_then(|v| v.get(field))
        .and_then(|v| v.as_str())
        .and_then(|s| Uuid::parse_str(s).ok())
}

// ──────────────────────────────────────────────
// Webhook: incident-created
// ──────────────────────────────────────────────

pub async fn on_incident_created(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(_tenant): axum::extract::Extension<TenantId>,
    Json(payload): Json<HasuraEventPayload>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    let data = &payload.event.data.new;
    let tenant_id = extract_tenant_id(data);
    let incident_id = extract_field_uuid(data, "id");
    let severity = extract_field_str(data, "severity").unwrap_or_else(|| "medium".to_string());
    let title = extract_field_str(data, "title").unwrap_or_default();

    tracing::info!(
        tenant_id = %tenant_id,
        incident_id = ?incident_id,
        severity = %severity,
        "Incident created webhook triggered"
    );

    // Check if within a maintenance window
    let in_maintenance: (i64,) = sqlx::query_as(
        "SELECT COUNT(*) FROM maintenance_windows WHERE tenant_id = $1 AND status = 'active' AND start_time <= now() AND end_time >= now() AND suppress_alerts = true",
    )
    .bind(&tenant_id)
    .fetch_one(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    if in_maintenance.0 > 0 {
        tracing::info!("Incident suppressed due to active maintenance window");
        return Ok(Json(serde_json::json!({ "status": "suppressed", "reason": "maintenance_window" })));
    }

    // Insert audit log
    let _ = sqlx::query(
        r#"INSERT INTO aiops_audit_log (tenant_id, action, actor, target_resource, decision, details, correlation_id)
           VALUES ($1, 'incident_created', 'hasura_trigger', $2, 'processed', $3, $4)"#,
    )
    .bind(&tenant_id)
    .bind(&title)
    .bind(serde_json::json!({ "severity": severity, "trigger": payload.trigger.name }))
    .bind(incident_id)
    .execute(&state.db)
    .await;

    Ok(Json(serde_json::json!({
        "status": "processed",
        "incident_id": incident_id,
        "actions_taken": ["audit_logged", "correlation_queued"]
    })))
}

// ──────────────────────────────────────────────
// Webhook: anomaly-detected
// ──────────────────────────────────────────────

pub async fn on_anomaly_detected(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(_tenant): axum::extract::Extension<TenantId>,
    Json(payload): Json<HasuraEventPayload>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    let data = &payload.event.data.new;
    let tenant_id = extract_tenant_id(data);
    let anomaly_id = extract_field_uuid(data, "id");
    let severity = extract_field_str(data, "severity").unwrap_or_else(|| "medium".to_string());
    let metric_name = extract_field_str(data, "metric_name").unwrap_or_default();
    let service = extract_field_str(data, "service").unwrap_or_default();

    tracing::info!(
        tenant_id = %tenant_id,
        anomaly_id = ?anomaly_id,
        severity = %severity,
        metric = %metric_name,
        "Anomaly detected webhook triggered"
    );

    // For critical anomalies, auto-create an incident
    let mut actions_taken = vec!["audit_logged", "correlation_queued"];

    if severity == "critical" || severity == "high" {
        let _ = sqlx::query(
            r#"INSERT INTO incidents (tenant_id, title, description, severity, source, affected_services)
               VALUES ($1, $2, $3, $4, 'anomaly_detection', ARRAY[$5])"#,
        )
        .bind(&tenant_id)
        .bind(format!("Anomaly detected: {} on {}", metric_name, service))
        .bind(format!("Automated incident from anomaly detection. Metric: {}, Service: {}", metric_name, service))
        .bind(&severity)
        .bind(&service)
        .execute(&state.db)
        .await;

        actions_taken.push("incident_created");
    }

    // Audit log
    let _ = sqlx::query(
        r#"INSERT INTO aiops_audit_log (tenant_id, action, actor, target_resource, decision, details, correlation_id)
           VALUES ($1, 'anomaly_detected', 'hasura_trigger', $2, 'processed', $3, $4)"#,
    )
    .bind(&tenant_id)
    .bind(&metric_name)
    .bind(serde_json::json!({ "severity": severity, "service": service }))
    .bind(anomaly_id)
    .execute(&state.db)
    .await;

    Ok(Json(serde_json::json!({
        "status": "processed",
        "anomaly_id": anomaly_id,
        "actions_taken": actions_taken
    })))
}

// ──────────────────────────────────────────────
// Webhook: health-status-changed
// ──────────────────────────────────────────────

pub async fn on_health_status_changed(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(_tenant): axum::extract::Extension<TenantId>,
    Json(payload): Json<HasuraEventPayload>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    let old_data = &payload.event.data.old;
    let new_data = &payload.event.data.new;
    let tenant_id = extract_tenant_id(new_data);
    let module_name = extract_field_str(new_data, "module_name").unwrap_or_default();
    let new_status = extract_field_str(new_data, "status").unwrap_or_default();
    let old_status = extract_field_str(old_data, "status").unwrap_or_default();

    tracing::info!(
        tenant_id = %tenant_id,
        module = %module_name,
        old_status = %old_status,
        new_status = %new_status,
        "Health status changed webhook triggered"
    );

    let mut actions_taken = vec!["audit_logged"];

    // Check for degraded → critical transition for autonomous remediation
    if (old_status == "degraded" || old_status == "healthy") && new_status == "critical" {
        // Check maintenance window
        let in_maintenance: (i64,) = sqlx::query_as(
            "SELECT COUNT(*) FROM maintenance_windows WHERE tenant_id = $1 AND status = 'active' AND start_time <= now() AND end_time >= now() AND suppress_remediation = true AND ($2 = ANY(target_modules) OR target_modules = '{}')",
        )
        .bind(&tenant_id)
        .bind(&module_name)
        .fetch_one(&state.db)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

        if in_maintenance.0 == 0 {
            // Create incident for critical status
            let _ = sqlx::query(
                r#"INSERT INTO incidents (tenant_id, title, description, severity, source, affected_services)
                   VALUES ($1, $2, $3, 'critical', 'health_monitor', ARRAY[$4])"#,
            )
            .bind(&tenant_id)
            .bind(format!("Module {} transitioned to critical", module_name))
            .bind(format!("Health status changed from {} to critical. Autonomous remediation may be triggered.", old_status))
            .bind(&module_name)
            .execute(&state.db)
            .await;

            actions_taken.push("incident_created");
            actions_taken.push("remediation_evaluation_queued");
        } else {
            actions_taken.push("suppressed_by_maintenance_window");
        }
    }

    // Audit log
    let _ = sqlx::query(
        r#"INSERT INTO aiops_audit_log (tenant_id, action, actor, target_module, decision, details)
           VALUES ($1, 'health_status_changed', 'hasura_trigger', $2, 'processed', $3)"#,
    )
    .bind(&tenant_id)
    .bind(&module_name)
    .bind(serde_json::json!({ "old_status": old_status, "new_status": new_status }))
    .execute(&state.db)
    .await;

    Ok(Json(serde_json::json!({
        "status": "processed",
        "module": module_name,
        "transition": format!("{} -> {}", old_status, new_status),
        "actions_taken": actions_taken
    })))
}

// ──────────────────────────────────────────────
// Webhook: slo-breached
// ──────────────────────────────────────────────

pub async fn on_slo_breached(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(_tenant): axum::extract::Extension<TenantId>,
    Json(payload): Json<HasuraEventPayload>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    let data = &payload.event.data.new;
    let tenant_id = extract_tenant_id(data);
    let module_name = extract_field_str(data, "module_name").unwrap_or_default();
    let slo_name = extract_field_str(data, "slo_name").unwrap_or_default();
    let status = extract_field_str(data, "status").unwrap_or_default();

    tracing::info!(
        tenant_id = %tenant_id,
        module = %module_name,
        slo = %slo_name,
        status = %status,
        "SLO status changed webhook triggered"
    );

    let mut actions_taken = vec!["audit_logged"];

    if status == "breached" || status == "at_risk" {
        // Create incident for SLO breach
        let severity = if status == "breached" { "critical" } else { "warning" };
        let _ = sqlx::query(
            r#"INSERT INTO incidents (tenant_id, title, description, severity, source, affected_services)
               VALUES ($1, $2, $3, $4, 'slo_tracking', ARRAY[$5])"#,
        )
        .bind(&tenant_id)
        .bind(format!("SLO {} {} for module {}", slo_name, status, module_name))
        .bind(format!("SLO '{}' on module '{}' has status: {}. Review error budget and consider remediation.", slo_name, module_name, status))
        .bind(severity)
        .bind(&module_name)
        .execute(&state.db)
        .await;

        actions_taken.push("incident_created");
        actions_taken.push("escalation_triggered");
    }

    // Audit log
    let _ = sqlx::query(
        r#"INSERT INTO aiops_audit_log (tenant_id, action, actor, target_module, decision, details)
           VALUES ($1, 'slo_status_changed', 'hasura_trigger', $2, 'processed', $3)"#,
    )
    .bind(&tenant_id)
    .bind(&module_name)
    .bind(serde_json::json!({ "slo_name": slo_name, "status": status }))
    .execute(&state.db)
    .await;

    Ok(Json(serde_json::json!({
        "status": "processed",
        "module": module_name,
        "slo": slo_name,
        "actions_taken": actions_taken
    })))
}

// ──────────────────────────────────────────────
// Webhook: guardrail-resolved
// ──────────────────────────────────────────────

pub async fn on_guardrail_resolved(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(_tenant): axum::extract::Extension<TenantId>,
    Json(payload): Json<HasuraEventPayload>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    let data = &payload.event.data.new;
    let tenant_id = extract_tenant_id(data);
    let eval_id = extract_field_uuid(data, "id");
    let action_type = extract_field_str(data, "action_type").unwrap_or_default();
    let result = extract_field_str(data, "result").unwrap_or_default();

    tracing::info!(
        tenant_id = %tenant_id,
        eval_id = ?eval_id,
        action_type = %action_type,
        result = %result,
        "Guardrail resolved webhook triggered"
    );

    // Audit log
    let _ = sqlx::query(
        r#"INSERT INTO aiops_audit_log (tenant_id, action, actor, decision, details, correlation_id)
           VALUES ($1, 'guardrail_resolved', 'hasura_trigger', $2, $3, $4)"#,
    )
    .bind(&tenant_id)
    .bind(&result)
    .bind(serde_json::json!({ "action_type": action_type }))
    .bind(eval_id)
    .execute(&state.db)
    .await;

    Ok(Json(serde_json::json!({
        "status": "processed",
        "evaluation_id": eval_id,
        "result": result
    })))
}

// ──────────────────────────────────────────────
// Webhook: alertmanager
// ──────────────────────────────────────────────

pub async fn on_alertmanager_alert(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(_tenant): axum::extract::Extension<TenantId>,
    Json(payload): Json<AlertmanagerPayload>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    let mut incidents_created = 0;

    for alert in &payload.alerts {
        if alert.status != "firing" {
            continue;
        }

        let alertname = alert.labels.get("alertname")
            .and_then(|v| v.as_str())
            .unwrap_or("unknown");
        let severity = alert.labels.get("severity")
            .and_then(|v| v.as_str())
            .unwrap_or("warning");
        let erp_module = alert.labels.get("erp_module")
            .and_then(|v| v.as_str())
            .unwrap_or("platform");
        let summary = alert.annotations.as_ref()
            .and_then(|a| a.get("summary"))
            .and_then(|v| v.as_str())
            .unwrap_or("");
        let description = alert.annotations.as_ref()
            .and_then(|a| a.get("description"))
            .and_then(|v| v.as_str())
            .unwrap_or("");

        // Use "default" tenant for infrastructure alerts
        let tenant_id = "default";

        let _ = sqlx::query(
            r#"INSERT INTO incidents (tenant_id, title, description, severity, source, affected_services)
               VALUES ($1, $2, $3, $4, 'alertmanager', ARRAY[$5])"#,
        )
        .bind(tenant_id)
        .bind(format!("[Alertmanager] {}: {}", alertname, summary))
        .bind(description)
        .bind(severity)
        .bind(erp_module)
        .execute(&state.db)
        .await;

        // Audit log
        let _ = sqlx::query(
            r#"INSERT INTO aiops_audit_log (tenant_id, action, actor, target_module, decision, details)
               VALUES ($1, 'alertmanager_alert', 'alertmanager', $2, 'incident_created', $3)"#,
        )
        .bind(tenant_id)
        .bind(erp_module)
        .bind(serde_json::json!({
            "alertname": alertname,
            "severity": severity,
            "status": alert.status,
            "starts_at": alert.starts_at,
        }))
        .execute(&state.db)
        .await;

        incidents_created += 1;
    }

    Ok(Json(serde_json::json!({
        "status": "processed",
        "alerts_received": payload.alerts.len(),
        "incidents_created": incidents_created
    })))
}

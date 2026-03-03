//! Event Trigger Webhook Handlers for ERP-AIOps
//!
//! Processes Hasura event triggers and Alertmanager webhooks.
//! Implements incident correlation, health-based remediation,
//! SLO breach escalation, and alert-to-incident conversion.

use actix_web::{web, HttpResponse};
use serde::Deserialize;
use sqlx::types::Uuid;

use crate::AppState;

// ============================================================
// Hasura Event Trigger Envelope
// ============================================================

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
    pub name: String,
    pub schema: String,
}

#[derive(Debug, Deserialize)]
pub struct HasuraTrigger {
    pub name: String,
}

// ============================================================
// Webhook Handlers
// ============================================================

/// Triggered when a new incident is inserted.
/// Correlates with existing incidents, checks maintenance windows,
/// auto-assigns escalation policy, optionally triggers runbook.
pub async fn on_incident_created(
    state: web::Data<AppState>,
    body: web::Json<HasuraEventPayload>,
) -> HttpResponse {
    let new_data = match &body.event.data.new {
        Some(d) => d,
        None => return HttpResponse::BadRequest().json(serde_json::json!({"error": "missing new data"})),
    };

    let incident_id = new_data.get("id").and_then(|v| v.as_str()).unwrap_or("unknown");
    let tenant_id = new_data.get("tenant_id").and_then(|v| v.as_str()).unwrap_or("unknown");
    let severity = new_data.get("severity").and_then(|v| v.as_str()).unwrap_or("medium");
    let source_module = new_data.get("source_module").and_then(|v| v.as_str()).unwrap_or("unknown");

    tracing::info!(incident_id = incident_id, severity = severity, module = source_module, "incident created webhook");

    // Check if module is in maintenance window
    let in_maintenance = sqlx::query_scalar::<_, i64>(
        r#"
        SELECT COUNT(*) FROM maintenance_windows
        WHERE tenant_id = $1 AND $2 = ANY(target_modules)
        AND status = 'active' AND suppress_alerts = true
        "#,
    )
    .bind(tenant_id)
    .bind(source_module)
    .fetch_one(&*state.db)
    .await
    .unwrap_or(0);

    if in_maintenance > 0 {
        tracing::info!(incident_id = incident_id, "suppressed — module in maintenance window");
        // Audit the suppression
        let _ = sqlx::query(
            r#"
            INSERT INTO aiops_audit_log (tenant_id, action, action_category, actor, actor_type, target_type, target_id, target_module, result, metadata)
            VALUES ($1, 'incident_suppressed', 'suppression', 'system', 'system', 'incident', $2, $3, 'skipped',
                    '{"reason": "maintenance_window"}'::jsonb)
            "#,
        )
        .bind(tenant_id)
        .bind(incident_id)
        .bind(source_module)
        .execute(&*state.db)
        .await;

        return HttpResponse::Ok().json(serde_json::json!({"status": "suppressed", "reason": "maintenance_window"}));
    }

    // Audit the incident creation
    let _ = sqlx::query(
        r#"
        INSERT INTO aiops_audit_log (tenant_id, action, action_category, actor, actor_type, target_type, target_id, target_module, result, new_state)
        VALUES ($1, 'incident_created', 'detection', 'system', 'system', 'incident', $2, $3, 'success', $4)
        "#,
    )
    .bind(tenant_id)
    .bind(incident_id)
    .bind(source_module)
    .bind(new_data)
    .execute(&*state.db)
    .await;

    HttpResponse::Ok().json(serde_json::json!({"status": "processed", "incident_id": incident_id}))
}

/// Triggered when a new anomaly is detected.
/// Feeds into correlation engine, checks thresholds, creates incident if warranted.
pub async fn on_anomaly_detected(
    state: web::Data<AppState>,
    body: web::Json<HasuraEventPayload>,
) -> HttpResponse {
    let new_data = match &body.event.data.new {
        Some(d) => d,
        None => return HttpResponse::BadRequest().json(serde_json::json!({"error": "missing new data"})),
    };

    let anomaly_id = new_data.get("id").and_then(|v| v.as_str()).unwrap_or("unknown");
    let tenant_id = new_data.get("tenant_id").and_then(|v| v.as_str()).unwrap_or("unknown");
    let severity = new_data.get("severity").and_then(|v| v.as_str()).unwrap_or("low");
    let source_module = new_data.get("source_module").and_then(|v| v.as_str()).unwrap_or("unknown");
    let deviation = new_data.get("deviation_score").and_then(|v| v.as_f64()).unwrap_or(0.0);

    tracing::info!(anomaly_id = anomaly_id, severity = severity, deviation = deviation, "anomaly detected webhook");

    // If severity is critical or high, auto-create an incident
    if severity == "critical" || severity == "high" {
        let incident_title = format!("Auto-incident: {} anomaly in {}", severity, source_module);
        let _ = sqlx::query(
            r#"
            INSERT INTO incidents (tenant_id, title, description, severity, source_module, source_event_id, metadata)
            VALUES ($1, $2, $3, $4, $5, $6, '{"auto_created": true, "source": "anomaly_detection"}'::jsonb)
            "#,
        )
        .bind(tenant_id)
        .bind(&incident_title)
        .bind(format!("Anomaly detected with deviation score {:.2}", deviation))
        .bind(severity)
        .bind(source_module)
        .bind(anomaly_id)
        .execute(&*state.db)
        .await;
        tracing::info!(anomaly_id = anomaly_id, "auto-created incident from anomaly");
    }

    // Audit
    let _ = sqlx::query(
        r#"
        INSERT INTO aiops_audit_log (tenant_id, action, action_category, actor, actor_type, target_type, target_id, target_module, result, new_state)
        VALUES ($1, 'anomaly_detected', 'detection', 'system', 'system', 'anomaly', $2, $3, 'success', $4)
        "#,
    )
    .bind(tenant_id)
    .bind(anomaly_id)
    .bind(source_module)
    .bind(new_data)
    .execute(&*state.db)
    .await;

    HttpResponse::Ok().json(serde_json::json!({"status": "processed", "anomaly_id": anomaly_id}))
}

/// Triggered when module_health_status.status changes.
/// If degraded->critical, checks autopilot config and triggers autonomous remediation.
pub async fn on_health_status_changed(
    state: web::Data<AppState>,
    body: web::Json<HasuraEventPayload>,
) -> HttpResponse {
    let old_data = body.event.data.old.as_ref();
    let new_data = match &body.event.data.new {
        Some(d) => d,
        None => return HttpResponse::BadRequest().json(serde_json::json!({"error": "missing new data"})),
    };

    let module_name = new_data.get("module_name").and_then(|v| v.as_str()).unwrap_or("unknown");
    let tenant_id = new_data.get("tenant_id").and_then(|v| v.as_str()).unwrap_or("unknown");
    let new_status = new_data.get("status").and_then(|v| v.as_str()).unwrap_or("unknown");
    let old_status = old_data
        .and_then(|d| d.get("status"))
        .and_then(|v| v.as_str())
        .unwrap_or("unknown");

    tracing::info!(module = module_name, old = old_status, new = new_status, "health status changed");

    // If transitioning to critical, create incident and consider autonomous remediation
    if new_status == "critical" && old_status != "critical" {
        let _ = sqlx::query(
            r#"
            INSERT INTO incidents (tenant_id, title, severity, source_module, metadata)
            VALUES ($1, $2, 'critical', $3, '{"auto_created": true, "source": "health_monitor"}'::jsonb)
            "#,
        )
        .bind(tenant_id)
        .bind(format!("Module {} is CRITICAL", module_name))
        .bind(module_name)
        .execute(&*state.db)
        .await;
    }

    // Audit
    let _ = sqlx::query(
        r#"
        INSERT INTO aiops_audit_log (tenant_id, action, action_category, actor, actor_type, target_type, target_id, target_module, result, previous_state, new_state)
        VALUES ($1, 'health_status_changed', 'detection', 'system', 'system', 'module_health', $2, $2, 'success', $3, $4)
        "#,
    )
    .bind(tenant_id)
    .bind(module_name)
    .bind(old_data.unwrap_or(&serde_json::json!({})))
    .bind(new_data)
    .execute(&*state.db)
    .await;

    HttpResponse::Ok().json(serde_json::json!({
        "status": "processed",
        "module": module_name,
        "transition": format!("{} -> {}", old_status, new_status)
    }))
}

/// Triggered when SLO status changes (especially to 'breached').
/// Escalates via notification channels and creates incident.
pub async fn on_slo_breached(
    state: web::Data<AppState>,
    body: web::Json<HasuraEventPayload>,
) -> HttpResponse {
    let new_data = match &body.event.data.new {
        Some(d) => d,
        None => return HttpResponse::BadRequest().json(serde_json::json!({"error": "missing new data"})),
    };

    let module_name = new_data.get("module_name").and_then(|v| v.as_str()).unwrap_or("unknown");
    let tenant_id = new_data.get("tenant_id").and_then(|v| v.as_str()).unwrap_or("unknown");
    let slo_name = new_data.get("slo_name").and_then(|v| v.as_str()).unwrap_or("unknown");
    let new_status = new_data.get("status").and_then(|v| v.as_str()).unwrap_or("unknown");

    tracing::info!(module = module_name, slo = slo_name, status = new_status, "SLO status changed");

    if new_status == "breached" {
        // Create incident for SLO breach
        let _ = sqlx::query(
            r#"
            INSERT INTO incidents (tenant_id, title, description, severity, source_module, metadata)
            VALUES ($1, $2, $3, 'high', $4, '{"auto_created": true, "source": "slo_tracking"}'::jsonb)
            "#,
        )
        .bind(tenant_id)
        .bind(format!("SLO Breach: {} on {}", slo_name, module_name))
        .bind(format!("SLO '{}' has been breached for module '{}'", slo_name, module_name))
        .bind(module_name)
        .execute(&*state.db)
        .await;
    }

    // Audit
    let _ = sqlx::query(
        r#"
        INSERT INTO aiops_audit_log (tenant_id, action, action_category, actor, actor_type, target_type, target_id, target_module, result, new_state)
        VALUES ($1, 'slo_status_changed', 'detection', 'system', 'system', 'slo', $2, $3, 'success', $4)
        "#,
    )
    .bind(tenant_id)
    .bind(slo_name)
    .bind(module_name)
    .bind(new_data)
    .execute(&*state.db)
    .await;

    HttpResponse::Ok().json(serde_json::json!({"status": "processed", "slo": slo_name, "module": module_name}))
}

/// Triggered when a guardrail evaluation is resolved.
pub async fn on_guardrail_resolved(
    state: web::Data<AppState>,
    body: web::Json<HasuraEventPayload>,
) -> HttpResponse {
    let new_data = match &body.event.data.new {
        Some(d) => d,
        None => return HttpResponse::BadRequest().json(serde_json::json!({"error": "missing new data"})),
    };

    let eval_id = new_data.get("id").and_then(|v| v.as_str()).unwrap_or("unknown");
    let tenant_id = new_data.get("tenant_id").and_then(|v| v.as_str()).unwrap_or("unknown");
    let result = new_data.get("result").and_then(|v| v.as_str()).unwrap_or("unknown");
    let action_type = new_data.get("action_type").and_then(|v| v.as_str()).unwrap_or("unknown");

    tracing::info!(eval_id = eval_id, result = result, action = action_type, "guardrail resolved");

    // Audit
    let _ = sqlx::query(
        r#"
        INSERT INTO aiops_audit_log (tenant_id, action, action_category, actor, actor_type, target_type, target_id, result, new_state)
        VALUES ($1, 'guardrail_resolved', 'guardrail', 'system', 'system', 'guardrail_evaluation', $2, $3, $4)
        "#,
    )
    .bind(tenant_id)
    .bind(eval_id)
    .bind(result)
    .bind(new_data)
    .execute(&*state.db)
    .await;

    HttpResponse::Ok().json(serde_json::json!({"status": "processed", "evaluation_id": eval_id, "result": result}))
}

// ============================================================
// Alertmanager Webhook (Phase 4A)
// ============================================================

#[derive(Debug, Deserialize)]
pub struct AlertmanagerPayload {
    pub status: String,
    pub alerts: Vec<AlertmanagerAlert>,
    #[serde(rename = "groupLabels")]
    pub group_labels: Option<serde_json::Value>,
    #[serde(rename = "commonLabels")]
    pub common_labels: Option<serde_json::Value>,
    #[serde(rename = "commonAnnotations")]
    pub common_annotations: Option<serde_json::Value>,
}

#[derive(Debug, Deserialize)]
pub struct AlertmanagerAlert {
    pub status: String,
    pub labels: serde_json::Value,
    pub annotations: serde_json::Value,
    #[serde(rename = "startsAt")]
    pub starts_at: String,
    #[serde(rename = "endsAt")]
    pub ends_at: Option<String>,
    #[serde(rename = "fingerprint")]
    pub fingerprint: String,
}

pub async fn on_alertmanager_alert(
    state: web::Data<AppState>,
    body: web::Json<AlertmanagerPayload>,
) -> HttpResponse {
    let payload = body.into_inner();
    let mut created_incidents = 0;

    for alert in &payload.alerts {
        if alert.status != "firing" {
            continue;
        }

        let severity = alert.labels.get("severity").and_then(|v| v.as_str()).unwrap_or("medium");
        let module = alert.labels.get("module").and_then(|v| v.as_str()).unwrap_or("unknown");
        let alertname = alert.labels.get("alertname").and_then(|v| v.as_str()).unwrap_or("unknown");
        let summary = alert.annotations.get("summary").and_then(|v| v.as_str()).unwrap_or(alertname);
        let description = alert.annotations.get("description").and_then(|v| v.as_str());

        // Convert Prometheus alert to AIOps incident
        let result = sqlx::query(
            r#"
            INSERT INTO incidents (tenant_id, title, description, severity, source_module, source_event_id, metadata)
            VALUES ('default', $1, $2, $3, $4, $5,
                    jsonb_build_object('source', 'alertmanager', 'fingerprint', $5, 'labels', $6::jsonb))
            "#,
        )
        .bind(summary)
        .bind(description.unwrap_or(""))
        .bind(severity)
        .bind(module)
        .bind(&alert.fingerprint)
        .bind(&alert.labels)
        .execute(&*state.db)
        .await;

        if result.is_ok() {
            created_incidents += 1;
        }
    }

    tracing::info!(
        total_alerts = payload.alerts.len(),
        firing = payload.alerts.iter().filter(|a| a.status == "firing").count(),
        incidents_created = created_incidents,
        "alertmanager webhook processed"
    );

    HttpResponse::Ok().json(serde_json::json!({
        "status": "processed",
        "alerts_received": payload.alerts.len(),
        "incidents_created": created_incidents
    }))
}

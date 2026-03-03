//! Hasura Action Handlers for ERP-AIOps
//!
//! Resolves GraphQL actions: module health checks, guardrail evaluations,
//! maintenance windows, runbook execution, and SLO status queries.

use actix_web::{web, HttpResponse};
use serde::{Deserialize, Serialize};
use sqlx::types::Uuid;

use crate::AppState;

// ============================================================
// Hasura Action Envelope
// ============================================================

#[derive(Debug, Deserialize)]
pub struct HasuraActionPayload<T> {
    pub action: HasuraAction,
    pub input: T,
    pub session_variables: HasuraSession,
}

#[derive(Debug, Deserialize)]
pub struct HasuraAction {
    pub name: String,
}

#[derive(Debug, Deserialize)]
pub struct HasuraSession {
    #[serde(rename = "x-hasura-role")]
    pub role: String,
    #[serde(rename = "x-hasura-tenant-id")]
    pub tenant_id: Option<String>,
}

// ============================================================
// Existing Actions
// ============================================================

pub async fn detect_anomalies(
    state: web::Data<AppState>,
    body: web::Json<serde_json::Value>,
) -> HttpResponse {
    tracing::info!("detect_anomalies action called");
    // TODO: Implement ML-based anomaly detection pipeline
    HttpResponse::Ok().json(serde_json::json!({
        "anomalies_detected": 0,
        "status": "analysis_complete"
    }))
}

pub async fn correlate_incidents(
    state: web::Data<AppState>,
    body: web::Json<serde_json::Value>,
) -> HttpResponse {
    tracing::info!("correlate_incidents action called");
    // TODO: Implement temporal + topological incident correlation
    HttpResponse::Ok().json(serde_json::json!({
        "correlations_found": 0,
        "status": "correlation_complete"
    }))
}

pub async fn execute_playbook(
    state: web::Data<AppState>,
    body: web::Json<serde_json::Value>,
) -> HttpResponse {
    tracing::info!("execute_playbook action called");
    // TODO: Implement playbook execution engine with step tracking
    HttpResponse::Accepted().json(serde_json::json!({
        "execution_id": Uuid::new_v4().to_string(),
        "status": "queued"
    }))
}

pub async fn topology_map(
    state: web::Data<AppState>,
    body: web::Json<serde_json::Value>,
) -> HttpResponse {
    let nodes = sqlx::query_as::<_, (Uuid, String, String, String, String)>(
        "SELECT id, module_name, service_name, node_type, health_status FROM topology_nodes ORDER BY module_name",
    )
    .fetch_all(&*state.db)
    .await;

    let edges = sqlx::query_as::<_, (Uuid, Uuid, Uuid, String)>(
        "SELECT id, source_node_id, target_node_id, edge_type FROM topology_edges",
    )
    .fetch_all(&*state.db)
    .await;

    match (nodes, edges) {
        (Ok(n), Ok(e)) => {
            let node_list: Vec<serde_json::Value> = n.iter().map(|(id, module, svc, ntype, health)| {
                serde_json::json!({"id": id.to_string(), "module": module, "service": svc, "type": ntype, "health": health})
            }).collect();
            let edge_list: Vec<serde_json::Value> = e.iter().map(|(id, src, tgt, etype)| {
                serde_json::json!({"id": id.to_string(), "source": src.to_string(), "target": tgt.to_string(), "type": etype})
            }).collect();
            HttpResponse::Ok().json(serde_json::json!({"nodes": node_list, "edges": edge_list}))
        }
        _ => HttpResponse::InternalServerError().json(serde_json::json!({"error": "topology query failed"})),
    }
}

// ============================================================
// New Cross-Module Actions
// ============================================================

#[derive(Debug, Deserialize)]
struct ModuleHealthCheckInput {
    module_name: String,
    #[serde(default)]
    probe_live: Option<bool>,
}

pub async fn module_health_check(
    state: web::Data<AppState>,
    body: web::Json<HasuraActionPayload<ModuleHealthCheckInput>>,
) -> HttpResponse {
    let input = &body.input;
    let tenant_id = body.session_variables.tenant_id.as_deref().unwrap_or("default");

    let row = sqlx::query_as::<_, (String, bool, bool, bool, Option<f64>, Option<f64>, Option<i32>, chrono::DateTime<chrono::Utc>)>(
        r#"
        SELECT status, gateway_healthy, hasura_healthy, database_healthy,
               latency_p95_ms, error_rate, pod_count, last_heartbeat_at
        FROM module_health_status
        WHERE tenant_id = $1 AND module_name = $2
        "#,
    )
    .bind(tenant_id)
    .bind(&input.module_name)
    .fetch_optional(&*state.db)
    .await;

    match row {
        Ok(Some((status, gw, hs, db, lat, err, pods, heartbeat))) => {
            // Count active incidents
            let incident_count: i64 = sqlx::query_scalar(
                "SELECT COUNT(*) FROM incidents WHERE tenant_id = $1 AND source_module = $2 AND status NOT IN ('resolved', 'closed')",
            )
            .bind(tenant_id)
            .bind(&input.module_name)
            .fetch_one(&*state.db)
            .await
            .unwrap_or(0);

            HttpResponse::Ok().json(serde_json::json!({
                "module_name": input.module_name,
                "status": status,
                "gateway_healthy": gw,
                "hasura_healthy": hs,
                "database_healthy": db,
                "latency_p95_ms": lat,
                "error_rate": err,
                "pod_count": pods,
                "last_heartbeat_at": heartbeat.to_rfc3339(),
                "active_incidents": incident_count
            }))
        }
        Ok(None) => HttpResponse::Ok().json(serde_json::json!({
            "module_name": input.module_name,
            "status": "unknown",
            "gateway_healthy": false,
            "hasura_healthy": false,
            "database_healthy": false,
            "active_incidents": 0
        })),
        Err(e) => {
            tracing::error!(error = %e, "module health check failed");
            HttpResponse::InternalServerError().json(serde_json::json!({"error": e.to_string()}))
        }
    }
}

#[derive(Debug, Deserialize)]
struct EvaluateGuardrailInput {
    action_type: String,
    target_module: String,
    #[serde(default)]
    parameters: Option<serde_json::Value>,
    requested_by: String,
}

pub async fn evaluate_guardrail(
    state: web::Data<AppState>,
    body: web::Json<HasuraActionPayload<EvaluateGuardrailInput>>,
) -> HttpResponse {
    let input = &body.input;
    let tenant_id = body.session_variables.tenant_id.as_deref().unwrap_or("default");

    // Determine tier and risk score based on action_type
    let (tier, risk, approvals, timeout) = match input.action_type.as_str() {
        "restart_pod" => ("autonomous", 1, 0, 0),
        "clear_cache" => ("autonomous", 1, 0, 0),
        "flush_queue" => ("autonomous", 2, 0, 0),
        "scale_up_by_one" => ("autonomous", 2, 0, 0),
        "toggle_circuit_breaker" => ("autonomous", 3, 0, 0),
        "create_incident" => ("autonomous", 1, 0, 0),
        "send_notification" => ("autonomous", 1, 0, 0),
        "scale_horizontally" => ("supervised", 4, 1, 30),
        "config_change" => ("supervised", 5, 1, 30),
        "db_pool_resize" => ("supervised", 5, 1, 30),
        "toggle_feature_flag" => ("supervised", 4, 1, 30),
        "rollback_deployment" => ("supervised", 6, 1, 30),
        "rate_limit_adjust" => ("supervised", 4, 1, 30),
        "failover_primary" => ("protected", 8, 2, 60),
        "data_migration" => ("protected", 9, 2, 60),
        "schema_change" => ("protected", 9, 2, 60),
        "destroy_resources" => ("protected", 10, 2, 60),
        "cross_module_remediation" => ("protected", 8, 2, 60),
        "db_failover" => ("protected", 9, 2, 60),
        "network_policy_change" => ("protected", 8, 2, 60),
        _ => ("protected", 10, 2, 60), // Unknown actions default to highest tier
    };

    // Autonomous actions are auto-approved
    let result = if tier == "autonomous" { "approved" } else { "pending_approval" };

    let evaluation_id = sqlx::query_scalar::<_, Uuid>(
        r#"
        INSERT INTO guardrail_evaluations
            (tenant_id, action_type, target_module, guardrail_tier, risk_score, result, requested_by, approvals_required, context)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id
        "#,
    )
    .bind(tenant_id)
    .bind(&input.action_type)
    .bind(&input.target_module)
    .bind(tier)
    .bind(risk)
    .bind(result)
    .bind(&input.requested_by)
    .bind(approvals)
    .bind(&input.parameters)
    .fetch_one(&*state.db)
    .await;

    match evaluation_id {
        Ok(id) => {
            tracing::info!(
                evaluation_id = %id,
                action = %input.action_type,
                tier = tier,
                result = result,
                "guardrail evaluated"
            );
            HttpResponse::Ok().json(serde_json::json!({
                "evaluation_id": id.to_string(),
                "guardrail_tier": tier,
                "result": result,
                "risk_score": risk,
                "approvals_required": approvals,
                "timeout_minutes": timeout,
                "message": format!("Action '{}' classified as {} tier", input.action_type, tier)
            }))
        }
        Err(e) => {
            tracing::error!(error = %e, "guardrail evaluation failed");
            HttpResponse::InternalServerError().json(serde_json::json!({"error": e.to_string()}))
        }
    }
}

#[derive(Debug, Deserialize)]
struct CreateMaintenanceWindowInput {
    name: String,
    #[serde(default)]
    description: Option<String>,
    target_modules: Vec<String>,
    starts_at: String,
    ends_at: String,
    #[serde(default = "default_true")]
    suppress_alerts: bool,
    #[serde(default = "default_true")]
    suppress_remediation: bool,
    #[serde(default)]
    suppress_notifications: bool,
    #[serde(default)]
    recurrence_rule: Option<String>,
}

fn default_true() -> bool { true }

pub async fn create_maintenance_window(
    state: web::Data<AppState>,
    body: web::Json<HasuraActionPayload<CreateMaintenanceWindowInput>>,
) -> HttpResponse {
    let input = &body.input;
    let tenant_id = body.session_variables.tenant_id.as_deref().unwrap_or("default");
    let role = &body.session_variables.role;

    let schedule_type = if input.recurrence_rule.is_some() { "recurring" } else { "one_time" };

    let result = sqlx::query_as::<_, (Uuid, String)>(
        r#"
        INSERT INTO maintenance_windows
            (tenant_id, name, description, target_modules, suppress_alerts, suppress_remediation,
             suppress_notifications, schedule_type, starts_at, ends_at, recurrence_rule, created_by)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::timestamptz, $10::timestamptz, $11, $12)
        RETURNING id, status
        "#,
    )
    .bind(tenant_id)
    .bind(&input.name)
    .bind(&input.description)
    .bind(&input.target_modules)
    .bind(input.suppress_alerts)
    .bind(input.suppress_remediation)
    .bind(input.suppress_notifications)
    .bind(schedule_type)
    .bind(&input.starts_at)
    .bind(&input.ends_at)
    .bind(&input.recurrence_rule)
    .bind(role)
    .fetch_one(&*state.db)
    .await;

    match result {
        Ok((id, status)) => {
            tracing::info!(window_id = %id, name = %input.name, "maintenance window created");
            HttpResponse::Created().json(serde_json::json!({
                "id": id.to_string(),
                "name": input.name,
                "status": status,
                "starts_at": input.starts_at,
                "ends_at": input.ends_at,
            }))
        }
        Err(e) => {
            tracing::error!(error = %e, "maintenance window creation failed");
            HttpResponse::InternalServerError().json(serde_json::json!({"error": e.to_string()}))
        }
    }
}

#[derive(Debug, Deserialize)]
struct ExecuteRunbookInput {
    runbook_id: String,
    #[serde(default)]
    incident_id: Option<String>,
    target_module: String,
    #[serde(default)]
    parameters: Option<serde_json::Value>,
}

pub async fn execute_runbook(
    state: web::Data<AppState>,
    body: web::Json<HasuraActionPayload<ExecuteRunbookInput>>,
) -> HttpResponse {
    let input = &body.input;
    let tenant_id = body.session_variables.tenant_id.as_deref().unwrap_or("default");

    // Look up runbook
    let runbook = sqlx::query_as::<_, (Uuid, String, i32, String)>(
        "SELECT id, name, risk_score, guardrail_tier FROM runbooks WHERE id = $1::uuid AND tenant_id = $2",
    )
    .bind(&input.runbook_id)
    .bind(tenant_id)
    .fetch_optional(&*state.db)
    .await;

    let (rb_id, rb_name, risk, tier) = match runbook {
        Ok(Some(r)) => r,
        Ok(None) => return HttpResponse::NotFound().json(serde_json::json!({"error": "runbook not found"})),
        Err(e) => return HttpResponse::InternalServerError().json(serde_json::json!({"error": e.to_string()})),
    };

    // Evaluate guardrail for the runbook execution
    let eval_result = if tier == "autonomous" { "approved" } else { "pending_approval" };
    let approvals = match tier.as_str() {
        "autonomous" => 0,
        "supervised" => 1,
        _ => 2,
    };

    let eval_id = sqlx::query_scalar::<_, Uuid>(
        r#"
        INSERT INTO guardrail_evaluations
            (tenant_id, action_type, target_module, guardrail_tier, risk_score, result, requested_by, approvals_required, context)
        VALUES ($1, $2, $3, $4, $5, $6, 'runbook-executor', $7, $8)
        RETURNING id
        "#,
    )
    .bind(tenant_id)
    .bind(format!("execute_runbook:{}", rb_name))
    .bind(&input.target_module)
    .bind(&tier)
    .bind(risk)
    .bind(eval_result)
    .bind(approvals)
    .bind(serde_json::json!({"runbook_id": input.runbook_id, "parameters": input.parameters}))
    .fetch_one(&*state.db)
    .await;

    match eval_id {
        Ok(eid) => {
            tracing::info!(runbook_id = %rb_id, evaluation_id = %eid, tier = %tier, "runbook execution requested");
            HttpResponse::Accepted().json(serde_json::json!({
                "execution_id": Uuid::new_v4().to_string(),
                "runbook_id": rb_id.to_string(),
                "status": eval_result,
                "guardrail_evaluation_id": eid.to_string(),
            }))
        }
        Err(e) => {
            tracing::error!(error = %e, "runbook guardrail evaluation failed");
            HttpResponse::InternalServerError().json(serde_json::json!({"error": e.to_string()}))
        }
    }
}

#[derive(Debug, Deserialize)]
struct GetSLOStatusInput {
    module_name: String,
    #[serde(default)]
    slo_name: Option<String>,
}

pub async fn slo_status(
    state: web::Data<AppState>,
    body: web::Json<HasuraActionPayload<GetSLOStatusInput>>,
) -> HttpResponse {
    let input = &body.input;
    let tenant_id = body.session_variables.tenant_id.as_deref().unwrap_or("default");

    let query = if let Some(ref name) = input.slo_name {
        sqlx::query_as::<_, (String, String, f64, Option<f64>, String, Option<f64>, Option<f64>, Option<chrono::DateTime<chrono::Utc>>)>(
            r#"
            SELECT slo_name, slo_type, target_value, current_value, status,
                   error_budget_remaining, error_budget_burn_rate, last_evaluated_at
            FROM slo_tracking
            WHERE tenant_id = $1 AND module_name = $2 AND slo_name = $3
            "#,
        )
        .bind(tenant_id)
        .bind(&input.module_name)
        .bind(name)
        .fetch_all(&*state.db)
        .await
    } else {
        sqlx::query_as::<_, (String, String, f64, Option<f64>, String, Option<f64>, Option<f64>, Option<chrono::DateTime<chrono::Utc>>)>(
            r#"
            SELECT slo_name, slo_type, target_value, current_value, status,
                   error_budget_remaining, error_budget_burn_rate, last_evaluated_at
            FROM slo_tracking
            WHERE tenant_id = $1 AND module_name = $2
            ORDER BY slo_name
            "#,
        )
        .bind(tenant_id)
        .bind(&input.module_name)
        .fetch_all(&*state.db)
        .await
    };

    match query {
        Ok(rows) => {
            let slos: Vec<serde_json::Value> = rows.iter().map(|(name, stype, target, current, status, budget, burn, evaluated)| {
                serde_json::json!({
                    "slo_name": name,
                    "slo_type": stype,
                    "target_value": target,
                    "current_value": current,
                    "status": status,
                    "error_budget_remaining_pct": budget,
                    "burn_rate": burn,
                    "last_evaluated_at": evaluated.map(|t| t.to_rfc3339()),
                })
            }).collect();
            HttpResponse::Ok().json(serde_json::json!({
                "module_name": input.module_name,
                "slos": slos,
            }))
        }
        Err(e) => {
            tracing::error!(error = %e, "SLO status query failed");
            HttpResponse::InternalServerError().json(serde_json::json!({"error": e.to_string()}))
        }
    }
}

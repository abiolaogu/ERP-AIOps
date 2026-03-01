use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::Json,
};
use serde::{Deserialize, Serialize};
use sqlx::PgPool;
use std::sync::Arc;
use uuid::Uuid;

use crate::{AppState, TenantId};

// ──────────────────────────────────────────────
// Request / Response Types
// ──────────────────────────────────────────────

#[derive(Debug, Deserialize)]
pub struct HasuraActionPayload<T> {
    pub input: T,
    pub session_variables: Option<serde_json::Value>,
}

#[derive(Debug, Deserialize)]
pub struct ModuleHealthCheckInput {
    pub module_name: Option<String>,
    pub probe_live: Option<bool>,
}

#[derive(Debug, Serialize)]
pub struct ModuleHealthCheckOutput {
    pub module_name: String,
    pub status: String,
    pub latency_ms: Option<f64>,
    pub error_rate: Option<f64>,
    pub pod_count: Option<i32>,
    pub last_heartbeat_at: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct EvaluateGuardrailInput {
    pub action_type: String,
    pub target_module: Option<String>,
    pub risk_score: Option<i32>,
    pub requested_by: String,
    pub reason: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct EvaluateGuardrailOutput {
    pub id: Uuid,
    pub tier: String,
    pub result: String,
    pub approval_chain: serde_json::Value,
    pub reason: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct CreateMaintenanceWindowInput {
    pub name: String,
    pub description: Option<String>,
    pub target_modules: Option<Vec<String>>,
    pub start_time: String,
    pub end_time: String,
    pub suppress_alerts: Option<bool>,
    pub suppress_remediation: Option<bool>,
}

#[derive(Debug, Serialize)]
pub struct CreateMaintenanceWindowOutput {
    pub id: Uuid,
    pub name: String,
    pub status: String,
    pub start_time: String,
    pub end_time: String,
}

#[derive(Debug, Deserialize)]
pub struct ExecuteRunbookInput {
    pub runbook_id: Uuid,
    pub target_module: Option<String>,
    pub parameters: Option<serde_json::Value>,
}

#[derive(Debug, Serialize)]
pub struct ExecuteRunbookOutput {
    pub execution_id: Uuid,
    pub runbook_id: Uuid,
    pub status: String,
    pub steps_completed: Option<i32>,
}

#[derive(Debug, Deserialize)]
pub struct GetSLOStatusInput {
    pub module_name: Option<String>,
    pub slo_name: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct SLOStatusOutput {
    pub module_name: String,
    pub slo_name: String,
    pub slo_type: String,
    pub target: f64,
    pub current_value: Option<f64>,
    pub error_budget_remaining: Option<f64>,
    pub status: String,
    pub burn_rate: Option<f64>,
}

#[derive(Debug, Deserialize)]
pub struct ModuleCommandInput {
    pub command_type: String,
    pub parameters: Option<serde_json::Value>,
}

// ──────────────────────────────────────────────
// Guardrail Tier Evaluation
// ──────────────────────────────────────────────

struct GuardrailTier {
    name: &'static str,
    max_risk: i32,
    requires_approval: bool,
    approval_count: i32,
}

const AUTONOMOUS_ACTIONS: &[&str] = &[
    "restart_pod", "clear_cache", "flush_queue", "scale_up_by_one",
    "toggle_circuit_breaker", "create_incident", "send_notification",
];

const SUPERVISED_ACTIONS: &[&str] = &[
    "scale_horizontally", "config_change", "db_pool_resize",
    "toggle_feature_flag", "rollback_deployment",
];

fn evaluate_tier(action_type: &str, risk_score: i32, is_cross_module: bool) -> GuardrailTier {
    // Cross-module operations always escalate to protected
    if is_cross_module {
        return GuardrailTier {
            name: "protected",
            max_risk: 10,
            requires_approval: true,
            approval_count: 2,
        };
    }

    if risk_score <= 3 && AUTONOMOUS_ACTIONS.contains(&action_type) {
        GuardrailTier {
            name: "autonomous",
            max_risk: 3,
            requires_approval: false,
            approval_count: 0,
        }
    } else if risk_score <= 7 && SUPERVISED_ACTIONS.contains(&action_type) {
        GuardrailTier {
            name: "supervised",
            max_risk: 7,
            requires_approval: true,
            approval_count: 1,
        }
    } else {
        GuardrailTier {
            name: "protected",
            max_risk: 10,
            requires_approval: true,
            approval_count: 2,
        }
    }
}

// ──────────────────────────────────────────────
// Handler: moduleHealthCheck
// ──────────────────────────────────────────────

pub async fn module_health_check(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Json(payload): Json<HasuraActionPayload<ModuleHealthCheckInput>>,
) -> Result<Json<Vec<ModuleHealthCheckOutput>>, StatusCode> {
    let input = payload.input;

    let rows: Vec<(String, String, Option<f64>, Option<f64>, Option<i32>, Option<chrono::DateTime<chrono::Utc>>)> = if let Some(ref module) = input.module_name {
        sqlx::query_as(
            "SELECT module_name, status, latency_ms, error_rate, pod_count, last_heartbeat_at FROM module_health_status WHERE tenant_id = $1 AND module_name = $2",
        )
        .bind(&tenant.0)
        .bind(module)
        .fetch_all(&state.db)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?
    } else {
        sqlx::query_as(
            "SELECT module_name, status, latency_ms, error_rate, pod_count, last_heartbeat_at FROM module_health_status WHERE tenant_id = $1 ORDER BY module_name",
        )
        .bind(&tenant.0)
        .fetch_all(&state.db)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?
    };

    let output: Vec<ModuleHealthCheckOutput> = rows
        .into_iter()
        .map(|(module_name, status, latency_ms, error_rate, pod_count, last_heartbeat)| {
            ModuleHealthCheckOutput {
                module_name,
                status,
                latency_ms,
                error_rate,
                pod_count,
                last_heartbeat_at: last_heartbeat.map(|t| t.to_rfc3339()),
            }
        })
        .collect();

    Ok(Json(output))
}

// ──────────────────────────────────────────────
// Handler: evaluateGuardrail
// ──────────────────────────────────────────────

pub async fn evaluate_guardrail(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Json(payload): Json<HasuraActionPayload<EvaluateGuardrailInput>>,
) -> Result<Json<EvaluateGuardrailOutput>, StatusCode> {
    let input = payload.input;
    let risk_score = input.risk_score.unwrap_or(5);
    let is_cross_module = input.target_module.is_some();

    let tier = evaluate_tier(&input.action_type, risk_score, is_cross_module);

    let result = if tier.requires_approval {
        "pending_approval"
    } else {
        "approved"
    };

    let approval_chain = if tier.requires_approval {
        serde_json::json!({
            "required_approvals": tier.approval_count,
            "current_approvals": 0,
            "timeout_minutes": if tier.name == "protected" { 60 } else { 30 }
        })
    } else {
        serde_json::json!([])
    };

    let row: (Uuid,) = sqlx::query_as(
        r#"INSERT INTO guardrail_evaluations (tenant_id, action_type, target_module, tier, risk_score, result, approval_chain, requested_by, reason)
           VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
           RETURNING id"#,
    )
    .bind(&tenant.0)
    .bind(&input.action_type)
    .bind(&input.target_module)
    .bind(tier.name)
    .bind(risk_score)
    .bind(result)
    .bind(&approval_chain)
    .bind(&input.requested_by)
    .bind(&input.reason)
    .fetch_one(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    // Insert audit log entry
    let _ = sqlx::query(
        r#"INSERT INTO aiops_audit_log (tenant_id, action, actor, target_module, decision, tier, risk_score, details)
           VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"#,
    )
    .bind(&tenant.0)
    .bind(format!("guardrail_evaluation:{}", input.action_type))
    .bind(&input.requested_by)
    .bind(&input.target_module)
    .bind(result)
    .bind(tier.name)
    .bind(risk_score)
    .bind(serde_json::json!({ "reason": input.reason }))
    .execute(&state.db)
    .await;

    Ok(Json(EvaluateGuardrailOutput {
        id: row.0,
        tier: tier.name.to_string(),
        result: result.to_string(),
        approval_chain,
        reason: input.reason,
    }))
}

// ──────────────────────────────────────────────
// Handler: createMaintenanceWindow
// ──────────────────────────────────────────────

pub async fn create_maintenance_window(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Json(payload): Json<HasuraActionPayload<CreateMaintenanceWindowInput>>,
) -> Result<(StatusCode, Json<CreateMaintenanceWindowOutput>), StatusCode> {
    let input = payload.input;

    let row: (Uuid, String) = sqlx::query_as(
        r#"INSERT INTO maintenance_windows (tenant_id, name, description, target_modules, start_time, end_time, suppress_alerts, suppress_remediation, created_by)
           VALUES ($1, $2, $3, $4, $5::timestamptz, $6::timestamptz, $7, $8, 'admin')
           RETURNING id, status"#,
    )
    .bind(&tenant.0)
    .bind(&input.name)
    .bind(&input.description)
    .bind(&input.target_modules.unwrap_or_default())
    .bind(&input.start_time)
    .bind(&input.end_time)
    .bind(input.suppress_alerts.unwrap_or(true))
    .bind(input.suppress_remediation.unwrap_or(true))
    .fetch_one(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok((StatusCode::CREATED, Json(CreateMaintenanceWindowOutput {
        id: row.0,
        name: input.name,
        status: row.1,
        start_time: input.start_time,
        end_time: input.end_time,
    })))
}

// ──────────────────────────────────────────────
// Handler: executeRunbook
// ──────────────────────────────────────────────

pub async fn execute_runbook(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Json(payload): Json<HasuraActionPayload<ExecuteRunbookInput>>,
) -> Result<(StatusCode, Json<ExecuteRunbookOutput>), StatusCode> {
    let input = payload.input;

    // Load runbook
    let runbook: Option<(String, serde_json::Value, String, bool)> = sqlx::query_as(
        "SELECT name, steps, risk_tier, auto_execute FROM runbooks WHERE id = $1 AND tenant_id = $2",
    )
    .bind(input.runbook_id)
    .bind(&tenant.0)
    .fetch_optional(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let (name, steps, risk_tier, _auto_execute) = runbook.ok_or(StatusCode::NOT_FOUND)?;

    let steps_array = steps.as_array().map(|a| a.len() as i32).unwrap_or(0);

    // Evaluate guardrail for runbook execution
    let is_cross_module = input.target_module.is_some();
    let tier = evaluate_tier("execute_runbook", 5, is_cross_module);

    let execution_id = Uuid::new_v4();

    // Insert audit log
    let _ = sqlx::query(
        r#"INSERT INTO aiops_audit_log (tenant_id, action, actor, target_module, target_resource, decision, tier, risk_score, details, correlation_id)
           VALUES ($1, 'execute_runbook', 'system', $2, $3, $4, $5, 5, $6, $7)"#,
    )
    .bind(&tenant.0)
    .bind(&input.target_module)
    .bind(&name)
    .bind(if tier.requires_approval { "pending_approval" } else { "approved" })
    .bind(tier.name)
    .bind(serde_json::json!({
        "runbook_id": input.runbook_id,
        "risk_tier": risk_tier,
        "parameters": input.parameters,
    }))
    .bind(execution_id)
    .execute(&state.db)
    .await;

    // Update execution count
    let _ = sqlx::query(
        "UPDATE runbooks SET execution_count = execution_count + 1, last_executed_at = now() WHERE id = $1",
    )
    .bind(input.runbook_id)
    .execute(&state.db)
    .await;

    Ok((StatusCode::ACCEPTED, Json(ExecuteRunbookOutput {
        execution_id,
        runbook_id: input.runbook_id,
        status: if tier.requires_approval {
            "pending_approval".to_string()
        } else {
            "executing".to_string()
        },
        steps_completed: Some(0),
    })))
}

// ──────────────────────────────────────────────
// Handler: getSLOStatus
// ──────────────────────────────────────────────

pub async fn slo_status(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Json(payload): Json<HasuraActionPayload<GetSLOStatusInput>>,
) -> Result<Json<Vec<SLOStatusOutput>>, StatusCode> {
    let input = payload.input;

    let rows: Vec<(String, String, String, f64, Option<f64>, Option<f64>, Option<f64>, String)> = match (&input.module_name, &input.slo_name) {
        (Some(module), Some(slo)) => {
            sqlx::query_as(
                "SELECT module_name, slo_name, slo_type, target, current_value, error_budget_total, error_budget_remaining, status FROM slo_tracking WHERE tenant_id = $1 AND module_name = $2 AND slo_name = $3",
            )
            .bind(&tenant.0)
            .bind(module)
            .bind(slo)
            .fetch_all(&state.db)
            .await
            .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?
        }
        (Some(module), None) => {
            sqlx::query_as(
                "SELECT module_name, slo_name, slo_type, target, current_value, error_budget_total, error_budget_remaining, status FROM slo_tracking WHERE tenant_id = $1 AND module_name = $2 ORDER BY slo_name",
            )
            .bind(&tenant.0)
            .bind(module)
            .fetch_all(&state.db)
            .await
            .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?
        }
        _ => {
            sqlx::query_as(
                "SELECT module_name, slo_name, slo_type, target, current_value, error_budget_total, error_budget_remaining, status FROM slo_tracking WHERE tenant_id = $1 ORDER BY module_name, slo_name",
            )
            .bind(&tenant.0)
            .fetch_all(&state.db)
            .await
            .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?
        }
    };

    let output: Vec<SLOStatusOutput> = rows
        .into_iter()
        .map(|(module_name, slo_name, slo_type, target, current_value, budget_total, budget_remaining, status)| {
            let burn_rate = match (budget_total, budget_remaining) {
                (Some(total), Some(remaining)) if total > 0.0 => {
                    Some((total - remaining) / total)
                }
                _ => None,
            };
            SLOStatusOutput {
                module_name,
                slo_name,
                slo_type,
                target,
                current_value,
                error_budget_remaining: budget_remaining,
                status,
                burn_rate,
            }
        })
        .collect();

    Ok(Json(output))
}

// ──────────────────────────────────────────────
// Handler: Module Commands
// ──────────────────────────────────────────────

pub async fn get_module_commands(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Path(module): Path<String>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    let rows: Vec<(Uuid, String, String, serde_json::Value, String)> = sqlx::query_as(
        r#"SELECT id, action, actor, details, created_at::text
           FROM aiops_audit_log
           WHERE tenant_id = $1 AND target_module = $2
           ORDER BY created_at DESC LIMIT 50"#,
    )
    .bind(&tenant.0)
    .bind(&module)
    .fetch_all(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let commands: Vec<serde_json::Value> = rows
        .into_iter()
        .map(|(id, action, actor, details, created_at)| {
            serde_json::json!({
                "id": id,
                "action": action,
                "actor": actor,
                "details": details,
                "created_at": created_at,
            })
        })
        .collect();

    Ok(Json(serde_json::json!({
        "module": module,
        "commands": commands
    })))
}

pub async fn send_module_command(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Path(module): Path<String>,
    Json(input): Json<ModuleCommandInput>,
) -> Result<(StatusCode, Json<serde_json::Value>), StatusCode> {
    // Evaluate guardrail for the command
    let tier = evaluate_tier(&input.command_type, 5, true);

    let result = if tier.requires_approval {
        "pending_approval"
    } else {
        "approved"
    };

    let row: (Uuid,) = sqlx::query_as(
        r#"INSERT INTO aiops_audit_log (tenant_id, action, actor, target_module, decision, tier, risk_score, details)
           VALUES ($1, $2, 'aiops', $3, $4, $5, 5, $6)
           RETURNING id"#,
    )
    .bind(&tenant.0)
    .bind(format!("command:{}", input.command_type))
    .bind(&module)
    .bind(result)
    .bind(tier.name)
    .bind(input.parameters.as_ref().unwrap_or(&serde_json::json!({})))
    .fetch_one(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok((StatusCode::CREATED, Json(serde_json::json!({
        "command_id": row.0,
        "module": module,
        "command_type": input.command_type,
        "tier": tier.name,
        "result": result,
    }))))
}

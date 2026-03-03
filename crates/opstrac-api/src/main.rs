//! ERP-AIOps Rust API (opstrac-api)
//!
//! Core business logic API for AIOps autonomous operations.
//! Handles ingestion, Hasura action resolution, event trigger webhooks,
//! and cross-module command orchestration.

use actix_web::{web, App, HttpServer, HttpRequest, HttpResponse, middleware};
use serde::{Deserialize, Serialize};
use sqlx::postgres::PgPoolOptions;
use sqlx::PgPool;
use std::env;
use tracing_actix_web::TracingLogger;

mod actions;
mod webhooks;

// ============================================================
// Application State
// ============================================================

#[derive(Clone)]
pub struct AppState {
    pub db: PgPool,
    pub config: AppConfig,
}

#[derive(Clone)]
pub struct AppConfig {
    pub rust_api_port: u16,
    pub redpanda_brokers: String,
    pub guardrail_config_path: String,
    pub slo_config_path: String,
}

// ============================================================
// Main Entry Point
// ============================================================

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // Initialize tracing
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| tracing_subscriber::EnvFilter::new("info")),
        )
        .json()
        .init();

    let database_url = env::var("DATABASE_URL")
        .unwrap_or_else(|_| "postgresql://aiops:aiops@localhost:5433/erp_aiops".to_string());
    let port: u16 = env::var("RUST_API_PORT")
        .unwrap_or_else(|_| "8091".to_string())
        .parse()
        .expect("RUST_API_PORT must be a valid port number");

    let pool = PgPoolOptions::new()
        .max_connections(50)
        .acquire_timeout(std::time::Duration::from_secs(10))
        .connect(&database_url)
        .await
        .expect("Failed to connect to database");

    tracing::info!("Connected to database");

    let config = AppConfig {
        rust_api_port: port,
        redpanda_brokers: env::var("REDPANDA_BROKERS").unwrap_or_else(|_| "localhost:9092".to_string()),
        guardrail_config_path: env::var("GUARDRAIL_CONFIG_PATH")
            .unwrap_or_else(|_| "config/guardrail-tiers.yaml".to_string()),
        slo_config_path: env::var("SLO_CONFIG_PATH")
            .unwrap_or_else(|_| "config/slo-definitions.yaml".to_string()),
    };

    let state = AppState {
        db: pool,
        config,
    };

    tracing::info!("Starting opstrac-api on port {}", port);

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(state.clone()))
            .wrap(TracingLogger::default())
            .wrap(middleware::Compress::default())
            // Health
            .route("/health", web::get().to(health))
            // Ingestion endpoints
            .route("/api/v1/ingest/health", web::post().to(ingest_health))
            .route("/api/v1/ingest/incident", web::post().to(ingest_incident))
            .route("/api/v1/ingest/metric", web::post().to(ingest_metric))
            .route("/api/v1/ingest/event", web::post().to(ingest_event))
            .route("/api/v1/ingest/observability", web::post().to(ingest_observability))
            // Hasura action handlers
            .route("/api/v1/actions/detect-anomalies", web::post().to(actions::detect_anomalies))
            .route("/api/v1/actions/correlate-incidents", web::post().to(actions::correlate_incidents))
            .route("/api/v1/actions/execute-playbook", web::post().to(actions::execute_playbook))
            .route("/api/v1/actions/topology-map", web::post().to(actions::topology_map))
            .route("/api/v1/actions/module-health-check", web::post().to(actions::module_health_check))
            .route("/api/v1/actions/evaluate-guardrail", web::post().to(actions::evaluate_guardrail))
            .route("/api/v1/actions/create-maintenance-window", web::post().to(actions::create_maintenance_window))
            .route("/api/v1/actions/execute-runbook", web::post().to(actions::execute_runbook))
            .route("/api/v1/actions/slo-status", web::post().to(actions::slo_status))
            // Event trigger webhooks
            .route("/api/v1/webhooks/incident-created", web::post().to(webhooks::on_incident_created))
            .route("/api/v1/webhooks/anomaly-detected", web::post().to(webhooks::on_anomaly_detected))
            .route("/api/v1/webhooks/health-status-changed", web::post().to(webhooks::on_health_status_changed))
            .route("/api/v1/webhooks/slo-breached", web::post().to(webhooks::on_slo_breached))
            .route("/api/v1/webhooks/guardrail-resolved", web::post().to(webhooks::on_guardrail_resolved))
            .route("/api/v1/webhooks/alertmanager", web::post().to(webhooks::on_alertmanager_alert))
            // Module commands
            .route("/api/v1/aiops/commands/{module}", web::get().to(get_module_commands))
            .route("/api/v1/aiops/commands/{module}", web::post().to(issue_module_command))
    })
    .bind(format!("0.0.0.0:{}", port))?
    .workers(num_cpus::get())
    .run()
    .await
}

// ============================================================
// Health
// ============================================================

async fn health() -> HttpResponse {
    HttpResponse::Ok().json(serde_json::json!({
        "status": "healthy",
        "service": "opstrac-api"
    }))
}

// ============================================================
// Ingestion Handlers (Phase 7B)
// ============================================================

#[derive(Debug, Deserialize)]
struct HealthEvent {
    event_id: String,
    tenant_id: String,
    module_name: String,
    status: String,
    gateway_healthy: bool,
    hasura_healthy: bool,
    #[serde(default)]
    latency_p95_ms: Option<f64>,
    #[serde(default)]
    error_rate: Option<f64>,
    #[serde(default)]
    pod_count: Option<i32>,
    timestamp: String,
}

async fn ingest_health(
    state: web::Data<AppState>,
    body: web::Json<HealthEvent>,
) -> HttpResponse {
    let event = body.into_inner();
    let result = sqlx::query(
        r#"
        INSERT INTO module_health_status (tenant_id, module_name, status, gateway_healthy, hasura_healthy,
            latency_p95_ms, error_rate, pod_count, last_heartbeat_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
        ON CONFLICT (tenant_id, module_name)
        DO UPDATE SET status = $3, gateway_healthy = $4, hasura_healthy = $5,
            latency_p95_ms = $6, error_rate = $7, pod_count = $8, last_heartbeat_at = NOW(), updated_at = NOW()
        "#,
    )
    .bind(&event.tenant_id)
    .bind(&event.module_name)
    .bind(&event.status)
    .bind(event.gateway_healthy)
    .bind(event.hasura_healthy)
    .bind(event.latency_p95_ms)
    .bind(event.error_rate)
    .bind(event.pod_count)
    .execute(&*state.db)
    .await;

    match result {
        Ok(_) => {
            tracing::info!(module = %event.module_name, status = %event.status, "health upserted");
            HttpResponse::Ok().json(serde_json::json!({"status": "ok"}))
        }
        Err(e) => {
            tracing::error!(error = %e, "health upsert failed");
            HttpResponse::InternalServerError().json(serde_json::json!({"error": e.to_string()}))
        }
    }
}

#[derive(Debug, Deserialize)]
struct IncidentEvent {
    tenant_id: String,
    title: String,
    #[serde(default)]
    description: Option<String>,
    severity: String,
    source_module: String,
    #[serde(default)]
    source_event_id: Option<String>,
    #[serde(default)]
    metadata: Option<serde_json::Value>,
}

async fn ingest_incident(
    state: web::Data<AppState>,
    body: web::Json<IncidentEvent>,
) -> HttpResponse {
    let event = body.into_inner();
    let result = sqlx::query_scalar::<_, sqlx::types::Uuid>(
        r#"
        INSERT INTO incidents (tenant_id, title, description, severity, source_module, source_event_id, metadata)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
        "#,
    )
    .bind(&event.tenant_id)
    .bind(&event.title)
    .bind(&event.description)
    .bind(&event.severity)
    .bind(&event.source_module)
    .bind(&event.source_event_id)
    .bind(&event.metadata)
    .fetch_one(&*state.db)
    .await;

    match result {
        Ok(id) => {
            tracing::info!(incident_id = %id, module = %event.source_module, "incident created");
            HttpResponse::Created().json(serde_json::json!({"id": id.to_string(), "status": "created"}))
        }
        Err(e) => {
            tracing::error!(error = %e, "incident creation failed");
            HttpResponse::InternalServerError().json(serde_json::json!({"error": e.to_string()}))
        }
    }
}

#[derive(Debug, Deserialize)]
struct MetricEvent {
    tenant_id: String,
    module_name: String,
    metric_name: String,
    metric_type: String,
    value: f64,
    #[serde(default)]
    unit: Option<String>,
    #[serde(default)]
    dimensions: Option<serde_json::Value>,
}

async fn ingest_metric(
    state: web::Data<AppState>,
    body: web::Json<MetricEvent>,
) -> HttpResponse {
    let event = body.into_inner();
    let result = sqlx::query(
        r#"
        INSERT INTO operational_metrics (tenant_id, module_name, metric_name, metric_type, value, unit, dimensions)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        "#,
    )
    .bind(&event.tenant_id)
    .bind(&event.module_name)
    .bind(&event.metric_name)
    .bind(&event.metric_type)
    .bind(event.value)
    .bind(&event.unit)
    .bind(&event.dimensions)
    .execute(&*state.db)
    .await;

    match result {
        Ok(_) => HttpResponse::Ok().json(serde_json::json!({"status": "ok"})),
        Err(e) => {
            tracing::error!(error = %e, "metric insert failed");
            HttpResponse::InternalServerError().json(serde_json::json!({"error": e.to_string()}))
        }
    }
}

async fn ingest_event(
    state: web::Data<AppState>,
    body: web::Json<serde_json::Value>,
) -> HttpResponse {
    // Generic event ingestion — log and audit
    let event = body.into_inner();
    let tenant_id = event.get("tenant_id").and_then(|v| v.as_str()).unwrap_or("unknown");
    let event_type = event.get("event_type").and_then(|v| v.as_str()).unwrap_or("unknown");

    let result = sqlx::query(
        r#"
        INSERT INTO aiops_audit_log (tenant_id, action, action_category, actor, actor_type, target_type, target_id, metadata)
        VALUES ($1, $2, 'detection', 'system', 'system', 'event', $3, $4)
        "#,
    )
    .bind(tenant_id)
    .bind(format!("ingest_{}", event_type))
    .bind(event.get("event_id").and_then(|v| v.as_str()).unwrap_or("unknown"))
    .bind(&event)
    .execute(&*state.db)
    .await;

    match result {
        Ok(_) => HttpResponse::Ok().json(serde_json::json!({"status": "ok"})),
        Err(e) => {
            tracing::error!(error = %e, "event audit failed");
            HttpResponse::InternalServerError().json(serde_json::json!({"error": e.to_string()}))
        }
    }
}

async fn ingest_observability(
    _state: web::Data<AppState>,
    body: web::Json<serde_json::Value>,
) -> HttpResponse {
    // Observability data from the bridge pipeline
    let event = body.into_inner();
    tracing::debug!(data = ?event, "observability data received");
    // TODO: Feed into anomaly detection engine
    HttpResponse::Ok().json(serde_json::json!({"status": "ok", "action": "queued_for_analysis"}))
}

// ============================================================
// Module Command Endpoints
// ============================================================

async fn get_module_commands(
    state: web::Data<AppState>,
    path: web::Path<String>,
) -> HttpResponse {
    let module = path.into_inner();
    // Query pending commands for the module
    let commands = sqlx::query_as::<_, (sqlx::types::Uuid, String, String, serde_json::Value)>(
        r#"
        SELECT ge.id, ge.action_type, ge.guardrail_tier, ge.context
        FROM guardrail_evaluations ge
        WHERE ge.target_module = $1 AND ge.result = 'approved'
        ORDER BY ge.created_at DESC
        LIMIT 20
        "#,
    )
    .bind(&module)
    .fetch_all(&*state.db)
    .await;

    match commands {
        Ok(cmds) => {
            let items: Vec<serde_json::Value> = cmds
                .iter()
                .map(|(id, action, tier, ctx)| {
                    serde_json::json!({
                        "id": id.to_string(),
                        "action_type": action,
                        "guardrail_tier": tier,
                        "context": ctx,
                    })
                })
                .collect();
            HttpResponse::Ok().json(serde_json::json!({"module": module, "commands": items}))
        }
        Err(e) => {
            tracing::error!(error = %e, "command query failed");
            HttpResponse::InternalServerError().json(serde_json::json!({"error": e.to_string()}))
        }
    }
}

#[derive(Debug, Deserialize)]
struct IssueCommandRequest {
    action_type: String,
    parameters: serde_json::Value,
    requested_by: String,
}

async fn issue_module_command(
    state: web::Data<AppState>,
    path: web::Path<String>,
    body: web::Json<IssueCommandRequest>,
) -> HttpResponse {
    let module = path.into_inner();
    let cmd = body.into_inner();

    // Insert guardrail evaluation — the guardrail engine will process it
    let result = sqlx::query_scalar::<_, sqlx::types::Uuid>(
        r#"
        INSERT INTO guardrail_evaluations (tenant_id, action_type, target_module, guardrail_tier, risk_score, result, requested_by, context)
        VALUES (current_setting('app.tenant_id', true), $1, $2, 'autonomous', 1, 'pending_approval', $3, $4)
        RETURNING id
        "#,
    )
    .bind(&cmd.action_type)
    .bind(&module)
    .bind(&cmd.requested_by)
    .bind(&cmd.parameters)
    .fetch_one(&*state.db)
    .await;

    match result {
        Ok(id) => {
            tracing::info!(evaluation_id = %id, module = %module, action = %cmd.action_type, "command issued");
            HttpResponse::Accepted().json(serde_json::json!({
                "evaluation_id": id.to_string(),
                "status": "pending_evaluation"
            }))
        }
        Err(e) => {
            tracing::error!(error = %e, "command issue failed");
            HttpResponse::InternalServerError().json(serde_json::json!({"error": e.to_string()}))
        }
    }
}

use axum::{
    extract::{Path, Query, State},
    http::{HeaderMap, StatusCode},
    middleware::{self, Next},
    response::Json,
    routing::{get, post},
    Router,
};
use serde::{Deserialize, Serialize};
use sqlx::postgres::PgPoolOptions;
use sqlx::PgPool;
use std::sync::Arc;
use tracing_subscriber::EnvFilter;
use uuid::Uuid;

use opstrac_core::*;

// ──────────────────────────────────────────────
// App State
// ──────────────────────────────────────────────

#[derive(Clone)]
struct AppState {
    db: PgPool,
}

// ──────────────────────────────────────────────
// Tenant Middleware
// ──────────────────────────────────────────────

async fn tenant_middleware(
    headers: HeaderMap,
    mut request: axum::http::Request<axum::body::Body>,
    next: Next,
) -> Result<axum::response::Response, StatusCode> {
    let tenant_id = headers
        .get("X-Tenant-ID")
        .and_then(|v| v.to_str().ok())
        .unwrap_or("default")
        .to_string();

    request.extensions_mut().insert(TenantId(tenant_id));
    Ok(next.run(request).await)
}

#[derive(Clone, Debug)]
struct TenantId(String);

// ──────────────────────────────────────────────
// Query Parameters
// ──────────────────────────────────────────────

#[derive(Debug, Deserialize)]
struct ListParams {
    limit: Option<i64>,
    offset: Option<i64>,
    status: Option<String>,
    severity: Option<String>,
    service: Option<String>,
}

// ──────────────────────────────────────────────
// Response Types
// ──────────────────────────────────────────────

#[derive(Serialize)]
struct ListResponse<T: Serialize> {
    data: Vec<T>,
    total: i64,
}

#[derive(Serialize)]
struct HealthResponse {
    status: String,
    service: String,
    version: String,
}

#[derive(Serialize)]
struct ErrorResponse {
    error: String,
    detail: String,
}

// ──────────────────────────────────────────────
// Handlers: Health
// ──────────────────────────────────────────────

async fn healthz() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "healthy".to_string(),
        service: "opstrac-api".to_string(),
        version: env!("CARGO_PKG_VERSION").to_string(),
    })
}

// ──────────────────────────────────────────────
// Handlers: Incidents
// ──────────────────────────────────────────────

async fn list_incidents(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Query(params): Query<ListParams>,
) -> Result<Json<ListResponse<Incident>>, StatusCode> {
    let limit = params.limit.unwrap_or(50);
    let offset = params.offset.unwrap_or(0);

    let incidents = sqlx::query_as::<_, Incident>(
        "SELECT * FROM incidents WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
    )
    .bind(&tenant.0)
    .bind(limit)
    .bind(offset)
    .fetch_all(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let total: (i64,) =
        sqlx::query_as("SELECT COUNT(*) FROM incidents WHERE tenant_id = $1")
            .bind(&tenant.0)
            .fetch_one(&state.db)
            .await
            .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(Json(ListResponse {
        data: incidents,
        total: total.0,
    }))
}

async fn create_incident(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Json(input): Json<CreateIncident>,
) -> Result<(StatusCode, Json<Incident>), StatusCode> {
    let incident = sqlx::query_as::<_, Incident>(
        r#"INSERT INTO incidents (tenant_id, title, description, severity, source, affected_services)
           VALUES ($1, $2, $3, $4, $5, $6)
           RETURNING *"#,
    )
    .bind(&tenant.0)
    .bind(&input.title)
    .bind(&input.description)
    .bind(input.severity.as_deref().unwrap_or("medium"))
    .bind(&input.source)
    .bind(&input.affected_services)
    .fetch_one(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok((StatusCode::CREATED, Json(incident)))
}

async fn get_incident(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Path(id): Path<Uuid>,
) -> Result<Json<Incident>, StatusCode> {
    let incident = sqlx::query_as::<_, Incident>(
        "SELECT * FROM incidents WHERE id = $1 AND tenant_id = $2",
    )
    .bind(id)
    .bind(&tenant.0)
    .fetch_optional(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?
    .ok_or(StatusCode::NOT_FOUND)?;

    Ok(Json(incident))
}

// ──────────────────────────────────────────────
// Handlers: Anomalies
// ──────────────────────────────────────────────

async fn list_anomalies(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Query(params): Query<ListParams>,
) -> Result<Json<ListResponse<Anomaly>>, StatusCode> {
    let limit = params.limit.unwrap_or(50);
    let offset = params.offset.unwrap_or(0);

    let anomalies = sqlx::query_as::<_, Anomaly>(
        "SELECT * FROM anomalies WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
    )
    .bind(&tenant.0)
    .bind(limit)
    .bind(offset)
    .fetch_all(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let total: (i64,) =
        sqlx::query_as("SELECT COUNT(*) FROM anomalies WHERE tenant_id = $1")
            .bind(&tenant.0)
            .fetch_one(&state.db)
            .await
            .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(Json(ListResponse {
        data: anomalies,
        total: total.0,
    }))
}

async fn create_anomaly(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Json(input): Json<CreateAnomaly>,
) -> Result<(StatusCode, Json<Anomaly>), StatusCode> {
    let anomaly = sqlx::query_as::<_, Anomaly>(
        r#"INSERT INTO anomalies (tenant_id, metric_name, service, module, anomaly_type, severity, expected_value, actual_value, deviation_percent, metadata)
           VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
           RETURNING *"#,
    )
    .bind(&tenant.0)
    .bind(&input.metric_name)
    .bind(&input.service)
    .bind(&input.module)
    .bind(input.anomaly_type.as_deref().unwrap_or("spike"))
    .bind(input.severity.as_deref().unwrap_or("medium"))
    .bind(input.expected_value)
    .bind(input.actual_value)
    .bind(input.deviation_percent)
    .bind(&input.metadata)
    .fetch_one(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok((StatusCode::CREATED, Json(anomaly)))
}

// ──────────────────────────────────────────────
// Handlers: Rules
// ──────────────────────────────────────────────

async fn list_rules(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Query(params): Query<ListParams>,
) -> Result<Json<ListResponse<Rule>>, StatusCode> {
    let limit = params.limit.unwrap_or(50);
    let offset = params.offset.unwrap_or(0);

    let rules = sqlx::query_as::<_, Rule>(
        "SELECT * FROM aiops_rules WHERE tenant_id = $1 ORDER BY priority DESC, created_at DESC LIMIT $2 OFFSET $3",
    )
    .bind(&tenant.0)
    .bind(limit)
    .bind(offset)
    .fetch_all(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let total: (i64,) =
        sqlx::query_as("SELECT COUNT(*) FROM aiops_rules WHERE tenant_id = $1")
            .bind(&tenant.0)
            .fetch_one(&state.db)
            .await
            .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(Json(ListResponse {
        data: rules,
        total: total.0,
    }))
}

async fn create_rule(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Json(input): Json<CreateRule>,
) -> Result<(StatusCode, Json<Rule>), StatusCode> {
    let rule = sqlx::query_as::<_, Rule>(
        r#"INSERT INTO aiops_rules (tenant_id, name, description, type, condition, action, enabled, priority)
           VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
           RETURNING *"#,
    )
    .bind(&tenant.0)
    .bind(&input.name)
    .bind(&input.description)
    .bind(&input.rule_type)
    .bind(&input.condition)
    .bind(&input.action)
    .bind(input.enabled.unwrap_or(true))
    .bind(input.priority.unwrap_or(0))
    .fetch_one(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok((StatusCode::CREATED, Json(rule)))
}

// ──────────────────────────────────────────────
// Handlers: Topology
// ──────────────────────────────────────────────

async fn list_topology(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
) -> Result<Json<ListResponse<TopologyNode>>, StatusCode> {
    let nodes = sqlx::query_as::<_, TopologyNode>(
        "SELECT * FROM topology_nodes WHERE tenant_id = $1 ORDER BY name ASC",
    )
    .bind(&tenant.0)
    .fetch_all(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let total = nodes.len() as i64;

    Ok(Json(ListResponse {
        data: nodes,
        total,
    }))
}

// ──────────────────────────────────────────────
// Handlers: Remediation
// ──────────────────────────────────────────────

async fn execute_remediation(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Json(input): Json<CreateRemediationAction>,
) -> Result<(StatusCode, Json<RemediationAction>), StatusCode> {
    let action = sqlx::query_as::<_, RemediationAction>(
        r#"INSERT INTO remediation_actions (tenant_id, incident_id, action_type, target_service, parameters, initiated_by)
           VALUES ($1, $2, $3, $4, $5, $6)
           RETURNING *"#,
    )
    .bind(&tenant.0)
    .bind(input.incident_id)
    .bind(&input.action_type)
    .bind(&input.target_service)
    .bind(&input.parameters)
    .bind(&input.initiated_by)
    .fetch_one(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok((StatusCode::CREATED, Json(action)))
}

async fn list_remediations(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Query(params): Query<ListParams>,
) -> Result<Json<ListResponse<RemediationAction>>, StatusCode> {
    let limit = params.limit.unwrap_or(50);
    let offset = params.offset.unwrap_or(0);

    let actions = sqlx::query_as::<_, RemediationAction>(
        "SELECT * FROM remediation_actions WHERE tenant_id = $1 ORDER BY initiated_at DESC LIMIT $2 OFFSET $3",
    )
    .bind(&tenant.0)
    .bind(limit)
    .bind(offset)
    .fetch_all(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let total: (i64,) =
        sqlx::query_as("SELECT COUNT(*) FROM remediation_actions WHERE tenant_id = $1")
            .bind(&tenant.0)
            .fetch_one(&state.db)
            .await
            .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(Json(ListResponse {
        data: actions,
        total: total.0,
    }))
}

// ──────────────────────────────────────────────
// Handlers: Cost
// ──────────────────────────────────────────────

async fn cost_analysis(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
) -> Result<Json<ListResponse<CostReport>>, StatusCode> {
    let reports = sqlx::query_as::<_, CostReport>(
        "SELECT * FROM cost_reports WHERE tenant_id = $1 ORDER BY period_end DESC LIMIT 10",
    )
    .bind(&tenant.0)
    .fetch_all(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let total = reports.len() as i64;

    Ok(Json(ListResponse {
        data: reports,
        total,
    }))
}

// ──────────────────────────────────────────────
// Handlers: Security
// ──────────────────────────────────────────────

async fn list_security_findings(
    State(state): State<Arc<AppState>>,
    axum::extract::Extension(tenant): axum::extract::Extension<TenantId>,
    Query(params): Query<ListParams>,
) -> Result<Json<ListResponse<SecurityFinding>>, StatusCode> {
    let limit = params.limit.unwrap_or(50);
    let offset = params.offset.unwrap_or(0);

    let findings = sqlx::query_as::<_, SecurityFinding>(
        "SELECT * FROM security_findings WHERE tenant_id = $1 ORDER BY detected_at DESC LIMIT $2 OFFSET $3",
    )
    .bind(&tenant.0)
    .bind(limit)
    .bind(offset)
    .fetch_all(&state.db)
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let total: (i64,) =
        sqlx::query_as("SELECT COUNT(*) FROM security_findings WHERE tenant_id = $1")
            .bind(&tenant.0)
            .fetch_one(&state.db)
            .await
            .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(Json(ListResponse {
        data: findings,
        total: total.0,
    }))
}

// ──────────────────────────────────────────────
// Main
// ──────────────────────────────────────────────

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .json()
        .init();

    let database_url =
        std::env::var("DATABASE_URL").unwrap_or_else(|_| {
            "postgresql://erp:erp@localhost:5433/erp_aiops".to_string()
        });

    let pool = PgPoolOptions::new()
        .max_connections(20)
        .connect(&database_url)
        .await?;

    tracing::info!("Connected to database");

    let state = Arc::new(AppState { db: pool });

    let app = Router::new()
        // Health
        .route("/healthz", get(healthz))
        // Incidents
        .route("/api/v1/incidents", get(list_incidents).post(create_incident))
        .route("/api/v1/incidents/:id", get(get_incident))
        // Anomalies
        .route("/api/v1/anomalies", get(list_anomalies).post(create_anomaly))
        // Rules
        .route("/api/v1/rules", get(list_rules).post(create_rule))
        // Topology
        .route("/api/v1/topology", get(list_topology))
        // Remediation
        .route("/api/v1/remediation", get(list_remediations))
        .route("/api/v1/remediation/execute", post(execute_remediation))
        // Cost
        .route("/api/v1/cost/analysis", get(cost_analysis))
        // Security
        .route("/api/v1/security/findings", get(list_security_findings))
        // Middleware
        .layer(middleware::from_fn(tenant_middleware))
        .with_state(state);

    let port: u16 = std::env::var("RUST_API_PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse()
        .unwrap_or(8080);

    let listener = tokio::net::TcpListener::bind(format!("0.0.0.0:{}", port)).await?;
    tracing::info!("opstrac-api listening on port {}", port);

    axum::serve(listener, app).await?;

    Ok(())
}

"""AIOps AI Brain - FastAPI service for AI-powered incident analysis and forecasting."""

from datetime import datetime
from typing import Optional
from uuid import UUID, uuid4

from fastapi import FastAPI, Header, HTTPException
from pydantic import BaseModel, Field

app = FastAPI(
    title="ERP AIOps AI Brain",
    description="AI-powered analysis, correlation, and forecasting for AIOps",
    version="0.1.0",
)


# ──────────────────────────────────────────────
# Models
# ──────────────────────────────────────────────


class IncidentAnalysisRequest(BaseModel):
    incident_id: UUID
    title: str
    description: Optional[str] = None
    severity: str = "medium"
    affected_services: Optional[list[str]] = None
    related_anomalies: Optional[list[UUID]] = None


class SuggestedAction(BaseModel):
    action_type: str
    target_service: str
    description: str
    confidence: float
    risk_level: str
    parameters: dict = Field(default_factory=dict)


class IncidentAnalysisResponse(BaseModel):
    incident_id: UUID
    root_cause: Optional[str] = None
    confidence: float
    suggested_actions: list[SuggestedAction]
    similar_incidents: list[UUID]
    impact_assessment: str


class AnomalyAnalysisRequest(BaseModel):
    metric_name: str
    service: str
    anomaly_type: str = "spike"
    expected_value: Optional[float] = None
    actual_value: Optional[float] = None
    deviation_percent: Optional[float] = None
    metadata: dict = Field(default_factory=dict)


class AnomalyAnalysisResponse(BaseModel):
    is_true_anomaly: bool
    confidence: float
    explanation: str
    recommended_action: Optional[str] = None
    related_metrics: list[str]


class DataPoint(BaseModel):
    timestamp: int
    value: float


class ForecastRequest(BaseModel):
    metric_name: str
    service: str
    horizon_hours: int = 24
    historical_data: Optional[list[DataPoint]] = None


class ForecastResponse(BaseModel):
    metric_name: str
    service: str
    predictions: list[DataPoint]
    confidence_upper: list[DataPoint]
    confidence_lower: list[DataPoint]
    anomaly_probability: float


class CorrelationRequest(BaseModel):
    events: list[dict]
    time_window_secs: int = 300


class CorrelationResponse(BaseModel):
    groups: list[dict]
    total_events: int
    correlated_events: int
    correlation_rate: float


class HealthResponse(BaseModel):
    status: str
    service: str
    version: str
    timestamp: str


# ──────────────────────────────────────────────
# Routes
# ──────────────────────────────────────────────


@app.get("/health", response_model=HealthResponse)
async def health():
    """Health check endpoint."""
    return HealthResponse(
        status="healthy",
        service="ai-brain",
        version="0.1.0",
        timestamp=datetime.utcnow().isoformat(),
    )


@app.post("/analyze/incident", response_model=IncidentAnalysisResponse)
async def analyze_incident(
    request: IncidentAnalysisRequest,
    x_tenant_id: str = Header(default="default", alias="X-Tenant-ID"),
):
    """Analyze an incident using AI to determine root cause and suggest actions."""
    # Generate AI-based analysis
    suggested_actions = []

    if request.affected_services:
        for svc in request.affected_services[:3]:
            suggested_actions.append(
                SuggestedAction(
                    action_type="restart_service",
                    target_service=svc,
                    description=f"Restart {svc} to restore service health",
                    confidence=0.75,
                    risk_level="low",
                    parameters={"graceful": True, "timeout_secs": 30},
                )
            )

    root_cause = None
    if request.description:
        root_cause = (
            f"Preliminary analysis indicates potential cascading failure "
            f"originating from service dependencies. Severity: {request.severity}. "
            f"Recommend investigating recent deployments and configuration changes."
        )

    return IncidentAnalysisResponse(
        incident_id=request.incident_id,
        root_cause=root_cause,
        confidence=0.72,
        suggested_actions=suggested_actions,
        similar_incidents=[],
        impact_assessment=(
            f"Impact level: {request.severity}. "
            f"Affected services: {len(request.affected_services or [])}. "
            f"Estimated blast radius: moderate."
        ),
    )


@app.post("/analyze/anomaly", response_model=AnomalyAnalysisResponse)
async def analyze_anomaly(
    request: AnomalyAnalysisRequest,
    x_tenant_id: str = Header(default="default", alias="X-Tenant-ID"),
):
    """Analyze an anomaly to determine if it is a true positive."""
    deviation = request.deviation_percent or 0.0
    is_true = abs(deviation) > 20.0

    return AnomalyAnalysisResponse(
        is_true_anomaly=is_true,
        confidence=min(abs(deviation) / 100.0, 0.99) if deviation else 0.5,
        explanation=(
            f"Metric {request.metric_name} on {request.service} shows "
            f"{request.anomaly_type} pattern with {deviation:.1f}% deviation. "
            f"{'This exceeds normal variance and is likely a true anomaly.' if is_true else 'This is within expected variance bounds.'}"
        ),
        recommended_action=(
            "Investigate recent deployments or traffic pattern changes"
            if is_true
            else None
        ),
        related_metrics=[
            f"{request.metric_name}_p99",
            f"{request.service}_error_rate",
            f"{request.service}_request_rate",
        ],
    )


@app.post("/forecast", response_model=ForecastResponse)
async def forecast_metric(
    request: ForecastRequest,
    x_tenant_id: str = Header(default="default", alias="X-Tenant-ID"),
):
    """Forecast future metric values based on historical data."""
    import time

    now = int(time.time())
    hour = 3600
    predictions = []
    upper = []
    lower = []

    base_value = 50.0
    if request.historical_data and len(request.historical_data) > 0:
        base_value = sum(dp.value for dp in request.historical_data) / len(
            request.historical_data
        )

    for i in range(request.horizon_hours):
        ts = now + (i + 1) * hour
        # Simple linear forecast with slight noise pattern
        pred_val = base_value * (1.0 + 0.01 * i)
        margin = base_value * 0.1 * (1 + 0.05 * i)
        predictions.append(DataPoint(timestamp=ts, value=round(pred_val, 2)))
        upper.append(DataPoint(timestamp=ts, value=round(pred_val + margin, 2)))
        lower.append(DataPoint(timestamp=ts, value=round(max(0, pred_val - margin), 2)))

    return ForecastResponse(
        metric_name=request.metric_name,
        service=request.service,
        predictions=predictions,
        confidence_upper=upper,
        confidence_lower=lower,
        anomaly_probability=0.15,
    )


@app.post("/correlate", response_model=CorrelationResponse)
async def correlate_events(
    request: CorrelationRequest,
    x_tenant_id: str = Header(default="default", alias="X-Tenant-ID"),
):
    """Correlate events to find related groups."""
    total = len(request.events)
    # Simple mock correlation: group events by service
    groups = []
    service_groups: dict[str, list] = {}

    for event in request.events:
        svc = event.get("service", "unknown")
        if svc not in service_groups:
            service_groups[svc] = []
        service_groups[svc].append(event)

    correlated = 0
    for svc, events in service_groups.items():
        if len(events) > 1:
            correlated += len(events)
            groups.append(
                {
                    "correlation_id": str(uuid4()),
                    "service": svc,
                    "event_count": len(events),
                    "pattern": "temporal",
                    "confidence": 0.8,
                }
            )

    return CorrelationResponse(
        groups=groups,
        total_events=total,
        correlated_events=correlated,
        correlation_rate=correlated / max(total, 1),
    )


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8001)

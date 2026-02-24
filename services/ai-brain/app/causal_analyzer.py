"""Causal Analyzer - Performs root cause analysis on incidents."""
from typing import List, Optional, Dict
from pydantic import BaseModel
from datetime import datetime

class CausalFactor(BaseModel):
    component: str
    probability: float
    evidence: List[str]
    related_metrics: List[str] = []

class RCAResult(BaseModel):
    incident_id: str
    root_causes: List[CausalFactor]
    timeline: List[Dict]
    confidence: float
    summary: str

class CausalAnalyzer:
    def __init__(self):
        self.correlation_threshold = 0.7

    async def analyze(self, incident_id: str, metrics: Dict[str, List[float]], events: List[Dict], topology: Optional[Dict] = None) -> RCAResult:
        factors = []
        timeline = []
        for name, values in metrics.items():
            if len(values) > 1:
                recent_change = abs(values[-1] - values[-2]) / (abs(values[-2]) if values[-2] != 0 else 1)
                if recent_change > 0.5:
                    factors.append(CausalFactor(
                        component=name.split("_")[0] if "_" in name else name,
                        probability=min(recent_change, 1.0),
                        evidence=[f"Metric {name} changed by {recent_change*100:.1f}%"],
                        related_metrics=[name],
                    ))
        for event in events:
            timeline.append({"timestamp": event.get("timestamp", str(datetime.utcnow())), "event": event.get("description", "Unknown event"), "component": event.get("component", "unknown")})
        factors.sort(key=lambda f: f.probability, reverse=True)
        summary = f"Identified {len(factors)} potential root causes." if factors else "No clear root cause identified."
        return RCAResult(
            incident_id=incident_id, root_causes=factors[:5], timeline=timeline,
            confidence=factors[0].probability if factors else 0.0, summary=summary,
        )

causal_analyzer = CausalAnalyzer()

"""Incident Summarizer - Generates human-readable incident summaries using LLM."""
from typing import List, Dict, Optional
from pydantic import BaseModel

class IncidentSummary(BaseModel):
    incident_id: str
    title: str
    summary: str
    impact: str
    timeline: str
    root_cause: str
    resolution: str
    lessons_learned: str = ""

class IncidentSummarizer:
    def __init__(self):
        self.templates = {
            "critical": "CRITICAL INCIDENT: {title}\n\nImpact: {impact}\n\nRoot Cause: {root_cause}\n\nTimeline:\n{timeline}\n\nResolution: {resolution}",
            "warning": "WARNING: {title}\n\nImpact: {impact}\n\nDetails: {summary}",
        }

    def summarize(self, incident_id: str, events: List[Dict], metrics: Dict[str, List[float]], severity: str = "warning") -> IncidentSummary:
        title = f"Incident {incident_id[:8]}"
        if events:
            title = events[0].get("description", title)
        affected = set()
        for e in events:
            if "component" in e:
                affected.add(e["component"])
        impact = f"Affected components: {', '.join(affected) if affected else 'Unknown'}"
        timeline_parts = []
        for e in sorted(events, key=lambda x: x.get("timestamp", "")):
            timeline_parts.append(f"- [{e.get('timestamp', 'N/A')}] {e.get('description', 'Unknown event')}")
        timeline = "\n".join(timeline_parts) if timeline_parts else "No timeline data available"
        anomalous = [name for name, vals in metrics.items() if len(vals) > 1 and abs(vals[-1] - vals[0]) / (abs(vals[0]) if vals[0] != 0 else 1) > 0.5]
        root_cause = f"Anomalous metrics detected: {', '.join(anomalous)}" if anomalous else "Root cause under investigation"
        return IncidentSummary(
            incident_id=incident_id, title=title,
            summary=f"Incident involving {len(events)} events and {len(metrics)} metrics",
            impact=impact, timeline=timeline, root_cause=root_cause,
            resolution="Pending investigation",
        )

incident_summarizer = IncidentSummarizer()

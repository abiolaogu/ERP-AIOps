"""Correlation Engine - Correlates events across services and infrastructure."""
from typing import List, Dict, Optional
from pydantic import BaseModel
from datetime import datetime, timedelta

class Event(BaseModel):
    id: str
    source: str
    event_type: str
    timestamp: datetime
    severity: str
    description: str
    labels: Dict[str, str] = {}

class CorrelationGroup(BaseModel):
    group_id: str
    events: List[Event]
    correlation_score: float
    root_event: Optional[Event] = None
    pattern: str = ""

class CorrelationEngine:
    def __init__(self):
        self.time_window = timedelta(minutes=5)
        self.min_correlation_score = 0.6

    def correlate(self, events: List[Event]) -> List[CorrelationGroup]:
        if not events:
            return []
        events_sorted = sorted(events, key=lambda e: e.timestamp)
        groups: List[CorrelationGroup] = []
        used = set()
        for i, event in enumerate(events_sorted):
            if event.id in used:
                continue
            group_events = [event]
            used.add(event.id)
            for j in range(i + 1, len(events_sorted)):
                other = events_sorted[j]
                if other.id in used:
                    continue
                if other.timestamp - event.timestamp > self.time_window:
                    break
                score = self._similarity(event, other)
                if score >= self.min_correlation_score:
                    group_events.append(other)
                    used.add(other.id)
            if len(group_events) > 1:
                root = min(group_events, key=lambda e: e.timestamp)
                groups.append(CorrelationGroup(
                    group_id=f"corr-{event.id[:8]}",
                    events=group_events,
                    correlation_score=0.8,
                    root_event=root,
                    pattern=f"Temporal correlation within {self.time_window}",
                ))
        return groups

    def _similarity(self, a: Event, b: Event) -> float:
        score = 0.0
        if a.source == b.source:
            score += 0.3
        if a.event_type == b.event_type:
            score += 0.3
        if a.severity == b.severity:
            score += 0.2
        common_labels = set(a.labels.keys()) & set(b.labels.keys())
        if common_labels:
            matching = sum(1 for k in common_labels if a.labels[k] == b.labels[k])
            score += 0.2 * (matching / len(common_labels))
        return score

correlation_engine = CorrelationEngine()

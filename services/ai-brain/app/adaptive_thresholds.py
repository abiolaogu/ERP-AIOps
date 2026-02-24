"""Adaptive Thresholds - Dynamically adjusts alert thresholds based on historical patterns."""
import numpy as np
from typing import List, Dict, Optional
from pydantic import BaseModel

class ThresholdConfig(BaseModel):
    metric_name: str
    warning: float
    critical: float
    method: str = "percentile"
    sensitivity: float = 0.95

class AdaptiveThresholds:
    def __init__(self):
        self.history: Dict[str, List[float]] = {}
        self.max_history = 10000

    def update(self, metric_name: str, values: List[float]):
        if metric_name not in self.history:
            self.history[metric_name] = []
        self.history[metric_name].extend(values)
        if len(self.history[metric_name]) > self.max_history:
            self.history[metric_name] = self.history[metric_name][-self.max_history:]

    def calculate(self, metric_name: str, sensitivity: float = 0.95) -> Optional[ThresholdConfig]:
        values = self.history.get(metric_name, [])
        if len(values) < 50:
            return None
        arr = np.array(values)
        warning_pct = sensitivity * 100
        critical_pct = min(99.9, (sensitivity + (1 - sensitivity) / 2) * 100)
        return ThresholdConfig(
            metric_name=metric_name,
            warning=float(np.percentile(arr, warning_pct)),
            critical=float(np.percentile(arr, critical_pct)),
            method="percentile",
            sensitivity=sensitivity,
        )

    def evaluate(self, metric_name: str, value: float) -> Optional[str]:
        config = self.calculate(metric_name)
        if not config:
            return None
        if value >= config.critical:
            return "critical"
        if value >= config.warning:
            return "warning"
        return "ok"

adaptive_thresholds = AdaptiveThresholds()

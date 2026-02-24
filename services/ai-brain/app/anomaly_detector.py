"""Anomaly Detection - Detects anomalies in metric streams using statistical and ML methods."""
import numpy as np
from typing import List, Optional
from pydantic import BaseModel
from datetime import datetime

class MetricPoint(BaseModel):
    timestamp: datetime
    value: float
    labels: dict = {}

class AnomalyResult(BaseModel):
    is_anomaly: bool
    score: float
    method: str
    threshold: float
    metric_name: str
    timestamp: datetime
    details: str = ""

class AnomalyDetector:
    def __init__(self):
        self.z_score_threshold = 3.0
        self.iqr_multiplier = 1.5
        self.window_size = 100

    def detect_zscore(self, values: List[float], metric_name: str = "") -> List[AnomalyResult]:
        if len(values) < 10:
            return []
        arr = np.array(values)
        mean, std = arr.mean(), arr.std()
        if std == 0:
            return []
        results = []
        for i, v in enumerate(values):
            z = abs((v - mean) / std)
            if z > self.z_score_threshold:
                results.append(AnomalyResult(
                    is_anomaly=True, score=z, method="z-score",
                    threshold=self.z_score_threshold, metric_name=metric_name,
                    timestamp=datetime.utcnow(),
                    details=f"Value {v:.2f} is {z:.1f} std devs from mean {mean:.2f}"
                ))
        return results

    def detect_iqr(self, values: List[float], metric_name: str = "") -> List[AnomalyResult]:
        if len(values) < 10:
            return []
        arr = np.array(values)
        q1, q3 = np.percentile(arr, 25), np.percentile(arr, 75)
        iqr = q3 - q1
        lower, upper = q1 - self.iqr_multiplier * iqr, q3 + self.iqr_multiplier * iqr
        results = []
        for v in values:
            if v < lower or v > upper:
                score = abs(v - (q1 + q3) / 2) / (iqr if iqr > 0 else 1)
                results.append(AnomalyResult(
                    is_anomaly=True, score=score, method="iqr",
                    threshold=self.iqr_multiplier, metric_name=metric_name,
                    timestamp=datetime.utcnow(),
                    details=f"Value {v:.2f} outside IQR bounds [{lower:.2f}, {upper:.2f}]"
                ))
        return results

    def detect_moving_average(self, values: List[float], metric_name: str = "", window: int = 20) -> List[AnomalyResult]:
        if len(values) < window + 5:
            return []
        arr = np.array(values)
        results = []
        for i in range(window, len(arr)):
            window_mean = arr[i-window:i].mean()
            window_std = arr[i-window:i].std()
            if window_std > 0:
                z = abs((arr[i] - window_mean) / window_std)
                if z > self.z_score_threshold:
                    results.append(AnomalyResult(
                        is_anomaly=True, score=z, method="moving-average",
                        threshold=self.z_score_threshold, metric_name=metric_name,
                        timestamp=datetime.utcnow(),
                        details=f"Value {arr[i]:.2f} deviates from moving avg {window_mean:.2f}"
                    ))
        return results

anomaly_detector = AnomalyDetector()

"""Forecaster - Time series forecasting for capacity planning."""
import numpy as np
from typing import List, Tuple
from pydantic import BaseModel
from datetime import datetime, timedelta

class ForecastPoint(BaseModel):
    timestamp: datetime
    predicted_value: float
    lower_bound: float
    upper_bound: float

class ForecastResult(BaseModel):
    metric_name: str
    forecast: List[ForecastPoint]
    trend: str
    confidence: float
    breach_predicted: bool = False
    breach_timestamp: str = ""

class Forecaster:
    def __init__(self):
        self.default_horizon = 24

    def linear_forecast(self, values: List[float], horizon_hours: int = 24, metric_name: str = "", threshold: float = None) -> ForecastResult:
        if len(values) < 5:
            return ForecastResult(metric_name=metric_name, forecast=[], trend="insufficient_data", confidence=0.0)
        x = np.arange(len(values))
        coeffs = np.polyfit(x, values, 1)
        slope, intercept = coeffs
        std = np.std(values)
        now = datetime.utcnow()
        forecasts = []
        breach = False
        breach_ts = ""
        for h in range(1, horizon_hours + 1):
            idx = len(values) + h
            pred = slope * idx + intercept
            lower = pred - 1.96 * std
            upper = pred + 1.96 * std
            forecasts.append(ForecastPoint(timestamp=now + timedelta(hours=h), predicted_value=pred, lower_bound=lower, upper_bound=upper))
            if threshold and pred > threshold and not breach:
                breach = True
                breach_ts = str(now + timedelta(hours=h))
        trend = "increasing" if slope > 0.01 else "decreasing" if slope < -0.01 else "stable"
        r_squared = 1 - np.sum((np.array(values) - (slope * x + intercept))**2) / np.sum((np.array(values) - np.mean(values))**2)
        return ForecastResult(
            metric_name=metric_name, forecast=forecasts, trend=trend,
            confidence=max(0, min(1, r_squared)), breach_predicted=breach, breach_timestamp=breach_ts,
        )

forecaster = Forecaster()

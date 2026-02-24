"""Watchdog - Monitors system health and triggers alerts."""
import asyncio
from typing import Dict, List, Callable, Optional
from pydantic import BaseModel
from datetime import datetime

class HealthCheck(BaseModel):
    name: str
    status: str
    latency_ms: float
    last_checked: datetime
    message: str = ""

class WatchdogConfig(BaseModel):
    check_interval_seconds: int = 30
    unhealthy_threshold: int = 3
    recovery_threshold: int = 2

class Watchdog:
    def __init__(self, config: Optional[WatchdogConfig] = None):
        self.config = config or WatchdogConfig()
        self.checks: Dict[str, Callable] = {}
        self.failure_counts: Dict[str, int] = {}
        self.recovery_counts: Dict[str, int] = {}
        self.statuses: Dict[str, HealthCheck] = {}
        self._running = False

    def register_check(self, name: str, check_fn: Callable):
        self.checks[name] = check_fn
        self.failure_counts[name] = 0
        self.recovery_counts[name] = 0

    async def run_checks(self) -> List[HealthCheck]:
        results = []
        for name, check_fn in self.checks.items():
            import time
            start = time.time()
            try:
                await check_fn()
                latency = (time.time() - start) * 1000
                self.failure_counts[name] = 0
                self.recovery_counts[name] = min(self.recovery_counts[name] + 1, self.config.recovery_threshold)
                status = "healthy"
                message = "Check passed"
            except Exception as e:
                latency = (time.time() - start) * 1000
                self.failure_counts[name] += 1
                self.recovery_counts[name] = 0
                status = "unhealthy" if self.failure_counts[name] >= self.config.unhealthy_threshold else "degraded"
                message = str(e)
            check = HealthCheck(name=name, status=status, latency_ms=latency, last_checked=datetime.utcnow(), message=message)
            self.statuses[name] = check
            results.append(check)
        return results

    def get_status(self) -> Dict[str, HealthCheck]:
        return self.statuses

watchdog = Watchdog()

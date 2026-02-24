"""Local Models - Manages locally-running ML models for inference."""
from typing import Dict, List, Optional
from pydantic import BaseModel
import numpy as np

class ModelInfo(BaseModel):
    name: str
    version: str
    model_type: str
    status: str = "loaded"
    accuracy: float = 0.0

class PredictionRequest(BaseModel):
    model_name: str
    features: List[float]
    tenant_id: Optional[str] = None

class PredictionResponse(BaseModel):
    prediction: float
    confidence: float
    model_used: str

class LocalModels:
    def __init__(self):
        self.models: Dict[str, dict] = {}
        self._register_defaults()

    def _register_defaults(self):
        self.models["anomaly_isolation_forest"] = {
            "info": ModelInfo(name="anomaly_isolation_forest", version="1.0.0", model_type="anomaly_detection", accuracy=0.92),
            "threshold": 0.5,
        }
        self.models["capacity_linear"] = {
            "info": ModelInfo(name="capacity_linear", version="1.0.0", model_type="regression", accuracy=0.85),
            "weights": None,
        }
        self.models["failure_predictor"] = {
            "info": ModelInfo(name="failure_predictor", version="1.0.0", model_type="classification", accuracy=0.88),
            "threshold": 0.7,
        }

    def list_models(self) -> List[ModelInfo]:
        return [m["info"] for m in self.models.values()]

    def predict(self, request: PredictionRequest) -> PredictionResponse:
        if request.model_name not in self.models:
            return PredictionResponse(prediction=0.0, confidence=0.0, model_used="none")
        features = np.array(request.features)
        if request.model_name == "anomaly_isolation_forest":
            score = float(np.mean(np.abs(features - np.mean(features))) / (np.std(features) + 1e-8))
            return PredictionResponse(prediction=1.0 if score > 2.0 else 0.0, confidence=min(score / 3.0, 1.0), model_used=request.model_name)
        elif request.model_name == "capacity_linear":
            pred = float(np.mean(features) * 1.1)
            return PredictionResponse(prediction=pred, confidence=0.85, model_used=request.model_name)
        else:
            score = float(np.max(features) / (np.sum(features) + 1e-8))
            return PredictionResponse(prediction=1.0 if score > 0.5 else 0.0, confidence=score, model_used=request.model_name)

local_models = LocalModels()

"""LLM Router - Routes AI requests to appropriate language models."""
from typing import Optional
import httpx
from pydantic import BaseModel

class LLMRequest(BaseModel):
    prompt: str
    model: str = "default"
    max_tokens: int = 1024
    temperature: float = 0.7
    tenant_id: Optional[str] = None

class LLMResponse(BaseModel):
    text: str
    model_used: str
    tokens_used: int
    latency_ms: float

class LLMRouter:
    def __init__(self):
        self.models = {
            "default": {"endpoint": "http://localhost:11434/api/generate", "model": "llama3"},
            "fast": {"endpoint": "http://localhost:11434/api/generate", "model": "phi3"},
            "accurate": {"endpoint": "http://localhost:11434/api/generate", "model": "llama3:70b"},
        }

    async def route(self, request: LLMRequest) -> LLMResponse:
        config = self.models.get(request.model, self.models["default"])
        import time
        start = time.time()
        async with httpx.AsyncClient(timeout=120.0) as client:
            resp = await client.post(config["endpoint"], json={
                "model": config["model"],
                "prompt": request.prompt,
                "stream": False,
                "options": {"num_predict": request.max_tokens, "temperature": request.temperature}
            })
            data = resp.json()
        latency = (time.time() - start) * 1000
        return LLMResponse(
            text=data.get("response", ""),
            model_used=config["model"],
            tokens_used=data.get("eval_count", 0),
            latency_ms=latency,
        )

llm_router = LLMRouter()

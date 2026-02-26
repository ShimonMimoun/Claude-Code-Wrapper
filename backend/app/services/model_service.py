"""
Mock AI model catalog service.

Returns a fixed list of models available through the enterprise proxy.
"""

from app.schemas.models import ModelInfo, ModelListResponse

# ── Mock data ────────────────────────────────────────────────────────────────

_MOCK_MODELS: list[ModelInfo] = [
    ModelInfo(id="claude-sonnet-4-20250514", created=1717200000, owned_by="anthropic"),
    ModelInfo(id="claude-3-5-sonnet-20241022", created=1713830400, owned_by="anthropic"),
    ModelInfo(id="claude-3-haiku-20240307", created=1709769600, owned_by="anthropic"),
    ModelInfo(id="claude-3-opus-20240229", created=1709164800, owned_by="anthropic"),
    ModelInfo(id="claude-3-5-haiku-20241022", created=1713830400, owned_by="anthropic"),
]


def list_models() -> ModelListResponse:
    """Return the full list of available AI models."""
    return ModelListResponse(data=_MOCK_MODELS)

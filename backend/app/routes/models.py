"""
AI model listing route (OpenAI-compatible).

GET /v1/models  →  { "object": "list", "data": [...] }

Requires a valid bearer token.
"""

from fastapi import APIRouter, Depends

from app.auth.bearer import require_auth
from app.schemas.models import ModelListResponse
from app.services.model_service import list_models

router = APIRouter(prefix="/v1", tags=["Models"])


@router.get("/models", response_model=ModelListResponse)
async def get_models(_user: dict = Depends(require_auth)):
    """Return the list of AI models available through the enterprise proxy."""
    return list_models()

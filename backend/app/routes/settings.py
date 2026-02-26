"""
Enterprise Claude settings route.

GET /api/claude-settings  →  { "env": {...}, "permissions": {...}, ... }

Requires a valid bearer token.
"""

from fastapi import APIRouter, Depends

from app.auth.bearer import require_auth
from app.services.settings_service import get_enterprise_settings

router = APIRouter(prefix="/api", tags=["Settings"])


@router.get("/claude-settings")
async def get_settings(_user: dict = Depends(require_auth)):
    """Return the enterprise-managed Claude Code settings."""
    return get_enterprise_settings()

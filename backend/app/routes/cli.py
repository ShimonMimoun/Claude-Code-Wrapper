"""
CLI binary download route.

GET /cli/{platform}  →  binary download

Supported platforms: win, mac-intel, mac-m-series, linux.
Requires a valid bearer token.
"""

from fastapi import APIRouter, Depends

from app.auth.bearer import require_auth
from app.services.cli_service import get_cli_binary

router = APIRouter(prefix="/cli", tags=["CLI Download"])


@router.get("/{platform}")
async def download_cli(platform: str, _user: dict = Depends(require_auth)):
    """Download the Claude CLI binary for the specified platform."""
    return get_cli_binary(platform)

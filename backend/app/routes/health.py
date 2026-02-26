"""
Health check route.

GET /health  →  { "status": "ok" }
"""

from fastapi import APIRouter

router = APIRouter(tags=["Health"])


@router.get("/health")
async def health_check():
    """Simple health check — no authentication required."""
    return {"status": "ok"}

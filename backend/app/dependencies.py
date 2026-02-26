"""
Shared FastAPI dependencies (settings injection).
"""

from fastapi import Depends
from app.config import Settings, get_settings


def settings_dependency() -> Settings:
    """Inject the application Settings into route handlers."""
    return get_settings()

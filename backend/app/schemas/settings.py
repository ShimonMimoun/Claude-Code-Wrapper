"""Pydantic schemas for enterprise Claude settings."""

from pydantic import BaseModel


class EnvSettings(BaseModel):
    ANTHROPIC_BASE_URL: str = ""
    ANTHROPIC_MODEL: str = ""


class ClaudeSettings(BaseModel):
    """Top-level settings object sent to the CLI via GET /api/claude-settings."""
    env: EnvSettings = EnvSettings()
    permissions: dict = {}
    allowedTools: list[str] = []

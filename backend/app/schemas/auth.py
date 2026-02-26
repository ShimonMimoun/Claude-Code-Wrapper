"""Pydantic schemas for authentication / SSO responses."""

from pydantic import BaseModel


class TokenResponse(BaseModel):
    access_token: str
    token_type: str = "bearer"
    expires_in: int


class SSOLoginParams(BaseModel):
    redirect: str  # The CLI callback URL (e.g. http://127.0.0.1:8080/callback)

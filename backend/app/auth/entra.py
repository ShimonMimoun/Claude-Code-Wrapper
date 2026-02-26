"""
Entra ID (Azure AD) OAuth2 / OIDC helpers.

In MOCK_MODE the helpers return fake tokens without contacting Azure.
In production mode they call the real Microsoft identity platform endpoints.
"""

from datetime import datetime, timedelta, timezone
from typing import Optional
from urllib.parse import urlencode
import uuid

import httpx
import jwt as pyjwt

from app.config import Settings, get_settings


# ─── Authorization URL ───────────────────────────────────────────────────────


def get_authorization_url(redirect_uri: str, state: str, settings: Optional[Settings] = None) -> str:
    """Build the Entra ID OAuth2 authorize URL."""
    settings = settings or get_settings()

    params = {
        "client_id": settings.ENTRA_CLIENT_ID,
        "response_type": "code",
        "redirect_uri": redirect_uri,
        "response_mode": "query",
        "scope": "openid profile email",
        "state": state,
    }
    return f"{settings.entra_authorize_url}?{urlencode(params)}"


# ─── Token Exchange ──────────────────────────────────────────────────────────


async def exchange_code_for_token(
    code: str,
    redirect_uri: str,
    settings: Optional[Settings] = None,
) -> dict:
    """
    Exchange an OAuth2 authorization code for an access token.

    Returns a dict with at least {"access_token": "...", "token_type": "bearer"}.
    In MOCK_MODE a locally-signed JWT is returned.
    """
    settings = settings or get_settings()

    if settings.MOCK_MODE:
        return _mock_token(settings)

    # ── Real Entra ID call ──
    async with httpx.AsyncClient() as client:
        response = await client.post(
            settings.entra_token_url,
            data={
                "client_id": settings.ENTRA_CLIENT_ID,
                "client_secret": settings.ENTRA_CLIENT_SECRET,
                "grant_type": "authorization_code",
                "code": code,
                "redirect_uri": redirect_uri,
                "scope": "openid profile email",
            },
        )
        response.raise_for_status()
        return response.json()


# ─── Token Validation ────────────────────────────────────────────────────────


def validate_token(token: str, settings: Optional[Settings] = None) -> Optional[dict]:
    """
    Validate a bearer token.

    In MOCK_MODE, decodes using the local JWT_SECRET.
    Returns the decoded payload dict on success, or None on failure.
    """
    settings = settings or get_settings()

    if settings.MOCK_MODE:
        try:
            payload = pyjwt.decode(
                token,
                settings.JWT_SECRET,
                algorithms=[settings.JWT_ALGORITHM],
                audience=settings.ENTRA_CLIENT_ID,
                issuer=settings.entra_authority,
            )
            return payload
        except pyjwt.PyJWTError:
            return None

    # In production you would verify against Entra ID JWKS:
    # 1. Fetch JWKS from https://login.microsoftonline.com/{tenant}/discovery/v2.0/keys
    # 2. Decode + verify audience, issuer, expiration
    # Placeholder — always reject in non-mock prod until JWKS is wired up
    return None


# ─── Helpers ──────────────────────────────────────────────────────────────────


def _mock_token(settings: Settings) -> dict:
    """Generate a locally-signed mock JWT for development / testing."""
    now = datetime.now(timezone.utc)
    payload = {
        "sub": "mock-user-id",
        "name": "Mock User",
        "email": "mock.user@enterprise.local",
        "oid": str(uuid.uuid4()),
        "iat": now,
        "exp": now + timedelta(minutes=settings.JWT_EXPIRATION_MINUTES),
        "iss": settings.entra_authority,
        "aud": settings.ENTRA_CLIENT_ID,
    }
    access_token = pyjwt.encode(payload, settings.JWT_SECRET, algorithm=settings.JWT_ALGORITHM)
    return {
        "access_token": access_token,
        "token_type": "bearer",
        "expires_in": settings.JWT_EXPIRATION_MINUTES * 60,
    }

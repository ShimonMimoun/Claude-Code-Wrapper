"""
Bearer token authentication dependency for FastAPI.

Extracts the JWT from the Authorization header and validates it.
In MOCK_MODE, any non-empty token structured as a JWT is accepted.
"""

from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer

from app.auth.entra import validate_token
from app.config import Settings, get_settings

_bearer_scheme = HTTPBearer(auto_error=True)


async def require_auth(
    credentials: HTTPAuthorizationCredentials = Depends(_bearer_scheme),
    settings: Settings = Depends(get_settings),
) -> dict:
    """
    FastAPI dependency that enforces a valid bearer token.

    Returns the decoded JWT payload dict on success.
    Raises HTTP 401 on failure.
    """
    token = credentials.credentials

    payload = validate_token(token, settings)
    if payload is None:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid or expired token",
            headers={"WWW-Authenticate": "Bearer"},
        )

    return payload

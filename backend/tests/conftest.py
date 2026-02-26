"""
Shared pytest fixtures and test configuration.

Provides a pre-configured TestClient and a valid mock JWT token
that can be used across all test modules.
"""

import pytest
from fastapi.testclient import TestClient

from app.main import app
from app.auth.entra import _mock_token
from app.config import get_settings


@pytest.fixture(scope="session")
def settings():
    """Return the application settings (mock mode enabled by default)."""
    return get_settings()


@pytest.fixture(scope="session")
def client():
    """FastAPI test client — shares a single instance across all tests."""
    with TestClient(app) as c:
        yield c


@pytest.fixture(scope="session")
def auth_token(settings) -> str:
    """A valid mock JWT bearer token for authenticated requests."""
    token_data = _mock_token(settings)
    return token_data["access_token"]


@pytest.fixture(scope="session")
def auth_headers(auth_token) -> dict:
    """Authorization headers dict ready to pass to client requests."""
    return {"Authorization": f"Bearer {auth_token}"}

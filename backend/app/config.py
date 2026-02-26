"""
Application configuration loaded from environment variables.
Uses pydantic-settings for type-safe config with .env support.
"""

from pydantic_settings import BaseSettings
from functools import lru_cache


class Settings(BaseSettings):
    """All application settings, loaded from env vars or .env file."""

    # ── App ──
    APP_NAME: str = "Wrapper AI Proxy"
    APP_VERSION: str = "1.0.0"
    DEBUG: bool = True
    MOCK_MODE: bool = True  # When True, all external calls are mocked

    # ── Entra ID (Azure AD) ──
    ENTRA_TENANT_ID: str = "mock-tenant-id"
    ENTRA_CLIENT_ID: str = "mock-client-id"
    ENTRA_CLIENT_SECRET: str = "mock-client-secret"
    ENTRA_REDIRECT_URI: str = "http://localhost:8000/sso/callback"
    ENTRA_AUTHORITY: str = ""  # computed in property

    # ── JWT ──
    JWT_SECRET: str = "super-secret-mock-key-change-in-prod"
    JWT_ALGORITHM: str = "HS256"
    JWT_EXPIRATION_MINUTES: int = 60

    # ── CLI Binaries ──
    CLI_BINARIES_DIR: str = "./cli_binaries"

    @property
    def entra_authority(self) -> str:
        return self.ENTRA_AUTHORITY or f"https://login.microsoftonline.com/{self.ENTRA_TENANT_ID}"

    @property
    def entra_authorize_url(self) -> str:
        return f"{self.entra_authority}/oauth2/v2.0/authorize"

    @property
    def entra_token_url(self) -> str:
        return f"{self.entra_authority}/oauth2/v2.0/token"

    model_config = {"env_file": ".env", "env_file_encoding": "utf-8", "extra": "ignore"}


@lru_cache()
def get_settings() -> Settings:
    """Cached singleton settings instance."""
    return Settings()

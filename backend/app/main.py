"""
Wrapper AI Proxy — FastAPI application entry point.

Creates and configures the FastAPI app, registers all route modules,
and sets up CORS middleware.
"""

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.config import get_settings
from app.routes import sso, models, settings, cli, health


def create_app() -> FastAPI:
    """Application factory — builds and returns the configured FastAPI instance."""
    _settings = get_settings()

    application = FastAPI(
        title=_settings.APP_NAME,
        version=_settings.APP_VERSION,
        description=(
            "Enterprise AI Proxy backend for the Wrapper AI CLI. "
            "Handles SSO authentication via Entra ID, model catalog, "
            "enterprise settings distribution, and CLI binary downloads."
        ),
        docs_url="/docs",
        redoc_url="/redoc",
    )

    # ── CORS ─────────────────────────────────────────────────────────────
    application.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],  # Restrict in production
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    # ── Routers ──────────────────────────────────────────────────────────
    application.include_router(health.router)
    application.include_router(sso.router)
    application.include_router(models.router)
    application.include_router(settings.router)
    application.include_router(cli.router)

    return application


app = create_app()

"""
SSO routes: login redirect and OAuth2 callback.

GET /sso/login?redirect=<cli_callback_url>
    → Redirects to Entra ID (or mock login page).

GET /sso/callback?code=<auth_code>&state=<cli_redirect>
    → Exchanges code for token, redirects CLI with ?token=<jwt>.
"""

from urllib.parse import urlencode

from fastapi import APIRouter, Depends, Query
from fastapi.responses import RedirectResponse, HTMLResponse

from app.auth.entra import exchange_code_for_token, get_authorization_url
from app.config import Settings, get_settings

router = APIRouter(prefix="/sso", tags=["SSO Authentication"])


@router.get("/login")
async def sso_login(
    redirect: str = Query(..., description="CLI callback URL (e.g. http://127.0.0.1:8080/callback)"),
    settings: Settings = Depends(get_settings),
):
    """
    Initiate SSO login flow.

    In MOCK_MODE: immediately generates a token and redirects back to the CLI.
    In production: redirects the browser to Entra ID's authorize endpoint.
    """
    if settings.MOCK_MODE:
        # Skip Entra ID entirely — generate a token and redirect back to CLI
        token_data = await exchange_code_for_token(
            code="mock-code",
            redirect_uri=redirect,
            settings=settings,
        )
        callback_url = f"{redirect}?token={token_data['access_token']}"
        return RedirectResponse(url=callback_url)

    # Production: redirect to Entra ID with state = CLI redirect URI
    auth_url = get_authorization_url(
        redirect_uri=settings.ENTRA_REDIRECT_URI,
        state=redirect,  # We stash the CLI callback in `state` to use after code exchange
        settings=settings,
    )
    return RedirectResponse(url=auth_url)


@router.get("/callback")
async def sso_callback(
    code: str = Query("", description="OAuth2 authorization code"),
    state: str = Query("", description="CLI callback URL (passed as state)"),
    settings: Settings = Depends(get_settings),
):
    """
    OAuth2 callback endpoint — Entra ID redirects here after user login.

    Exchanges the authorization code for a JWT, then redirects the browser
    back to the CLI's local callback server with ?token=<jwt>.
    """
    if not code:
        return HTMLResponse(
            content="<h1>Error</h1><p>No authorization code received.</p>",
            status_code=400,
        )

    token_data = await exchange_code_for_token(
        code=code,
        redirect_uri=settings.ENTRA_REDIRECT_URI,
        settings=settings,
    )

    access_token = token_data.get("access_token", "")

    if not access_token:
        return HTMLResponse(
            content="<h1>Error</h1><p>Token exchange failed.</p>",
            status_code=500,
        )

    # Redirect back to the CLI's local HTTP server with the token
    cli_redirect = state or "http://127.0.0.1:8080/callback"
    callback_url = f"{cli_redirect}?token={access_token}"
    return RedirectResponse(url=callback_url)

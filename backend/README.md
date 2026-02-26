# Wrapper AI Proxy — Backend

Enterprise backend that powers the **Wrapper AI CLI** for Claude Code.

## Features

| Feature | Endpoint | Auth |
|---|---|---|
| Health check | `GET /health` | ✗ |
| SSO Login | `GET /sso/login?redirect=...` | ✗ |
| SSO Callback | `GET /sso/callback?code=...&state=...` | ✗ |
| Model catalog | `GET /v1/models` | ✓ |
| Enterprise settings | `GET /api/claude-settings` | ✓ |
| CLI download | `GET /cli/{platform}` | ✓ |

## Quick Start

```bash
cd backend

# Create virtual env
python -m venv .venv && source .venv/bin/activate

# Install deps
pip install -r requirements.txt

# Run server
uvicorn app.main:app --reload --port 8000
```

Open **http://localhost:8000/docs** for interactive Swagger UI.

## Running Tests

```bash
cd backend
pytest tests/ -v
```

All tests use **mock mode** — no external services required.

## Project Structure

```
backend/
├── app/
│   ├── main.py           # App factory
│   ├── config.py          # Settings (env vars)
│   ├── dependencies.py    # Shared DI
│   ├── auth/
│   │   ├── entra.py       # Entra ID OAuth2
│   │   └── bearer.py      # JWT bearer dependency
│   ├── routes/
│   │   ├── sso.py         # SSO login/callback
│   │   ├── models.py      # /v1/models
│   │   ├── settings.py    # /api/claude-settings
│   │   ├── cli.py         # /cli/{platform}
│   │   └── health.py      # /health
│   ├── services/
│   │   ├── model_service.py
│   │   ├── settings_service.py
│   │   └── cli_service.py
│   └── schemas/
│       ├── auth.py
│       ├── models.py
│       └── settings.py
└── tests/
    ├── conftest.py        # Shared fixtures
    ├── test_health.py
    ├── test_sso.py
    ├── test_models.py
    ├── test_settings.py
    └── test_cli.py
```

## Configuration

All settings are loaded from environment variables (or `.env` file):

| Variable | Default | Description |
|---|---|---|
| `MOCK_MODE` | `True` | Skip real Entra ID calls |
| `ENTRA_TENANT_ID` | `mock-tenant-id` | Azure AD tenant |
| `ENTRA_CLIENT_ID` | `mock-client-id` | App registration client ID |
| `ENTRA_CLIENT_SECRET` | `mock-client-secret` | App registration secret |
| `JWT_SECRET` | `super-secret-...` | Signing key for mock JWTs |

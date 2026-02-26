"""Tests for GET /api/claude-settings."""


class TestSettings:
    """Enterprise settings endpoint tests."""

    def test_settings_authenticated(self, client, auth_headers):
        """Authenticated request should return enterprise settings."""
        response = client.get("/api/claude-settings", headers=auth_headers)
        assert response.status_code == 200

        data = response.json()
        assert "env" in data
        assert "permissions" in data
        assert "allowedTools" in data
        assert data["env"]["ANTHROPIC_BASE_URL"] != ""

    def test_settings_unauthenticated(self, client):
        """Request without a token should return 403."""
        response = client.get("/api/claude-settings")
        assert response.status_code == 403

    def test_settings_invalid_token(self, client):
        """Request with an invalid token should return 401."""
        response = client.get(
            "/api/claude-settings",
            headers={"Authorization": "Bearer not-a-real-jwt"},
        )
        assert response.status_code == 401

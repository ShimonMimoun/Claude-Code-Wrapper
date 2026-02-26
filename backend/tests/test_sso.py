"""Tests for SSO login and callback routes."""


class TestSSOLogin:
    """GET /sso/login?redirect=..."""

    def test_login_redirect_in_mock_mode(self, client):
        """In mock mode, /sso/login should redirect back to the CLI with a token."""
        response = client.get(
            "/sso/login",
            params={"redirect": "http://127.0.0.1:8080/callback"},
            follow_redirects=False,
        )
        assert response.status_code == 307
        location = response.headers["location"]
        assert "http://127.0.0.1:8080/callback" in location
        assert "token=" in location

    def test_login_missing_redirect_param(self, client):
        """Missing redirect parameter should return 422."""
        response = client.get("/sso/login")
        assert response.status_code == 422


class TestSSOCallback:
    """GET /sso/callback?code=...&state=..."""

    def test_callback_with_code(self, client):
        """Valid callback should redirect with a token."""
        response = client.get(
            "/sso/callback",
            params={"code": "test-auth-code", "state": "http://127.0.0.1:8080/callback"},
            follow_redirects=False,
        )
        assert response.status_code == 307
        location = response.headers["location"]
        assert "token=" in location

    def test_callback_without_code(self, client):
        """Missing authorization code should return 400."""
        response = client.get("/sso/callback")
        assert response.status_code == 400

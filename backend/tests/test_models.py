"""Tests for GET /v1/models."""


class TestModels:
    """Model listing endpoint tests."""

    def test_models_authenticated(self, client, auth_headers):
        """Authenticated request should return a list of models."""
        response = client.get("/v1/models", headers=auth_headers)
        assert response.status_code == 200

        data = response.json()
        assert data["object"] == "list"
        assert isinstance(data["data"], list)
        assert len(data["data"]) > 0

        # Verify model structure
        model = data["data"][0]
        assert "id" in model
        assert "object" in model
        assert model["object"] == "model"
        assert "created" in model
        assert "owned_by" in model

    def test_models_unauthenticated(self, client):
        """Request without a token should return 403."""
        response = client.get("/v1/models")
        assert response.status_code == 403

    def test_models_invalid_token(self, client):
        """Request with garbage token should return 401."""
        response = client.get(
            "/v1/models",
            headers={"Authorization": "Bearer invalid-garbage-token"},
        )
        assert response.status_code == 401

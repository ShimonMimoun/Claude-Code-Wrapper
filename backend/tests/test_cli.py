"""Tests for GET /cli/{platform}."""

import pytest


class TestCLIDownload:
    """CLI binary download endpoint tests."""

    @pytest.mark.parametrize("platform", ["win", "mac-intel", "mac-m-series", "linux"])
    def test_download_valid_platform(self, client, auth_headers, platform):
        """Valid platforms should return a binary download."""
        response = client.get(f"/cli/{platform}", headers=auth_headers)
        assert response.status_code == 200
        assert response.headers["content-type"] == "application/octet-stream"
        assert "content-disposition" in response.headers
        assert len(response.content) > 0

    def test_download_unknown_platform(self, client, auth_headers):
        """Unknown platform should return 404."""
        response = client.get("/cli/freebsd", headers=auth_headers)
        assert response.status_code == 404

    def test_download_unauthenticated(self, client):
        """Request without a token should return 403."""
        response = client.get("/cli/linux")
        assert response.status_code == 403

    def test_download_win_has_exe_extension(self, client, auth_headers):
        """Windows binary should have .exe in the filename."""
        response = client.get("/cli/win", headers=auth_headers)
        assert response.status_code == 200
        disposition = response.headers["content-disposition"]
        assert ".exe" in disposition

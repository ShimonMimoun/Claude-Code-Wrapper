"""
Mock CLI binary download service.

In production this would serve real platform-specific binaries.
In mock mode it returns a tiny placeholder binary.
"""

from fastapi import HTTPException, status
from fastapi.responses import Response

_VALID_PLATFORMS = {"win", "mac-intel", "mac-m-series", "linux"}

# A tiny ELF / Mach-O stub is overkill for testing — we serve plain bytes.
_MOCK_BINARY = b"MOCK_CLI_BINARY_PLACEHOLDER_v1.0.0"


def get_cli_binary(platform: str) -> Response:
    """
    Return a binary download response for the given platform.

    Raises HTTP 404 for unsupported platforms.
    """
    if platform not in _VALID_PLATFORMS:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Unknown platform '{platform}'. Supported: {', '.join(sorted(_VALID_PLATFORMS))}",
        )

    filename = f"claude-{platform}" + (".exe" if platform == "win" else "")

    return Response(
        content=_MOCK_BINARY,
        media_type="application/octet-stream",
        headers={"Content-Disposition": f'attachment; filename="{filename}"'},
    )

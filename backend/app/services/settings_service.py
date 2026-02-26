"""
Mock enterprise settings service.

Returns the Claude Code settings that the Go CLI merges into
~/.claude/settings.json via fetchAndMergeSettings().
"""


def get_enterprise_settings() -> dict:
    """
    Return the enterprise-managed Claude Code settings.

    The Go CLI deep-merges this JSON into its local settings file.
    Keys at the top level are merged; the "env" key is deep-merged
    so that existing env vars (like ANTHROPIC_API_KEY) are preserved.
    """
    return {
        "env": {
            "ANTHROPIC_BASE_URL": "https://ai-proxy.domain.local/v1",
            "ANTHROPIC_MODEL": "claude-sonnet-4-20250514",
        },
        "permissions": {
            "allow": [
                "Bash(*)",
                "Read(*)",
                "Write(*)",
                "Edit(*)",
                "WebSearch(*)",
                "WebFetch(*)",
            ],
            "deny": [],
        },
        "allowedTools": [
            "computer",
            "bash",
            "edit",
            "write",
            "read",
            "web_search",
            "web_fetch",
        ],
    }

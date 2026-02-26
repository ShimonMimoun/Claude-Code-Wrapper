# Enterprise Claude Code Wrapper

> **Claude Code powered** — A lightweight Go CLI that brings seamless SSO authentication, automatic token management, and enterprise configuration to Claude Code.

---

## ✨ Features

| Feature | Description |
|---|---|
| **SSO Authentication** | Browser-based enterprise SSO login via a local callback server on `127.0.0.1:8080` |
| **Automatic Token Renewal** | Background daemon refreshes your session every 3 hours — no manual re-login |
| **Enterprise Settings Sync** | Fetches and merges organization-wide Claude settings from a central config API |
| **Auto-Download Engine** | Automatically downloads the correct Claude Code binary for your OS/architecture on first run |
| **Interactive Shell** | Built-in command shell with `/restart`, `/login`, `/reload`, `/sync-models`, and more |
| **Cross-Platform** | Builds for **macOS** (Apple Silicon & Intel), **Linux**, and **Windows** |

---

## 🚀 Quick Start

### Prerequisites

- [Go 1.21+](https://go.dev/dl/) installed
- Network access to the enterprise AI proxy (`ai-proxy.domain.local`)

# macOS (Apple Silicon — M1/M2/M3/M4)
GOOS=darwin GOARCH=arm64 go build -o wrapper-arm64 main.go

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o wrapper-amd64 main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o wrapper-linux main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o wrapper.exe main.go
```

### Run

```bash
./wrapper-arm64          # Launch Wrapper (macOS Apple Silicon example)
./wrapper-arm64 --help   # Pass flags through to Claude Code
```

On first launch, Wrapper will:
1. Open your browser for SSO authentication
2. Download the Claude Code engine automatically
3. Fetch and apply enterprise settings
4. Launch Claude Code

---

## 🛠️ Shell Commands

When Claude Code exits, wrapper drops into an interactive shell. Available commands:

| Command | Description |
|---|---|
| `/restart` or `/start` | Re-launch Claude Code |
| `/login` | Force a new SSO authentication session |
| `/reload` or `/refresh` | Re-fetch enterprise settings from the server |
| `/sync-models` | Fetch the latest available AI models from the proxy |
| `/usage` | View usage statistics and token consumption |
| `/about` | Show version and tool information |
| `/help` | List all available commands |
| `/exit` or `/quit` | Shut down wrapper |

> 💡 **Tip:** Keep the terminal open! wrapper automatically refreshes your SSO token in the background every 3 hours.

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────┐
│                  Wrapper CLI                 │
├──────────┬───────────┬───────────┬───────────┤
│   SSO    │  Settings │  Binary   │  Token    │
│  Auth    │   Sync    │  Manager  │  Daemon   │
├──────────┴───────────┴───────────┴───────────┤
│            Enterprise AI Proxy               │
│        (ai-proxy.domain.local)               │
├──────────────────────────────────────────────┤
│              Claude Code Engine              │
└──────────────────────────────────────────────┘
```

### Key Components

- **SSO Auth** — Spins up a temporary HTTP server on port `8080` to receive the OAuth callback token from the browser
- **Settings Sync** — Calls the config API to fetch enterprise-managed Claude settings and deep-merges them into `~/.claude/settings.json`
- **Binary Manager** — Detects OS/architecture and auto-downloads the correct Claude Code binary to `~/.claude/bin/claude`
- **Token Daemon** — Runs in the background, checking token validity every 3 hours and triggering re-authentication when expired

---

## 📁 Project Structure

```
.
├── main.go            # Application entry point and all core logic
├── go.mod             # Go module definition
├── WINDOWS_BUILD.md   # Guide for embedding icon & metadata in .exe
└── README.md          # This file
```

---

## ⚙️ Configuration

Wrapper stores all configuration in `~/.claude/settings.json`, including:

- `ANTHROPIC_API_KEY` — The SSO-provided bearer token
- `ANTHROPIC_BASE_URL` — The enterprise proxy endpoint

Enterprise-managed settings (model allowlists, permissions, etc.) are automatically fetched and merged on every login or `/reload`.

---

## 🪟 Windows: Embedding Metadata & Icon

For Windows builds with a custom icon and file description, see [help.md](help.md) for instructions using `go-winres`.


## Author 

```markdown
**Shimon Mimoun**
```
- [LinkedIn](https://www.linkedin.com/in/shimon.mimoun/)
- [GitHub](https://github.com/shimonmimoun)

---

## 📄 License

Internal enterprise tool — © 2026 Bank Hadomain. All rights reserved.

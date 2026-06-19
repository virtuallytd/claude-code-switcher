# Claude Code Switcher

A CLI tool for running [Claude Code](https://claude.ai/code) in isolated Podman containers, each with its own authentication, MCP servers, and settings.

## Features

- **Profile Isolation** — each profile runs in its own container with separate auth, MCP servers, and settings
- **Dual Auth Support** — Google Vertex AI (ADC) and direct Anthropic API key profiles side by side
- **MCP Server Management** — per-profile MCP server configurations, automatically detected and loaded
- **Simultaneous Sessions** — run multiple profiles (or the same profile) in different terminals at the same time
- **Interactive Setup** — `ccs init` walks you through creating a new profile
- **Container-Based** — Podman containers ensure clean separation with no credential leakage between profiles

## Installation

### Homebrew

```bash
brew install virtuallytd/tap/claude-code-switcher
```

### From Source

```bash
git clone https://github.com/virtuallytd/claude-code-switcher
cd claude-code-switcher
make build
```

### Prerequisites

- [Podman](https://podman.io/) — `brew install podman`
- A running Podman machine — `podman machine init && podman machine start`

## Quick Start

```bash
# Build the container image (one-time)
ccs build

# Create a profile
ccs init work
# Auth type (api/vertex): vertex
# Vertex project ID: my-gcp-project
# Vertex region [global]: global
# Model [claude-sonnet-4-6]: claude-opus-4-6

# Launch Claude Code
ccs work ~/Projects/my-repo
```

## Usage

```bash
ccs <profile> [path] [-- claude-args...]    # Launch Claude Code with a profile
ccs init <name>                              # Create a new profile interactively
ccs build                                    # Build/rebuild the container image
ccs profiles                                 # List available profiles
ccs status                                   # Show running containers
ccs stop [profile]                           # Stop container(s)
ccs --version                                # Show version
```

### Running Multiple Sessions

Open separate terminals and launch different profiles — or the same profile on different directories:

```bash
# Terminal 1
ccs work ~/Projects/work/api-service

# Terminal 2
ccs personal ~/Projects/personal/side-project

# Terminal 3
ccs work ~/Projects/work/frontend
```

### Adding MCP Servers

Create a `mcp.json` file in the profile directory:

```bash
cat > ~/.ccs_profiles/work/mcp.json << 'EOF'
{
  "mcpServers": {
    "my-server": {
      "type": "http",
      "url": "http://localhost:3001/mcp"
    }
  }
}
EOF
```

MCP servers running on localhost are automatically accessible from the container via host networking.

## How It Works

Each profile is a directory in `~/.ccs_profiles/` containing:

| File | Purpose |
|------|---------|
| `.env` | Auth credentials (API key or Vertex AI config) |
| `settings.json` | Claude Code settings (model, theme) |
| `settings.local.json` | Permissions and tool allowlists |
| `mcp.json` | MCP server definitions (optional) |

When you run `ccs <profile> <path>`, the tool:

1. Starts a Podman container from a shared image (Node 22 + Claude Code + gcloud CLI)
2. Copies profile settings into the container
3. Bind-mounts your project directory
4. Launches Claude Code with the profile's auth and MCP config

Containers are ephemeral (`--rm`) and isolated — nothing persists between sessions except your project files.

## License

[MIT](LICENSE)

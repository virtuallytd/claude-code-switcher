# ccs — Claude Code Switcher

A profile switcher for running Claude Code in isolated Podman containers. Two profiles — `work` (Google Vertex AI auth) and `personal` (Anthropic API key) — each with separate MCP servers, settings, and credentials.

## Setup

```bash
go build -o ccs .          # Build the CLI
./ccs build                # Build the Podman container image
```

## Usage

```bash
ccs work ~/Projects/work/my-repo       # Launch work profile
ccs personal ~/Projects/personal/app   # Launch personal profile
ccs work ~/repo -- --resume            # Pass args to claude
ccs build                              # Build/rebuild container image
ccs stop                               # Stop all containers
ccs stop work                          # Stop work containers only
ccs status                             # Show running containers
ccs profiles                           # List available profiles
```

## Architecture

- **`main.go`** — Go CLI (cobra) that shells out to `podman` to run Claude Code in isolated containers.
- **`Containerfile`** — Single image with Node 22, Claude Code (npm), and gcloud CLI. Runs as non-root user matching host UID/GID.
- **`~/.ccs_profiles/<name>/`** — Per-profile config stored outside the repo: `.env` (auth credentials), `settings.json`, `settings.local.json` (MCP servers + permissions). Profiles with `CLAUDE_CODE_USE_VERTEX=1` in `.env` automatically get `~/.config/gcloud` mounted.

Containers are named `cc-<profile>-<path-hash>` so multiple instances can run simultaneously in different terminals.

## Adding a new profile

1. Create `~/.ccs_profiles/<name>/` with `.env`, `settings.json`, and `settings.local.json`
2. Run `ccs <name> /path/to/project`

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Purpose

Go-based Claude Code profile switcher (`ccs`) — enables quick switching between Claude Code profiles with different MCP server configurations and authentication methods (standard Claude account vs. Vertex AI).

**Core functionality**:
- Interactive profile selection via bubbletea TUI
- Two authentication modes: standard Claude account (keyring-based) and Vertex AI (environment variables)
- MCP server configuration merging into active settings.json
- Cross-platform keyring support (macOS Keychain, Linux Secret Service)
- Self-contained binary with no external dependencies

## Architecture

### Go CLI + Shell Wrapper (Hybrid)

```
ccs (Go binary)
  ├─ Commands:
  │  ├─ switch    # Interactive TUI selector (bubbletea)
  │  ├─ reload    # Re-apply current profile without TUI
  │  └─ current   # Show active profile
  │
  ├─ Outputs shell commands to stdout:
  │  ├─ export statements (Vertex AI env vars)
  │  └─ unset statements (clear Vertex vars)
  │
  └─ Shell wrapper evals output:
     ccs() { eval "$(command ccs switch)" }
```

### Project Structure

```
main.go              # CLI entrypoint, command router
cmd/
  switch.go          # Interactive switcher with bubbletea TUI
  current.go         # Show active profile
internal/
  profile/           # Profile discovery, loading, env parsing
  keyring/           # Cross-platform keyring (99designs/keyring)
  config/            # Settings.json merging
```

### Profile Structure

Profiles live in `~/.claude/profiles/<name>/`:
- `settings.json` — MCP server configuration (required)
- `env.zsh` — Environment variables for Vertex AI auth (optional, presence indicates Vertex AI mode)

**Authentication modes**:
1. **Standard auth** (no `env.zsh`): Uses Claude account session tokens stored in system keyring under `"Claude Code - <profile-name>"`
2. **Vertex AI auth** (has `env.zsh`): Sources environment variables:
   - `CLAUDE_CODE_USE_VERTEX=1`
   - `CLOUD_ML_REGION=<region>`
   - `ANTHROPIC_VERTEX_PROJECT_ID=<project-id>`

**MCP server configuration format** in `settings.json`:
- SSE transport (local servers): `{"type": "sse", "url": "http://localhost:PORT/sse"}`
- HTTP transport (OAuth-based): `{"type": "http", "url": "https://..."}`
- NEVER use `"http"` for SSE servers — causes OAuth authentication errors

## Commands

**`ccs switch`** (or just `ccs` via shell wrapper)
- Interactive profile switcher using bubbletea TUI
- Discovers profiles in `~/.claude/profiles/`
- Clears Vertex AI environment variables on every switch (prevents leakage)
- For Vertex AI profiles: sources `env.zsh` and outputs export commands
- For standard profiles: restores session token from keyring to active `"Claude Code"` entry
- Merges `mcpServers` from profile's `settings.json` into active `~/.claude/settings.json`
- Saves current profile to `~/.claude/.current-profile`

**`ccs reload`**
- Re-applies the currently active profile without going through the TUI
- Useful after editing a profile's `settings.json` (e.g. adding a new MCP server)
- Same effect as re-selecting the current profile in `ccs switch`

**`ccs save <profile>`**
- Saves the current active session token to the keyring under `"Claude Code - <profile-name>"`
- Required once per standard-auth profile: `claude login && ccs save <profile>`
- Not needed for Vertex AI profiles (they use environment variables instead)

**`ccs current`**
- Shows active profile name and auth type
- Reads from `~/.claude/.current-profile` state file

**`ccs version`**
- Shows version number

**Shell wrapper**: Add to `.zshrc` or `.bashrc`:
```bash
ccs() {
  eval "$(command ccs switch)"
}
```

## Build & Install

**Build binary**:
```bash
go build -o ccs
```

**Install to PATH**:
```bash
sudo mv ccs /usr/local/bin/
```

**Add shell integration** to `~/.zshrc` or `~/.bashrc`:
```bash
ccs() {
  if [[ $# -eq 0 ]] || [[ "$1" == "switch" ]] || [[ "$1" == "reload" ]]; then
    eval "$(command ccs "$@")"
  else
    command ccs "$@"
  fi
}
```

**Reload shell**:
```bash
source ~/.zshrc
```

## Development Workflow

**Adding new commands**:
1. Create command file in `cmd/<command>.go`
2. Implement function with signature: `func CommandName() error`
3. Add case to switch in `main.go`

**Modifying TUI**:
- Edit `cmd/switch.go`
- TUI model implements bubbletea's `tea.Model` interface (Init, Update, View)
- Styles defined at package level using lipgloss

**No automated tests exist** — the codebase relies on manual testing.

**Testing profile discovery**:
```bash
go run . switch
```

**Testing keyring operations**:
- macOS: Check Keychain Access app for "Claude Code" entries
- Linux: Use `secret-tool` or `seahorse` to inspect Secret Service

## Key Implementation Details

**Profile discovery** (`internal/profile/profile.go`):
- Scans `~/.claude/profiles/` for directories
- Requires `settings.json` to be valid
- Detects Vertex AI profiles by presence of `env.zsh`

**Environment variable parsing** (`internal/profile.LoadEnvVars`):
- Parses `export VAR=value` lines from `env.zsh`
- Handles quoted and unquoted values
- Returns struct with parsed Vertex AI config

**Settings merging** (`internal/config/merge.go`):
- Loads profile's `settings.json`
- Reads `~/.claude.json` (Claude Code's main config, NOT `~/.claude/settings.json`)
- Replaces top-level `mcpServers` key while preserving all other state
- Writes back with pretty-printed JSON

**Keyring wrapper** (`internal/keyring/keyring.go`):
- Uses 99designs/keyring for cross-platform support
- macOS: Keychain backend
- Linux: Secret Service backend (gnome-keyring, KWallet)
- Service name: `"Claude Code"` for active token
- Profile tokens: `"Claude Code - <profile-name>"`

**TUI interaction** (`cmd/switch.go`):
- Arrow keys or `j`/`k` to navigate
- `Enter` to select
- `q`, `Esc`, `Ctrl+C` to quit
- Visual indicators: `●` (Vertex), `○` (standard)

**Shell command output**:
- Outputs to stdout (for eval)
- Error/success messages to stderr
- Prevents shell pollution

## Platform Support

- ✅ **macOS**: Full support (Keychain)
- ✅ **Linux**: Full support (Secret Service)
- ❌ **Windows**: Not supported (keyring backend needed)

## Dependencies

All embedded in binary:
- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — Terminal styling
- `github.com/99designs/keyring` — Cross-platform keyring

## Security Notes

- Session tokens stored in system keyring (encrypted at rest)
- Vertex AI credentials rely on `gcloud auth` (not stored in profiles)
- Profile switching clears Vertex environment variables before switching to prevent cross-contamination
- MCP server configs may contain sensitive URLs/settings — treat `settings.json` as sensitive data
- State file `~/.claude/.current-profile` is plain text (just profile name, no secrets)

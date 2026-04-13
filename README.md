# ccs - Claude Code Switcher

A lightweight CLI tool for switching between Claude Code profiles with different MCP server configurations and authentication methods.

## Features

- 🔄 **Quick Profile Switching**: Interactive TUI for selecting profiles
- 🔐 **Dual Auth Support**: Standard Claude account (keychain) and Vertex AI (env vars)
- ⚙️ **MCP Configuration**: Automatic merging of MCP server configs
- 🎨 **Beautiful TUI**: Clean, keyboard-driven interface with visual indicators
- 🔒 **Secure**: Credentials stored in system keyring (macOS Keychain / Linux Secret Service)
- 📦 **Self-Contained**: Single binary with no external dependencies

## Installation

### Homebrew (macOS/Linux)

```bash
brew install virtuallytd/tap/ccs
```

Then add shell integration to `~/.zshrc` or `~/.bashrc`:
```bash
ccs() {
  # If no arguments, "switch", or "reload", run with eval
  if [[ $# -eq 0 ]] || [[ "$1" == "switch" ]] || [[ "$1" == "reload" ]]; then
    eval "$(command ccs "$@")"
  else
    # Pass through other commands (save, current, version, help) directly
    command ccs "$@"
  fi
}
```

Reload your shell:
```bash
source ~/.zshrc
```

### Manual Install

1. Download the latest binary from [releases](https://github.com/virtuallytd/claude-code-switcher/releases)
2. Extract and move to your PATH:
   ```bash
   tar -xzf ccs_*.tar.gz
   sudo mv ccs /usr/local/bin/
   ```
3. Add the shell function above to your shell config

### From Source

```bash
git clone https://github.com/virtuallytd/claude-code-switcher
cd claude-code-switcher
make install
```

Then add the shell function above to your shell config.

## Usage

### Switch Profiles

Interactive profile selector:
```bash
ccs
```

This launches a TUI where you can:
- Navigate with arrow keys or `j`/`k`
- Select with `Enter`
- Quit with `q`, `Esc`, or `Ctrl+C`

### Reload Current Profile

Re-apply the current profile without interactive selection (useful after editing `settings.json`):
```bash
ccs reload
```

### Save Session Token

Save your current Claude Code session token for a profile (required once per standard auth profile):
```bash
claude login              # Login to Claude first
ccs save personal         # Save token for "personal" profile
```

This is only needed for standard auth profiles. Vertex AI profiles use `gcloud` credentials.

### Show Current Profile

```bash
ccs current
```

Output example:
```
work (vertex)
```

## Profile Structure

Profiles live in `~/.claude/profiles/<name>/`:

```
~/.claude/profiles/
├── personal/
│   └── settings.json          # MCP server config (required)
└── work/
    ├── settings.json          # MCP server config (required)
    └── env.zsh                # Vertex AI vars (optional)
```

### Standard Auth Profile (Claude Account)

Create directory and settings:
```bash
mkdir -p ~/.claude/profiles/personal
```

Add `settings.json`:
```json
{
  "mcpServers": {
    "server-name": {
      "type": "sse",
      "url": "http://localhost:3000/sse"
    }
  }
}
```

Then save your session token:
```bash
claude login
ccs save personal
```

### Vertex AI Profile

Create directory, settings, and env vars:
```bash
mkdir -p ~/.claude/profiles/work
```

Add `settings.json` (same format as above).

Add `env.zsh`:
```bash
export CLAUDE_CODE_USE_VERTEX=1
export CLOUD_ML_REGION=us-east5
export ANTHROPIC_VERTEX_PROJECT_ID=your-project-id
```

Vertex AI authentication relies on `gcloud auth` being configured.

## How It Works

### Switching Process

1. **Profile Discovery**: Scans `~/.claude/profiles/` for valid profiles
2. **Interactive Selection**: Shows TUI with profile list
3. **Environment Cleanup**: Clears Vertex AI vars (prevents leakage)
4. **Authentication**:
   - **Vertex AI profiles**: Sources `env.zsh` environment variables
   - **Standard profiles**: Restores session token from keyring to active `"Claude Code"` entry
5. **Settings Merge**: Merges `mcpServers` from profile into `~/.claude/settings.json`
6. **Shell Commands**: Outputs `export`/`unset` commands to stdout (eval'd by shell function)

### Visual Indicators

- `●` Filled dot = Currently active profile
- `○` Empty dot = Inactive profile
- Orange text = Vertex AI profile
- Blue text = Standard auth profile
- `❯` Cursor on selected item

## Platform Support

- ✅ **macOS**: Uses Keychain for secure token storage
- ✅ **Linux**: Uses Secret Service (gnome-keyring, KWallet)
- ❌ **Windows**: Not currently supported

## MCP Server Configuration

### Transport Types

**SSE (Server-Sent Events)** for local servers:
```json
{
  "type": "sse",
  "url": "http://localhost:3000/sse"
}
```

**HTTP** for OAuth-based remote servers:
```json
{
  "type": "http",
  "url": "https://api.example.com"
}
```

⚠️ **Important**: Never use `"http"` for SSE servers — it will cause OAuth authentication errors.

## Security

- **Session tokens**: Encrypted in system keyring (Keychain/Secret Service)
- **Vertex AI credentials**: Uses `gcloud auth` (not stored in profiles)
- **Environment isolation**: Vertex vars cleared before every switch
- **File permissions**: Settings files should be `0644` (readable by owner)

## Troubleshooting

### "No active Claude session found"

Run `claude login` first, then switch profiles.

### "No saved token for profile"

For standard auth profiles, you need to save the token:
```bash
claude login
ccs save <profile-name>
```

### Profile not appearing in list

Ensure the profile directory has a valid `settings.json` file.

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

### Build

```bash
make build              # Build binary
make test               # Run tests
make snapshot           # Test GoReleaser build locally
```

### Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [99designs/keyring](https://github.com/99designs/keyring) - Cross-platform keyring

### Project Structure

```
ccs/
├── main.go              # CLI entrypoint
├── cmd/
│   ├── switch.go        # Interactive switcher (TUI)
│   ├── current.go       # Show active profile
│   ├── reload.go        # Reload current profile
│   └── save.go          # Save session token
└── internal/
    ├── profile/         # Profile discovery & loading
    ├── keyring/         # Keyring wrapper
    └── config/          # Settings.json merging
```

## Security

See [SECURITY.md](SECURITY.md) for security policy and reporting vulnerabilities

## License

MIT

## Contributing

Issues and pull requests welcome!

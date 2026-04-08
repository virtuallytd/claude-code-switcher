# How CCS Works with Claude Code

This document explains how `ccs` (Claude Code Switcher) integrates with Claude Code's configuration system.

## Claude Code Configuration Files

### Active Configuration (What Claude Code Reads)

```
~/.claude/
├── settings.json              # Main active configuration
├── settings.local.json        # Local overrides (optional)
└── .current-profile           # Current profile name (ccs only)
```

**Claude Code reads from:**
- `~/.claude/settings.json` - Main configuration file
- `~/.claude/settings.local.json` - Overrides (takes precedence)
- Keychain entry `"Claude Code"` - Session token (standard auth)
- Environment variables - `CLAUDE_CODE_USE_VERTEX`, etc. (Vertex AI auth)

### Profile Storage (Templates)

```
~/.claude/profiles/
├── personal/
│   └── settings.json          # Profile-specific MCP config
└── work/
    ├── settings.json          # Profile-specific MCP config
    └── env.zsh                # Vertex AI variables (optional)
```

**Profiles are templates** - they store different MCP server configurations and auth settings, but are NOT directly read by Claude Code.

## How Profile Switching Works

When you run `ccs` and select a profile, here's what happens:

### Step 1: Profile Discovery

```go
// internal/profile/profile.go
profiles := Discover()  // Scans ~/.claude/profiles/
```

Finds all directories in `~/.claude/profiles/` that contain `settings.json`.

### Step 2: Interactive Selection

```go
// cmd/switch.go
selected := showTUI(profiles)  // User picks a profile
```

Displays a bubbletea TUI where you select the profile to switch to.

### Step 3: Environment Cleanup

```bash
# Always clear Vertex vars first (prevents leakage)
unset CLAUDE_CODE_USE_VERTEX
unset CLOUD_ML_REGION
unset ANTHROPIC_VERTEX_PROJECT_ID
```

This ensures that if you switch from a Vertex AI profile to a standard profile, the environment variables don't leak.

### Step 4: Authentication Setup

#### For Standard Claude Account Profiles (no `env.zsh`)

```go
// internal/keyring/keyring.go
token := GetProfileToken("personal")     // Get saved token from keyring
SetActiveToken(token)                    // Copy to active "Claude Code" entry
```

**Keychain entries:**
- `"Claude Code - personal"` → Profile's saved session token
- `"Claude Code - work"` → Another profile's token
- `"Claude Code"` → **Active token** (what Claude Code reads)

When you switch profiles, `ccs` copies the profile-specific token to the active entry.

#### For Vertex AI Profiles (has `env.zsh`)

```go
// internal/profile/profile.go
envVars := LoadEnvVars(profilePath)  // Parse env.zsh

// cmd/switch.go outputs to stdout:
fmt.Println("export CLAUDE_CODE_USE_VERTEX=1")
fmt.Println("export CLOUD_ML_REGION=us-east5")
fmt.Println("export ANTHROPIC_VERTEX_PROJECT_ID=project-id")
```

The shell wrapper (`eval "$(ccs switch)"`) executes these export commands in the current shell.

### Step 5: Settings Merge

```go
// internal/config/merge.go
profileSettings := LoadSettings("~/.claude/profiles/work/settings.json")
activeSettings := LoadSettings("~/.claude/settings.json")

// Replace ONLY the mcpServers key
activeSettings["mcpServers"] = profileSettings["mcpServers"]

// Write back to active config (preserves other settings)
SaveSettings("~/.claude/settings.json", activeSettings)
```

**What gets preserved:**
- Theme settings
- Model preferences
- Editor settings
- Any other non-MCP configuration

**What gets replaced:**
- `mcpServers` object (entire key)

### Step 6: State Tracking

```go
// internal/profile/profile.go
SetCurrentProfile("work")  // Saves to ~/.claude/.current-profile
```

This allows `ccs current` to show which profile is active.

## Example: Complete Profile Switch

### Before Switch

**Active config** (`~/.claude/settings.json`):
```json
{
  "theme": "dark",
  "model": "claude-opus-4",
  "mcpServers": {
    "personal-server": {
      "type": "sse",
      "url": "http://localhost:3000/sse"
    }
  }
}
```

**Keychain:**
- `"Claude Code"` → `token_abc123` (currently active)
- `"Claude Code - personal"` → `token_abc123`
- `"Claude Code - work"` → `token_xyz789`

**Environment:**
```bash
# No Vertex vars set
```

### Run: `ccs` → Select "work" profile

**Profile config** (`~/.claude/profiles/work/settings.json`):
```json
{
  "mcpServers": {
    "work-server": {
      "type": "sse",
      "url": "http://localhost:8080/sse"
    },
    "slack-server": {
      "type": "http",
      "url": "https://slack.example.com"
    }
  }
}
```

**Profile env** (`~/.claude/profiles/work/env.zsh`):
```bash
export CLAUDE_CODE_USE_VERTEX=1
export CLOUD_ML_REGION=us-east5
export ANTHROPIC_VERTEX_PROJECT_ID=my-project
```

### After Switch

**Active config** (`~/.claude/settings.json`):
```json
{
  "theme": "dark",              // ✅ Preserved
  "model": "claude-opus-4",     // ✅ Preserved
  "mcpServers": {               // ⬇️ Replaced from work profile
    "work-server": {
      "type": "sse",
      "url": "http://localhost:8080/sse"
    },
    "slack-server": {
      "type": "http",
      "url": "https://slack.example.com"
    }
  }
}
```

**Keychain:**
- `"Claude Code"` → `token_xyz789` ✅ (copied from work profile - if standard auth)
- `"Claude Code - personal"` → `token_abc123`
- `"Claude Code - work"` → `token_xyz789`

**Environment** (set in current shell):
```bash
export CLAUDE_CODE_USE_VERTEX=1
export CLOUD_ML_REGION=us-east5
export ANTHROPIC_VERTEX_PROJECT_ID=my-project
```

**State file** (`~/.claude/.current-profile`):
```
work
```

## Shell Wrapper: Why It's Required

The Go binary runs in a subprocess and cannot modify the parent shell's environment. The shell wrapper solves this:

```bash
# ~/.zshrc or ~/.bashrc
ccs() {
  # If no arguments or "switch" argument, run interactive switcher
  if [[ $# -eq 0 ]] || [[ "$1" == "switch" ]]; then
    eval "$(command ccs switch)"
  else
    # Pass through other commands (current, version, help) directly
    command ccs "$@"
  fi
}
```

### What happens:

1. `command ccs switch` runs the Go binary
2. Go binary outputs shell commands to stdout:
   ```bash
   unset CLAUDE_CODE_USE_VERTEX
   unset CLOUD_ML_REGION
   unset ANTHROPIC_VERTEX_PROJECT_ID
   export CLAUDE_CODE_USE_VERTEX=1
   export CLOUD_ML_REGION=us-east5
   export ANTHROPIC_VERTEX_PROJECT_ID=my-project
   ```
3. `eval` executes these commands in the current shell
4. Environment variables are now set for this shell session

**Without the wrapper:** Environment variables won't be set, and Vertex AI auth won't work.

**With the wrapper:** Both settings.json updates AND environment variables work correctly.

## File Modification Summary

| File/Location | Modified? | When? | How? |
|--------------|-----------|-------|------|
| `~/.claude/settings.json` | ✅ Always | Every switch | `mcpServers` replaced |
| `~/.claude/settings.local.json` | ❌ Never | - | Untouched |
| `~/.claude/.current-profile` | ✅ Always | Every switch | Profile name saved |
| Keychain `"Claude Code"` | ✅ Standard auth | Standard profile | Token copied from profile entry |
| Environment vars | ✅ Vertex auth | Vertex profile | Exported via shell wrapper |
| Profile configs | ❌ Never | - | Read-only templates |

## Integration with Version-Controlled Profiles

If your profiles are symlinked from a dotfiles repo (like `mac-setup`):

```bash
~/.claude/profiles/work/settings.json
  ↓ (symlink)
~/Projects/private/mac-setup/configs/claude/profiles/work/settings.json
```

**Benefits:**
- Profiles are version-controlled
- Can sync across machines via git
- `ccs` doesn't care about symlinks (just reads the files)
- Edit source files in your dotfiles repo, changes appear immediately

## Authentication Deep Dive

### Standard Auth Flow

1. **First time setup:**
   ```bash
   claude login                          # Login to Claude
   # Token is stored in keychain as "Claude Code"

   # (Future feature: ccs save personal)
   # For now, manually copy with security command
   ```

2. **When switching:**
   ```go
   token := keyring.Get("Claude Code - personal")
   keyring.Set("Claude Code", token)
   ```

3. **Claude Code reads:**
   - Looks for keychain entry `"Claude Code"`
   - Uses that token for API requests

### Vertex AI Auth Flow

1. **Prerequisites:**
   ```bash
   gcloud auth application-default login
   gcloud config set project my-project
   ```

2. **Profile has env.zsh:**
   ```bash
   export CLAUDE_CODE_USE_VERTEX=1
   export CLOUD_ML_REGION=us-east5
   export ANTHROPIC_VERTEX_PROJECT_ID=my-project
   ```

3. **When switching:**
   ```bash
   eval "$(ccs switch)"  # Sets environment variables
   ```

4. **Claude Code reads:**
   - Detects `CLAUDE_CODE_USE_VERTEX=1`
   - Uses Vertex AI API with `gcloud` credentials
   - Ignores keychain (not needed for Vertex)

## Common Scenarios

### Switching Between Two Standard Profiles

```bash
ccs  # Select "personal"
# → Keychain "Claude Code" = personal token
# → settings.json mcpServers = personal servers

ccs  # Select "work-standard"
# → Keychain "Claude Code" = work-standard token
# → settings.json mcpServers = work servers
```

### Switching from Vertex to Standard

```bash
# Currently on work (Vertex)
echo $CLAUDE_CODE_USE_VERTEX  # → 1

ccs  # Select "personal" (standard)
# → Env vars cleared (unset)
# → Keychain "Claude Code" = personal token
# → settings.json mcpServers = personal servers

echo $CLAUDE_CODE_USE_VERTEX  # → (empty)
```

### Switching from Standard to Vertex

```bash
# Currently on personal (standard)
echo $CLAUDE_CODE_USE_VERTEX  # → (empty)

ccs  # Select "work" (Vertex)
# → Env vars set (export)
# → Keychain unchanged (Vertex doesn't use it)
# → settings.json mcpServers = work servers

echo $CLAUDE_CODE_USE_VERTEX  # → 1
```

## Security Considerations

### What's Encrypted

- ✅ **Keychain tokens** - Encrypted by macOS/Linux Secret Service
- ✅ **Vertex credentials** - Managed by `gcloud` (encrypted separately)

### What's Plain Text

- ❌ `settings.json` - Contains MCP server URLs (may include localhost ports)
- ❌ `env.zsh` - Contains project IDs and region names (not secrets, but metadata)
- ❌ `.current-profile` - Just a profile name

### Preventing Cross-Contamination

**Why we clear Vertex vars:**
```bash
# Without clearing:
ccs  # Switch from work (Vertex) to personal (standard)
# → Environment still has CLAUDE_CODE_USE_VERTEX=1
# → Claude Code tries to use Vertex API with standard token
# → AUTH FAILURE

# With clearing:
unset CLAUDE_CODE_USE_VERTEX  # First thing ccs does
# → Claude Code correctly uses standard auth
```

## Troubleshooting

### "No saved token for profile"

**Problem:** Standard auth profile, but no keyring entry exists.

**Solution:**
```bash
claude login
# Then manually save token (future: ccs save <profile>)
security add-generic-password -U \
  -s "Claude Code - personal" \
  -a "$(whoami)" \
  -w "$(security find-generic-password -s 'Claude Code' -w)"
```

### Profile switch doesn't affect Claude Code

**Problem 1:** Shell wrapper not configured.

**Solution:** Add to `~/.zshrc`:
```bash
ccs() {
  eval "$(command ccs switch)"
}
```

**Problem 2:** Claude Code was already running.

**Solution:** Restart Claude Code after switching profiles.

### Vertex auth not working

**Problem 1:** Environment vars not set.

**Check:**
```bash
echo $CLAUDE_CODE_USE_VERTEX  # Should be "1"
echo $ANTHROPIC_VERTEX_PROJECT_ID  # Should be your project
```

**Problem 2:** `gcloud` not authenticated.

**Solution:**
```bash
gcloud auth application-default login
gcloud config set project your-project-id
```

## Summary

`ccs` acts as a **profile switcher** that:
1. Reads profile templates from `~/.claude/profiles/`
2. Merges MCP config into `~/.claude/settings.json`
3. Sets up authentication (keychain for standard, env vars for Vertex)
4. Tracks current profile state

It **does NOT:**
- Modify profile templates (read-only)
- Change non-MCP settings (preserves your preferences)
- Require Claude Code restart (settings auto-reload)
- Store credentials in plain text (uses system keyring)

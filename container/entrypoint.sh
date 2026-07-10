#!/bin/bash
CLAUDE_HOME="${HOME}/.claude"
mkdir -p "$CLAUDE_HOME"

cp /ccs-profile/settings.json "$CLAUDE_HOME/settings.json" 2>/dev/null
cp /ccs-profile/settings.local.json "$CLAUDE_HOME/settings.local.json" 2>/dev/null

if [ -f /ccs-profile/mcp.json ]; then
  cp /ccs-profile/mcp.json "$(pwd)/.mcp.json"
fi

# Restore .claude.json from backup if it doesn't exist.
# Claude Code stores this at ~/.claude.json (outside ~/.claude/) so it doesn't
# survive container restarts. Backups are inside ~/.claude/ which is mounted.
if [ ! -f "${HOME}/.claude.json" ]; then
  latest=$(ls -t "${CLAUDE_HOME}/backups/.claude.json.backup."* 2>/dev/null | head -1)
  if [ -n "$latest" ]; then
    cp "$latest" "${HOME}/.claude.json"
  fi
fi

# Write the OAuth token to Claude Code's credential store so it recognizes
# the session as authenticated without going through the interactive login flow.
if [ -n "$CLAUDE_CODE_OAUTH_TOKEN" ]; then
  cat > "${CLAUDE_HOME}/.credentials.json" <<EOF
{"claudeAiOauth":{"accessToken":"${CLAUDE_CODE_OAUTH_TOKEN}","refreshToken":"","scopes":["user:inference"]}}
EOF
  chmod 600 "${CLAUDE_HOME}/.credentials.json"
fi

exec claude "$@"

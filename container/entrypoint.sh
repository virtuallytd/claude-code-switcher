#!/bin/bash
CLAUDE_HOME="${HOME}/.claude"
CLAUDE_JSON="${HOME}/.claude.json"
CLAUDE_JSON_PERSIST="${CLAUDE_HOME}/.claude.json.persist"

mkdir -p "$CLAUDE_HOME"

cp /ccs-profile/settings.json "$CLAUDE_HOME/settings.json" 2>/dev/null
cp /ccs-profile/settings.local.json "$CLAUDE_HOME/settings.local.json" 2>/dev/null

if [ -f /ccs-profile/mcp.json ]; then
  cp /ccs-profile/mcp.json "$(pwd)/.mcp.json"
fi

# Restore .claude.json — it lives at ~/.claude.json (outside the mounted ~/.claude/)
# so it doesn't survive container restarts. We persist a copy inside the mount.
if [ -f "$CLAUDE_JSON_PERSIST" ]; then
  cp "$CLAUDE_JSON_PERSIST" "$CLAUDE_JSON"
elif [ ! -f "$CLAUDE_JSON" ]; then
  latest=$(ls -t "${CLAUDE_HOME}/backups/.claude.json.backup."* 2>/dev/null | head -1)
  if [ -n "$latest" ]; then
    cp "$latest" "$CLAUDE_JSON"
  fi
fi

# Save .claude.json back to the mounted volume on exit.
save_state() {
  if [ -f "$CLAUDE_JSON" ]; then
    cp "$CLAUDE_JSON" "$CLAUDE_JSON_PERSIST"
  fi
}
trap save_state EXIT

if [ -n "$CLAUDE_CODE_OAUTH_TOKEN" ]; then
  cat > "${CLAUDE_HOME}/.credentials.json" <<EOF
{"claudeAiOauth":{"accessToken":"${CLAUDE_CODE_OAUTH_TOKEN}","refreshToken":"","scopes":["user:inference"]}}
EOF
  chmod 600 "${CLAUDE_HOME}/.credentials.json"
fi

claude "$@"

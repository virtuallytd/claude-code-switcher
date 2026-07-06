#!/bin/bash
CLAUDE_HOME="${HOME}/.claude"
mkdir -p "$CLAUDE_HOME"

cp /ccs-profile/settings.json "$CLAUDE_HOME/settings.json" 2>/dev/null
cp /ccs-profile/settings.local.json "$CLAUDE_HOME/settings.local.json" 2>/dev/null

if [ -f /ccs-profile/mcp.json ]; then
  cp /ccs-profile/mcp.json "$(pwd)/.mcp.json"
fi
exec claude "$@"

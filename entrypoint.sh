#!/bin/bash
CLAUDE_HOME="${HOME}/.claude"
mkdir -p "$CLAUDE_HOME"

if [ -f /ccs-profile/settings.json ]; then
  cp /ccs-profile/settings.json "$CLAUDE_HOME/settings.json"
fi
if [ -f /ccs-profile/settings.local.json ]; then
  cp /ccs-profile/settings.local.json "$CLAUDE_HOME/settings.local.json"
fi
if [ -f /ccs-profile/mcp.json ]; then
  cp /ccs-profile/mcp.json "$(pwd)/.mcp.json"
fi
exec claude "$@"

#!/bin/bash
CLAUDE_HOME="${HOME}/.claude"
mkdir -p "$CLAUDE_HOME"

cp /ccs-profile/settings.json "$CLAUDE_HOME/settings.json" 2>/dev/null
cp /ccs-profile/settings.local.json "$CLAUDE_HOME/settings.local.json" 2>/dev/null

if [ -f /ccs-profile/mcp.json ]; then
  cp /ccs-profile/mcp.json "$(pwd)/.mcp.json"
fi

# Start D-Bus session bus for gnome-keyring (credential storage backend).
export DBUS_SESSION_BUS_ADDRESS="unix:path=/tmp/dbus-session"
dbus-daemon --session --address="$DBUS_SESSION_BUS_ADDRESS" --nofork --nopidfile &
sleep 0.2

# Unlock gnome-keyring with empty password so keytar/libsecret can store credentials.
mkdir -p "${HOME}/.local/share/keyrings"
eval $(echo "" | gnome-keyring-daemon --unlock --components=secrets 2>/dev/null)
export GNOME_KEYRING_CONTROL

# Point secure storage config at the persisted auth directory.
export CLAUDE_SECURESTORAGE_CONFIG_DIR="$CLAUDE_HOME"

exec claude "$@"

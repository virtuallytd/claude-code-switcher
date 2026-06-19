#!/bin/bash
if [ -f /ccs-profile/settings.json ]; then
  cp /ccs-profile/settings.json /home/claude/.claude/settings.json
fi
if [ -f /ccs-profile/settings.local.json ]; then
  cp /ccs-profile/settings.local.json /home/claude/.claude/settings.local.json
fi
exec claude "$@"

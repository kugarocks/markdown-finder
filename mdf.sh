#!/bin/bash

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title mdf
# @raycast.mode silent

# Optional parameters:
# @raycast.icon 📋
# @raycast.argument1 { "type": "text", "placeholder": "Snippet" }
# @raycast.packageName charm

# Documentation:
# @raycast.description markdown finder
# @raycast.author kugarocks

LANG=en_US.UTF-8 \
MDF_HOME=/Users/kuga/mdf \
/Applications/Alacritty.app/Contents/MacOS/alacritty \
    --config-file /Users/kuga/alacritty.toml \
    -e /usr/local/bin/mdf "$1" \
    > /dev/null 2>&1


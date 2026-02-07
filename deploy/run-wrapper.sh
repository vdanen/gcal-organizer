#!/bin/bash
# run-wrapper.sh — Wrapper script for running gcal-organizer as a service.
# Sources environment from ~/.gcal-organizer/.env and runs the full workflow.

set -euo pipefail

# Source env file if it exists
ENV_FILE="${HOME}/.gcal-organizer/.env"
if [ -f "$ENV_FILE" ]; then
    set -a
    source "$ENV_FILE"
    set +a
fi

# Override days to look back for service mode (1 day)
export GCAL_DAYS_TO_LOOK_BACK=1

# Find the binary
BINARY="${GCAL_ORGANIZER_BIN:-$(command -v gcal-organizer 2>/dev/null || echo "${HOME}/go/bin/gcal-organizer")}"

if [ ! -x "$BINARY" ]; then
    echo "ERROR: gcal-organizer binary not found at $BINARY" >&2
    echo "Set GCAL_ORGANIZER_BIN or install to PATH" >&2
    exit 1
fi

echo "$(date '+%Y-%m-%d %H:%M:%S') — Starting gcal-organizer run"
"$BINARY" run --verbose
echo "$(date '+%Y-%m-%d %H:%M:%S') — Completed gcal-organizer run"

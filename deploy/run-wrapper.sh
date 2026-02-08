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
if [ -n "${GCAL_ORGANIZER_BIN:-}" ] && [ -x "$GCAL_ORGANIZER_BIN" ]; then
    BINARY="$GCAL_ORGANIZER_BIN"
elif command -v gcal-organizer &>/dev/null; then
    BINARY="$(command -v gcal-organizer)"
else
    GOPATH_BIN="$(go env GOPATH 2>/dev/null)/bin/gcal-organizer"
    if [ -x "$GOPATH_BIN" ]; then
        BINARY="$GOPATH_BIN"
    else
        echo "ERROR: gcal-organizer binary not found" >&2
        echo "Run 'make install' first, or set GCAL_ORGANIZER_BIN" >&2
        exit 1
    fi
fi

echo "$(date '+%Y-%m-%d %H:%M:%S') — Starting gcal-organizer run"
"$BINARY" run #--verbose
echo "$(date '+%Y-%m-%d %H:%M:%S') — Completed gcal-organizer run"

#!/bin/sh
# Runs VsChat as PUID:PGID (default 10001). When started as root, first make the
# data directory writable for that user so bind mounts like ./data:/data work.
set -e

PUID="${PUID:-10001}"
PGID="${PGID:-10001}"
DATA="${VSCHAT_DATA:-/data}"

if [ "$(id -u)" = "0" ]; then
	mkdir -p "$DATA"
	if [ "$(stat -c %u:%g "$DATA")" != "$PUID:$PGID" ]; then
		chown -R "$PUID:$PGID" "$DATA"
	fi
	exec su-exec "$PUID:$PGID" vschat "$@"
fi
exec vschat "$@"

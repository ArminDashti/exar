#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CREDS_FILE="$SCRIPT_DIR/creds.toml"

CONTAINER_NAME="${CONTAINER_NAME:-netchecker-server}"
REMOTE_CONFIG_DIR="${REMOTE_CONFIG_DIR:-/opt/netchecker}"

log() { printf '[remove-server] %s\n' "$*"; }
err() { printf '[remove-server] ERROR: %s\n' "$*" >&2; }

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    err "missing required command: $1"
    exit 1
  }
}

trim() {
  local value="$1"
  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  printf '%s' "$value"
}

parse_toml() {
  local section="$1"
  local key="$2"
  awk -F= -v section="$section" -v key="$key" '
    function trim(s) {
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", s)
      return s
    }
    $0 ~ /^[[:space:]]*#/ || $0 ~ /^[[:space:]]*$/ { next }
    /^\[/ {
      current = $0
      gsub(/[\[\]]/, "", current)
      current = trim(current)
      next
    }
    current == section {
      split($0, parts, "=")
      k = trim(parts[1])
      if (k != key) {
        next
      }
      value = substr($0, index($0, "=") + 1)
      value = trim(value)
      gsub(/^"/, "", value)
      gsub(/"$/, "", value)
      print value
      exit
    }
  ' "$CREDS_FILE"
}

if [[ ! -f "$CREDS_FILE" ]]; then
  err "credentials file not found: $CREDS_FILE"
  exit 1
fi

need_cmd ssh
need_cmd sshpass

SSH_HOST_RAW="$(parse_toml server address)"
SSH_USER="$(parse_toml server username)"
SSH_PASSWORD="$(parse_toml server password)"

if [[ -z "$SSH_HOST_RAW" || -z "$SSH_USER" || -z "$SSH_PASSWORD" ]]; then
  err "creds.toml must include [server] address/username/password"
  exit 1
fi

SSH_HOST="$(trim "$SSH_HOST_RAW")"
while [[ "$SSH_HOST" == .* ]]; do
  SSH_HOST="${SSH_HOST#.}"
done

SSH_PORT=22
if [[ "$SSH_HOST" == *:* ]]; then
  SSH_PORT="${SSH_HOST##*:}"
  SSH_HOST="${SSH_HOST%:*}"
fi

if [[ -z "$SSH_HOST" ]]; then
  err "invalid server.address in creds.toml"
  exit 1
fi

log "Removing deployment on $SSH_USER@$SSH_HOST:$SSH_PORT ..."
sshpass -p "$SSH_PASSWORD" ssh -o StrictHostKeyChecking=no -p "$SSH_PORT" "$SSH_USER@$SSH_HOST" "bash -s" <<EOF_REMOTE
set -euo pipefail

SUDO_PASS="$SSH_PASSWORD"
CONTAINER_NAME="$CONTAINER_NAME"
REMOTE_CONFIG_DIR="$REMOTE_CONFIG_DIR"

run_root() {
  if [[ "\${EUID:-\$(id -u)}" -eq 0 ]]; then
    "\$@"
    return
  fi
  if ! command -v sudo >/dev/null 2>&1; then
    echo "[remote] ERROR: sudo is required for removal commands" >&2
    exit 1
  fi
  printf '%s\n' "\$SUDO_PASS" | sudo -S -p '' "\$@"
}

if command -v docker >/dev/null 2>&1; then
  if run_root docker ps -a --format '{{.Names}}' | grep -Fxq "\$CONTAINER_NAME"; then
    echo "[remote] Removing container \$CONTAINER_NAME ..."
    run_root docker rm -f "\$CONTAINER_NAME" >/dev/null
  else
    echo "[remote] Container \$CONTAINER_NAME does not exist."
  fi
else
  echo "[remote] Docker not found. Skipping container removal."
fi

if [[ "${PURGE_CONFIG:-0}" == "1" ]]; then
  if run_root test -d "\$REMOTE_CONFIG_DIR"; then
    echo "[remote] Deleting config directory \$REMOTE_CONFIG_DIR ..."
    run_root rm -rf "\$REMOTE_CONFIG_DIR"
  else
    echo "[remote] Config directory \$REMOTE_CONFIG_DIR does not exist."
  fi
else
  echo "[remote] Keeping \$REMOTE_CONFIG_DIR (set PURGE_CONFIG=1 to delete it)."
fi
EOF_REMOTE

log "Done."

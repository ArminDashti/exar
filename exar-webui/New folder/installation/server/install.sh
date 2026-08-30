#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
CREDS_FILE="$SCRIPT_DIR/creds.toml"

IMAGE_NAME="${IMAGE_NAME:-netchecker-server:latest}"
CONTAINER_NAME="${CONTAINER_NAME:-netchecker-server}"
REMOTE_CONFIG_DIR="${REMOTE_CONFIG_DIR:-/opt/netchecker}"
REMOTE_CONFIG_FILE="$REMOTE_CONFIG_DIR/configs.toml"
APP_PORT="${APP_PORT:-8080}"
SUDO=""

log() { printf '[install] %s\n' "$*"; }
err() { printf '[install] ERROR: %s\n' "$*" >&2; }

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    err "missing required command: $1"
    exit 1
  }
}

init_local_sudo() {
  if docker info >/dev/null 2>&1; then
    return
  fi
  if [[ "${EUID:-$(id -u)}" -eq 0 ]]; then
    return
  fi
  if ! command -v sudo >/dev/null 2>&1; then
    err "docker requires elevated privileges, but sudo is not installed"
    exit 1
  fi
  if sudo docker info >/dev/null 2>&1; then
    SUDO="sudo"
    return
  fi
  err "docker is not accessible (with or without sudo) on this host"
  exit 1
}

run_local_docker() {
  if [[ -n "$SUDO" ]]; then
    $SUDO docker "$@"
  else
    docker "$@"
  fi
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

need_cmd docker
need_cmd ssh
need_cmd scp
need_cmd sshpass
init_local_sudo

SSH_HOST_RAW="$(parse_toml server address)"
SSH_USER="$(parse_toml server username)"
SSH_PASSWORD="$(parse_toml server password)"
DB_SERVER="$(parse_toml database server)"
DB_NAME="$(parse_toml database database)"
DB_USER="$(parse_toml database username)"
DB_PASSWORD="$(parse_toml database password)"

if [[ -z "$SSH_HOST_RAW" || -z "$SSH_USER" || -z "$SSH_PASSWORD" ]]; then
  err "creds.toml must include [server] address/username/password"
  exit 1
fi
if [[ -z "$DB_SERVER" || -z "$DB_NAME" || -z "$DB_USER" || -z "$DB_PASSWORD" ]]; then
  err "creds.toml must include [database] server/database/username/password"
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

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

LOCAL_IMAGE_TAR="$TMP_DIR/image.tar"
LOCAL_RUNTIME_CONFIG="$TMP_DIR/configs.toml"
REMOTE_BUNDLE_DIR="/tmp/netchecker-deploy-$$"
REMOTE_IMAGE_TAR="$REMOTE_BUNDLE_DIR/image.tar"
REMOTE_RUNTIME_CONFIG="$REMOTE_BUNDLE_DIR/configs.toml"

cat >"$LOCAL_RUNTIME_CONFIG" <<EOF
[database]
server-address = "$DB_SERVER"
db-name = "$DB_NAME"
username = "$DB_USER"
password = "$DB_PASSWORD"

[api]
address = "0.0.0.0"
port = 8080
EOF

log "Building Docker image $IMAGE_NAME from project source..."
run_local_docker build \
  -t "$IMAGE_NAME" \
  -f - \
  "$PROJECT_ROOT" <<'DOCKERFILE'
FROM golang:1.22-alpine AS build
WORKDIR /src
ENV GOPROXY=https://package-mirror.liara.ir/repository/go/,direct
COPY server/deb-amd64/go.mod server/deb-amd64/go.sum ./
RUN go mod download
COPY server/deb-amd64/. .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/ncs ./cmd/ncs

FROM alpine:3.20
RUN ALPINE_VERSION="v$(cut -d. -f1,2 /etc/alpine-release)" \
 && printf 'https://mirror.arvancloud.ir/alpine/%s/main\nhttps://mirror.arvancloud.ir/alpine/%s/community\n' "$ALPINE_VERSION" "$ALPINE_VERSION" > /etc/apk/repositories \
 && apk add --no-cache ca-certificates tzdata
COPY --from=build /out/ncs /usr/local/bin/ncs
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/ncs", "server"]
DOCKERFILE

log "Exporting image to tar archive..."
run_local_docker save -o "$LOCAL_IMAGE_TAR" "$IMAGE_NAME"

log "Connecting to $SSH_USER@$SSH_HOST:$SSH_PORT and uploading image..."
sshpass -p "$SSH_PASSWORD" ssh -o StrictHostKeyChecking=no -p "$SSH_PORT" "$SSH_USER@$SSH_HOST" \
  "mkdir -p '$REMOTE_BUNDLE_DIR'"
sshpass -p "$SSH_PASSWORD" scp -P "$SSH_PORT" -o StrictHostKeyChecking=no "$LOCAL_IMAGE_TAR" "$SSH_USER@$SSH_HOST:$REMOTE_IMAGE_TAR"
sshpass -p "$SSH_PASSWORD" scp -P "$SSH_PORT" -o StrictHostKeyChecking=no "$LOCAL_RUNTIME_CONFIG" "$SSH_USER@$SSH_HOST:$REMOTE_RUNTIME_CONFIG"

log "Deploying container on remote host..."
sshpass -p "$SSH_PASSWORD" ssh -o StrictHostKeyChecking=no -p "$SSH_PORT" "$SSH_USER@$SSH_HOST" "bash -s" <<EOF
set -euo pipefail

SUDO_PASS="$SSH_PASSWORD"

run_root() {
  if [[ "\${EUID:-\$(id -u)}" -eq 0 ]]; then
    "\$@"
    return
  fi
  if ! command -v sudo >/dev/null 2>&1; then
    echo "[remote] ERROR: sudo is required for deployment commands" >&2
    exit 1
  fi
  printf '%s\n' "\$SUDO_PASS" | sudo -S -p '' "\$@"
}

if ! command -v docker >/dev/null 2>&1; then
  echo "[remote] ERROR: docker is not installed on remote server" >&2
  exit 1
fi

run_root docker load -i "$REMOTE_IMAGE_TAR" >/dev/null

if run_root docker ps -a --format '{{.Names}}' | grep -Fxq "$CONTAINER_NAME"; then
  echo "[remote] Container $CONTAINER_NAME already exists. Replacing with updated image..."
  run_root docker rm -f "$CONTAINER_NAME" >/dev/null
else
  echo "[remote] Container $CONTAINER_NAME does not exist. Installing new container..."
fi

run_root mkdir -p "$REMOTE_CONFIG_DIR"
run_root mv "$REMOTE_RUNTIME_CONFIG" "$REMOTE_CONFIG_FILE"
run_root chmod 600 "$REMOTE_CONFIG_FILE"

run_root docker run -d \
  --name "$CONTAINER_NAME" \
  --restart unless-stopped \
  -p "$APP_PORT:8080" \
  -v "$REMOTE_CONFIG_DIR:/home/.netchecker-server:ro" \
  "$IMAGE_NAME" >/dev/null

run_root rm -rf "$REMOTE_BUNDLE_DIR"
echo "[remote] Deployment finished."
EOF

log "Done. Running container: $CONTAINER_NAME"

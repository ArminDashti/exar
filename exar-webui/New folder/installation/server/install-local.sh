#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
CREDS_FILE="$SCRIPT_DIR/creds.toml"

IMAGE_NAME="${IMAGE_NAME:-netchecker-server:latest}"
CONTAINER_NAME="${CONTAINER_NAME:-netchecker-server}"
LOCAL_CONFIG_DIR="${LOCAL_CONFIG_DIR:-/opt/netchecker}"
LOCAL_CONFIG_FILE="$LOCAL_CONFIG_DIR/configs.toml"
APP_PORT="${APP_PORT:-8080}"
SUDO=""

log() { printf '[install-local] %s\n' "$*"; }
err() { printf '[install-local] ERROR: %s\n' "$*" >&2; }

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

run_local() {
  if [[ -n "$SUDO" ]]; then
    $SUDO "$@"
  else
    "$@"
  fi
}

if [[ ! -f "$CREDS_FILE" ]]; then
  err "credentials file not found: $CREDS_FILE"
  exit 1
fi

need_cmd docker
init_local_sudo

DB_SERVER="$(parse_toml database server)"
DB_NAME="$(parse_toml database database)"
DB_USER="$(parse_toml database username)"
DB_PASSWORD="$(parse_toml database password)"

if [[ -z "$DB_SERVER" || -z "$DB_NAME" || -z "$DB_USER" || -z "$DB_PASSWORD" ]]; then
  err "creds.toml must include [database] server/database/username/password"
  exit 1
fi

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

LOCAL_RUNTIME_CONFIG="$TMP_DIR/configs.toml"

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
run_local docker build \
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

log "Preparing runtime config in $LOCAL_CONFIG_DIR ..."
if ! run_local mkdir -p "$LOCAL_CONFIG_DIR" >/dev/null 2>&1; then
  err "cannot create $LOCAL_CONFIG_DIR (try running as root or set LOCAL_CONFIG_DIR)"
  exit 1
fi
run_local cp "$LOCAL_RUNTIME_CONFIG" "$LOCAL_CONFIG_FILE"
run_local chmod 600 "$LOCAL_CONFIG_FILE"

if run_local docker ps -a --format '{{.Names}}' | grep -Fxq "$CONTAINER_NAME"; then
  log "Container $CONTAINER_NAME already exists. Replacing with updated image..."
  run_local docker rm -f "$CONTAINER_NAME" >/dev/null
else
  log "Container $CONTAINER_NAME does not exist. Installing new container..."
fi

run_local docker run -d \
  --name "$CONTAINER_NAME" \
  --restart unless-stopped \
  -p "$APP_PORT:8080" \
  -v "$LOCAL_CONFIG_DIR:/home/.netchecker-server:ro" \
  "$IMAGE_NAME" >/dev/null

log "Done. Running container: $CONTAINER_NAME"

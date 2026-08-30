#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

IMAGE_NAME="${IMAGE_NAME:-exar-web:latest}"
CONTAINER_NAME="${CONTAINER_NAME:-exar-web}"
APP_PORT="${APP_PORT:-8080}"
DATA_DIR="${DATA_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/exar-web/data}"
DOCKERFILE="$PROJECT_ROOT/dockerfile"

log() { printf '[docker-install] %s\n' "$*"; }
err() { printf '[docker-install] ERROR: %s\n' "$*" >&2; }

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    err "missing required command: $1"
    exit 1
  }
}

need_cmd docker

if ! docker info >/dev/null 2>&1; then
  err "docker daemon is not running or not accessible"
  exit 1
fi

if [[ ! -f "$DOCKERFILE" ]]; then
  err "dockerfile not found: $DOCKERFILE"
  exit 1
fi

mkdir -p "$DATA_DIR"

log "Building Docker image $IMAGE_NAME (multi-stage) ..."
docker build -f "$DOCKERFILE" -t "$IMAGE_NAME" "$PROJECT_ROOT"

if docker ps -a --format '{{.Names}}' | grep -Fxq "$CONTAINER_NAME"; then
  log "Container $CONTAINER_NAME exists. Replacing with updated image ..."
  docker rm -f "$CONTAINER_NAME" >/dev/null
else
  log "Container $CONTAINER_NAME does not exist. Installing new container ..."
fi

docker run -d \
  --name "$CONTAINER_NAME" \
  --restart unless-stopped \
  -p "$APP_PORT:8080" \
  -v "$DATA_DIR:/app/data" \
  "$IMAGE_NAME" >/dev/null

log "Done. Running container: $CONTAINER_NAME on http://localhost:$APP_PORT"

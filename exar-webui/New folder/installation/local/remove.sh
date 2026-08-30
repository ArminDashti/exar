#!/usr/bin/env bash
set -euo pipefail

CONTAINER_NAME="${CONTAINER_NAME:-netchecker-server}"
LOCAL_CONFIG_DIR="${LOCAL_CONFIG_DIR:-/opt/netchecker}"
SUDO=""

log() { printf '[remove-local] %s\n' "$*"; }
err() { printf '[remove-local] ERROR: %s\n' "$*" >&2; }

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

run_local() {
  if [[ -n "$SUDO" ]]; then
    $SUDO "$@"
  else
    "$@"
  fi
}

need_cmd docker
init_local_sudo

if run_local docker ps -a --format '{{.Names}}' | grep -Fxq "$CONTAINER_NAME"; then
  log "Removing container $CONTAINER_NAME ..."
  run_local docker rm -f "$CONTAINER_NAME" >/dev/null
else
  log "Container $CONTAINER_NAME does not exist."
fi

if [[ "${PURGE_CONFIG:-0}" == "1" ]]; then
  if run_local test -d "$LOCAL_CONFIG_DIR"; then
    log "Deleting config directory $LOCAL_CONFIG_DIR ..."
    run_local rm -rf "$LOCAL_CONFIG_DIR"
  else
    log "Config directory $LOCAL_CONFIG_DIR does not exist."
  fi
else
  log "Keeping $LOCAL_CONFIG_DIR (set PURGE_CONFIG=1 to delete it)."
fi

log "Done."

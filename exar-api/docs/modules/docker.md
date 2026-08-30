# Docker deployment

Standalone API stack for `exar-api`.

## Files

| File | Purpose |
|------|---------|
| `dockerfile` | Builds the Go API binary |
| `docker-compose.yml` | `postgres` + `api` on external network `exar-net` (expose only) |
| `docker-compose.publish.yml` | Optional host port overlay (loaded when `publish_port` is set) |
| `.armin/docker-scripts/run-on-docker-local.ps1` | Local Docker daemon deploy (YAML-only) |
| `.armin/docker-scripts/run-on-docker-local.yaml` | Local stack settings |
| `.armin/docker-scripts/run-on-docker-server.ps1` | Remote SSH deploy (YAML-only) |
| `.armin/docker-scripts/run-on-docker-server.yaml` | Remote settings (`ssh` / `volume_dir`) |
| `.docker/stack.manifest.json` | Image tag, container name, ports |

## Services

| Service | Container | Host port | Notes |
|---------|-----------|-----------|-------|
| `postgres` | `exar-postgres` | — | PostgreSQL 16; volume `exar-postgres-data` |
| `api` | `exar` | optional (`publish_port`) | Connects via `DATABASE_URL` to `postgres` |

Requires external Docker network `exar-net` (scripts create it if missing).

## Deploy

```powershell
# Local
.\.armin\docker-scripts\run-on-docker-local.ps1

# Remote (fill ssh + volume_dir in YAML first)
.\.armin\docker-scripts\run-on-docker-server.ps1
```

Local YAML publishes the API on host port `8086` (container `:8080`) via `docker-compose.publish.yml`. Empty `publish_port` skips that overlay so the service stays on `exar-net` only (proxy-friendly).

Compose override env vars used by the scripts: `IMAGE_TAG`, `DOCKER_NETWORK`, `INTERNAL_PORT`, `PUBLISH_PORT` (publish overlay only).

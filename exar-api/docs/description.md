# exar-api

Go REST API for the shared expense tracker. Users authenticate with JWT, manage shops and item names, and record expenses with Persian calendar dates stored as Gregorian in PostgreSQL.

## Tech stack

- Go 1.22, Gin, pgx
- PostgreSQL 16
- Docker Compose via `.armin/docker-scripts/` (local + SSH deploy)

## Run locally

```bash
go mod tidy
go run ./cmd/server
```

Set `DATABASE_URL` if PostgreSQL is not at the default connection string.

## Run with Docker

```powershell
.\.armin\docker-scripts\run-on-docker-local.ps1
```

API on Docker network `exar-net` (`exar:8080`). Prefer [exar-web](../exar-web) as the HTTP front door.

## Related repo

Frontend lives in [exar-web](../exar-web). The web container proxies `/api/*` to this API on the shared Docker network `exar-net`.

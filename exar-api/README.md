# Daily Expenses API

REST API for tracking shared spending between two people. Built with **Go (Gin)** and **PostgreSQL**.

## Features

- Track expenses by shop and date
- JWT authentication for two hardcoded users (`armin`, `ramin`)
- Items with shop name, amount, and Persian calendar date (converted to Gregorian for storage)
- REST API with filtering by shop and date range
- PostgreSQL persistence with automatic schema migration
- Docker deployment with PostgreSQL service and named volume

## Project layout

```
cmd/server/       HTTP server entry point
internal/         handlers, database, models, jalali date conversion
docs/             API reference and OpenAPI spec
```

## Run locally

### Prerequisites

- Go 1.22+
- PostgreSQL 16+ (or use Docker Compose in this repo)

### Start the server

```bash
go mod tidy
go run ./cmd/server
```

The API listens on `http://localhost:8080`. Set `DATABASE_URL` if PostgreSQL is not at the default `postgres://exar:exar@localhost:5432/exar?sslmode=disable`.

## Run with Docker

Build and run:

```bash
docker build -t exar-api:latest .
docker run -p 8080:8080 exar-api:latest
```

Or with Compose (includes PostgreSQL):

```bash
docker compose up --build
```

API base URL: `http://localhost:8080/exar/api/v1`

## API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/exar/api/v1/auth/login` | No | Login `{ "username", "password" }` → JWT |
| GET | `/exar/api/v1/persons` | Yes | List persons (IDs 1 and 2) |
| GET | `/exar/api/v1/shops` | Yes | List shops |
| POST | `/exar/api/v1/shop` | Yes | Create shop `{ "name": "..." }` |
| DELETE | `/exar/api/v1/shop/:id` | Yes | Delete unused shop |
| GET | `/exar/api/v1/items` | Yes | List items (query: `shop`, `from_date`, `to_date`) |
| GET | `/exar/api/v1/item/:id` | Yes | Get one item |
| POST | `/exar/api/v1/item` | Yes | Create item with shop, amount, Persian date |
| DELETE | `/exar/api/v1/item/:id` | Yes | Delete item |

Send `Authorization: Bearer <token>` on all endpoints except login.

### Login example

```json
{ "username": "armin", "password": "Kp9#mX2vQwL4nT7" }
```

### Create item example

```json
{
  "shop": "Grocery Store",
  "amount": 24.50,
  "date": "1405-06-10"
}
```

The `date` field uses the Persian (Jalali) calendar (`YYYY-MM-DD`). The server converts it to Gregorian (e.g. `1405-06-10` → `2026-09-01`) before storing.

See [docs/endpoints.md](docs/endpoints.md) and [docs/openapi.yaml](docs/openapi.yaml) for full API documentation.

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ADDR` | `:8080` | HTTP listen address |
| `DATABASE_URL` | `postgres://exar:exar@localhost:5432/exar?sslmode=disable` | PostgreSQL connection string |

# API Endpoints

Base URL: `http://<host>:8080/exar/api/v1` (default port `8080`, configurable via `ADDR`)

| Method | Path | Full URL (local) | Auth | Description |
|--------|------|------------------|------|-------------|
| POST | `/auth/login` | `http://localhost:8080/exar/api/v1/auth/login` | No | Login and receive JWT token |
| GET | `/persons` | `http://localhost:8080/exar/api/v1/persons` | Yes | List persons (IDs 1 and 2) |
| GET | `/shops` | `http://localhost:8080/exar/api/v1/shops` | Yes | List shops |
| POST | `/shop` | `http://localhost:8080/exar/api/v1/shop` | Yes | Create shop |
| DELETE | `/shop/:id` | `http://localhost:8080/exar/api/v1/shop/{id}` | Yes | Delete shop by ID |
| GET | `/products` | `http://localhost:8080/exar/api/v1/products` | Yes | List product/item names |
| POST | `/product` | `http://localhost:8080/exar/api/v1/product` | Yes | Create product/item name |
| DELETE | `/product/:id` | `http://localhost:8080/exar/api/v1/product/{id}` | Yes | Delete product by ID |
| GET | `/items` | `http://localhost:8080/exar/api/v1/items` | Yes | List expenses |
| GET | `/item/:id` | `http://localhost:8080/exar/api/v1/item/{id}` | Yes | Get expense by ID |
| POST | `/item` | `http://localhost:8080/exar/api/v1/item` | Yes | Create expense |
| DELETE | `/item/:id` | `http://localhost:8080/exar/api/v1/item/{id}` | Yes | Delete expense by ID |

## Authentication

Send `Authorization: Bearer <token>` on all endpoints except `POST /auth/login`.

Hardcoded users:

| Username | Password | Person ID |
|----------|----------|-----------|
| `armin` | `Kp9#mX2vQwL4nT7` | 1 |
| `ramin` | `Hn7$rT5yBcF8wJ3` | 2 |

## Query parameters

### `GET /items`

| Parameter | Type | Description |
|-----------|------|-------------|
| `shop` | string | Filter by exact shop name |
| `person_id` | int | Filter by user/person ID |
| `from_date` | string | Include items on or after this date (`YYYY-MM-DD`, Gregorian) |
| `to_date` | string | Include items on or before this date (`YYYY-MM-DD`, Gregorian) |

Example:

`http://localhost:8080/exar/api/v1/items?shop=Grocery%20Store&from_date=2026-06-01&to_date=2026-06-30`

## Notes

- All request and response bodies use `application/json`.
- Item `date` on create uses Persian calendar (`YYYY-MM-DD`); responses return the stored Gregorian date.
- Unknown routes return `404` with `{"error": "not found"}`.

For request/response schemas and examples, see [docs/endpoints.md](docs/endpoints.md) and [docs/openapi.yaml](docs/openapi.yaml).

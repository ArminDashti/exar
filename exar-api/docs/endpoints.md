# API endpoints

Base path: `/exar/api/v1` (see `api/endpoints.md` for the full reference).

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/exar/api/v1/auth/login` | Login, returns JWT | No |
| GET | `/exar/api/v1/persons` | List persons | Yes |
| GET | `/exar/api/v1/shops` | List shops | Yes |
| POST | `/exar/api/v1/shop` | Create shop | Yes |
| DELETE | `/exar/api/v1/shop/:id` | Delete shop | Yes |
| GET | `/exar/api/v1/products` | List product/item names | Yes |
| POST | `/exar/api/v1/product` | Create product/item name | Yes |
| DELETE | `/exar/api/v1/product/:id` | Delete product | Yes |
| GET | `/exar/api/v1/items` | List expenses | Yes |
| GET | `/exar/api/v1/item/:id` | Get expense | Yes |
| POST | `/exar/api/v1/item` | Create expense | Yes |
| DELETE | `/exar/api/v1/item/:id` | Delete expense | Yes |

The web UI calls `/api/*`; nginx proxies those requests to the API container.

## Expense create body (`POST /item`)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `shop` | string | yes | Shop name (created if new) |
| `product` | string | yes | Item name (must exist in products, e.g. Milk) |
| `person_id` | int | yes | User who made the expense |
| `amount` | number | yes | Amount spent |
| `date` | string | yes | Persian date `YYYY-MM-DD` |

## Expense list filters (`GET /items`)

| Parameter | Description |
|-----------|-------------|
| `shop` | Filter by shop name |
| `person_id` | Filter by user ID |
| `from_date` | Gregorian `YYYY-MM-DD` |
| `to_date` | Gregorian `YYYY-MM-DD` |

# API endpoints

See [../endpoints.md](../endpoints.md) for the full REST reference.

Summary:

- `GET /api/persons`
- `GET|POST /api/shops`, `PUT|DELETE /api/shops/:id` (`GET` supports `?q=`)
- `GET|POST /api/items`, `PUT|DELETE /api/items/:id` (`GET` supports `?q=`; delete blocked if in use)
- `GET /api/stats` → `{ by_month: [...] }`
- `GET /api/expenses/check-duplicate` → soft same-item+same-date warning helper
- `GET|POST /api/expenses`, `GET|PUT|DELETE /api/expenses/:id` (amounts must be whole numbers; expenses link via `item_id`)

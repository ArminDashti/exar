## API

- On every server start, a new API token is generated and saved to `configs.toml` (default path in current working directory).
- Send that token in `X-API-Token` header (or `?token=<TOKEN>` query) for all `/expenses/api/v1/*` requests.
- `GET /expenses/api/v1/calculation?from-date=<FIRST-DAY-OF-MONTHS>&to-date=<TODAY>`
- `GET /expenses/api/v1/expenses-list?owner=1&store=all&from-date=<FIRST-DAY-OF-MONTHS>&to-date=<TODAY>`
- `POST /expenses/api/v1/add-expense` {"store":"1",  "store":"<REQUIRED>", "purchase":<REQUIRED>, "purchase-date":"<REQUIRED>"}
- `DELETE /expenses/api/v1/delete-expense?purchase=<REQUIRED>&purchase-date=<REQUIRED>`

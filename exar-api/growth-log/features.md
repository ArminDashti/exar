# Features

- JWT login for users `armin` / `ramin`
- CRUD-style shops, products, and expense items
- Persian (Jalali) date input converted to Gregorian for storage
- PostgreSQL persistence with auto-migration
- Docker Compose stack: `exar` API + `exar-postgres` on `exar-net`
- YAML-driven deploy under `.armin/docker-scripts/` (local + remote SSH)

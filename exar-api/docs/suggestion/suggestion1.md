# Align empty list JSON responses

`GET /shops` returned JSON `null` while `GET /items` returned `[]` on a fresh database.

Prefer empty arrays for list endpoints so clients can iterate without null checks.

**Effort:** low (initialize slices in handlers / repository).

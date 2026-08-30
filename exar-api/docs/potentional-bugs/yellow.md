# Yellow — minor risks

- **[API lists]** — `GET /shops` can serialize as JSON `null` when there are no rows, while `GET /items` returns `[]`. Frontend code that assumes arrays may throw.
- **[Compose publish]** — Empty `API_PUBLISH_PORT` may still publish a random host port instead of omitting the mapping.

# Clarify empty API_PUBLISH_PORT in compose

`docker-compose.yml` uses `published: ${API_PUBLISH_PORT-8080}`. An empty `API_PUBLISH_PORT` does not omit the port mapping — Compose may assign a random host port instead of leaving the service publish-free.

Prefer an explicit host port when needed, or change compose so empty truly means “expose only” (no `ports:` entry).

**Effort:** low–medium.

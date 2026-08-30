# Architecture (technical)

Go Gin API container (`exar`) and PostgreSQL 16 (`exar-postgres`) on external Docker network `exar-net`. Compose publishes API host port 8080. Image tag comes from `.docker/stack.manifest.json` (`apiImageTag`). Deploy via `run-on-docker.ps1` (`docker compose up -d --build`). Auth is JWT; data via `DATABASE_URL`.

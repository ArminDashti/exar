# Suggestions

- Align empty list responses to `[]` across list endpoints.
- Retire or thin root `run-on-docker.ps1` now that `.armin/docker-scripts/` is the deploy entry point.
- Fix compose so empty `API_PUBLISH_PORT` means no host bind (not a random port).

# Bugs

- Fixed: `run-on-docker.ps1` crashed under StrictMode when reading missing `imageTag` from `.docker/stack.manifest.json` (manifest uses `apiImageTag`).
- Open: empty shops list may return JSON `null` instead of `[]`.

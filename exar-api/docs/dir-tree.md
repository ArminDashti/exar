# Directory tree

```
exar-api/
├── db.md                      # PostgreSQL schema reference
├── dockerfile                 # Go API image
├── docker-compose.yml         # postgres + API stack
├── .armin/
│   └── docker-scripts/
│       ├── run-on-docker-local.ps1    # Local Docker deploy
│       ├── run-on-docker-local.yaml   # Local deploy settings
│       ├── run-on-docker-server.ps1   # Remote SSH deploy
│       └── run-on-docker-server.yaml  # Remote deploy settings
├── .docker/
│   └── stack.manifest.json    # imageTag, container names, ports
├── docs/
│   ├── description.md         # Project overview
│   ├── endpoints.md           # REST endpoint reference
│   ├── dir-tree.md            # This file
│   ├── modules/
│   │   └── docker.md          # Docker deployment notes
│   ├── suggestion/
│   │   └── suggestion1.md     # Empty-list JSON consistency idea
│   └── potentional-bugs/
│       ├── red.md             # Critical risks
│       └── yellow.md          # Minor risks
├── growth-log/                # Agent growth documentation
├── cmd/server/main.go         # HTTP server entry point
└── internal/                  # handlers, auth, database, models, jalali
```

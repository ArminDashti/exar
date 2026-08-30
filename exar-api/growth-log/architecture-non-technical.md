# Architecture (non-technical)

The app runs as two Docker containers: one for the API and one for the database. They talk to each other on a private Docker network. You reach the API at `http://localhost:8080`.

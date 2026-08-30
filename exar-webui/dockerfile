# syntax=docker/dockerfile:1

# --- frontend ---
FROM node:22-alpine AS frontend
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY index.html vite.config.js postcss.config.js tailwind.config.js ./
COPY public ./public
COPY src ./src
RUN npm run build

# --- backend ---
FROM golang:1.22-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# --- runtime ---
FROM alpine:3.20
RUN ALPINE_VERSION="v$(cut -d. -f1,2 /etc/alpine-release)" \
 && printf 'https://mirror.arvancloud.ir/alpine/%s/main\nhttps://mirror.arvancloud.ir/alpine/%s/community\n' "$ALPINE_VERSION" "$ALPINE_VERSION" > /etc/apk/repositories \
 && apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /out/server /usr/local/bin/server
COPY --from=frontend /src/dist /app/static
RUN mkdir -p /app/data
ENV DATABASE_PATH=/app/data/expenses.db
ENV STATIC_DIR=/app/static
ENV ADDR=:8080
EXPOSE 8080
VOLUME ["/app/data"]
ENTRYPOINT ["/usr/local/bin/server"]

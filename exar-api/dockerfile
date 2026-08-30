FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /server /app/server
ENV DATABASE_URL=postgres://exar:exar@localhost:5432/exar?sslmode=disable
ENV ADDR=:8080
EXPOSE 8080
CMD ["/app/server"]

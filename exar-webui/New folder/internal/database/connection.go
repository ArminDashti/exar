package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var (
	poolOnce sync.Once
	poolErr  error
	db       *sql.DB
)

func dbConn() (*sql.DB, error) {
	poolOnce.Do(func() {
		dsn, err := connectionString()
		if err != nil {
			poolErr = err
			return
		}
		conn, err := sql.Open("postgres", dsn)
		if err != nil {
			poolErr = fmt.Errorf("open database: %w", err)
			return
		}
		if err := conn.PingContext(context.Background()); err != nil {
			_ = conn.Close()
			poolErr = fmt.Errorf("ping database: %w", err)
			return
		}
		db = conn
	})
	return db, poolErr
}

func connectionString() (string, error) {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url, nil
	}

	host := envOr("PGHOST", "localhost")
	port := envOr("PGPORT", "5432")
	name := envOr("PGDATABASE", "")
	user := envOr("PGUSER", "")
	password := os.Getenv("PGPASSWORD")

	if name == "" || user == "" {
		return "", fmt.Errorf("set DATABASE_URL or PGHOST/PGPORT/PGDATABASE/PGUSER/PGPASSWORD")
	}

	if password != "" {
		return fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			user, password, host, port, name,
		), nil
	}
	return fmt.Sprintf(
		"postgres://%s@%s:%s/%s?sslmode=disable",
		user, host, port, name,
	), nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

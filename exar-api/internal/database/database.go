package database

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	*sql.DB
}

func Open(databaseURL string) (*DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(10)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &DB{DB: db}, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS persons (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS shops (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS items (
		id SERIAL PRIMARY KEY,
		shop_id INTEGER NOT NULL REFERENCES shops(id),
		product_id INTEGER REFERENCES products(id),
		person_id INTEGER REFERENCES persons(id),
		amount DOUBLE PRECISION NOT NULL,
		date TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_items_date ON items(date);
	CREATE INDEX IF NOT EXISTS idx_items_shop ON items(shop_id);
	CREATE INDEX IF NOT EXISTS idx_items_person ON items(person_id);
	CREATE INDEX IF NOT EXISTS idx_items_product ON items(product_id);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}

	return seedPersons(db)
}

func seedPersons(db *sql.DB) error {
	persons := []struct {
		id   int
		name string
	}{
		{1, "armin"},
		{2, "ramin"},
	}

	for _, p := range persons {
		_, err := db.Exec(
			`INSERT INTO persons (id, name) VALUES ($1, $2)
			 ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name`,
			p.id, p.name,
		)
		if err != nil {
			return fmt.Errorf("seed person %d: %w", p.id, err)
		}
	}

	return nil
}

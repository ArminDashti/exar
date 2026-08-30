package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/armin/expenses/backend/internal/database"
	_ "modernc.org/sqlite"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	path := os.Args[1]
	_ = os.Remove(path)
	db, err := sql.Open("sqlite", path)
	must(err)
	_, err = db.Exec(`PRAGMA foreign_keys = ON`)
	must(err)
	_, err = db.Exec(`
		CREATE TABLE persons (id INTEGER PRIMARY KEY, name TEXT NOT NULL);
		CREATE TABLE shops (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE);
		CREATE TABLE items (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE);
		CREATE TABLE expenses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			person_id INTEGER NOT NULL,
			shop_id INTEGER NOT NULL,
			date TEXT NOT NULL,
			name TEXT NOT NULL,
			amount REAL NOT NULL,
			FOREIGN KEY (person_id) REFERENCES persons(id),
			FOREIGN KEY (shop_id) REFERENCES shops(id)
		);
		CREATE TABLE expense_shares (
			expense_id INTEGER NOT NULL,
			person_id INTEGER NOT NULL,
			share REAL NOT NULL,
			PRIMARY KEY (expense_id, person_id),
			FOREIGN KEY (expense_id) REFERENCES expenses(id) ON DELETE CASCADE,
			FOREIGN KEY (person_id) REFERENCES persons(id)
		);
		INSERT INTO persons (id, name) VALUES (1, 'Armin'), (2, 'Ramin');
		INSERT INTO shops (name) VALUES ('Shop A');
		INSERT INTO items (name) VALUES ('Milk');
		INSERT INTO expenses (person_id, shop_id, date, name, amount) VALUES (1, 1, '2026-06-15', 'Milk', 100);
		INSERT INTO expenses (person_id, shop_id, date, name, amount) VALUES (2, 1, '2026-06-16', 'Bread', 50);
		INSERT INTO expense_shares VALUES (1, 1, 0.7), (1, 2, 0.3), (2, 1, 0.5), (2, 2, 0.5);
	`)
	must(err)
	must(db.Close())

	migrated, err := database.Open(path)
	must(err)
	defer migrated.Close()

	var expCount, shareCount, itemCount int
	must(migrated.QueryRow(`SELECT COUNT(*) FROM expenses`).Scan(&expCount))
	must(migrated.QueryRow(`SELECT COUNT(*) FROM expense_shares`).Scan(&shareCount))
	must(migrated.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&itemCount))

	var nameCol, itemCol int
	must(migrated.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('expenses') WHERE name='name'`).Scan(&nameCol))
	must(migrated.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('expenses') WHERE name='item_id'`).Scan(&itemCol))

	var milkName string
	must(migrated.QueryRow(`SELECT i.name FROM expenses e JOIN items i ON i.id = e.item_id WHERE e.id = 1`).Scan(&milkName))
	fmt.Printf("expenses=%d shares=%d items=%d nameCol=%d itemCol=%d milk=%s\n", expCount, shareCount, itemCount, nameCol, itemCol, milkName)

	_, err = migrated.Exec(`UPDATE items SET name = 'Milk2' WHERE name = 'Milk'`)
	must(err)
	must(migrated.QueryRow(`SELECT i.name FROM expenses e JOIN items i ON i.id = e.item_id WHERE e.id = 1`).Scan(&milkName))
	fmt.Printf("renamed=%s\n", milkName)

	if expCount != 2 || shareCount != 4 || itemCount < 2 || nameCol != 0 || itemCol != 1 || milkName != "Milk2" {
		fmt.Println("FAIL")
		os.Exit(1)
	}
	fmt.Println("OK")
}

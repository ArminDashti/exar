package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Open(path string) (*DB, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}

	if err := ensureExpensesTable(db); err != nil {
		return err
	}

	sharesSchema := `
	CREATE TABLE IF NOT EXISTS expense_shares (
		expense_id INTEGER NOT NULL,
		person_id INTEGER NOT NULL,
		share REAL NOT NULL,
		PRIMARY KEY (expense_id, person_id),
		FOREIGN KEY (expense_id) REFERENCES expenses(id) ON DELETE CASCADE,
		FOREIGN KEY (person_id) REFERENCES persons(id)
	);

	CREATE INDEX IF NOT EXISTS idx_expenses_person ON expenses(person_id);
	CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(date);
	CREATE INDEX IF NOT EXISTS idx_expenses_shop ON expenses(shop_id);
	CREATE INDEX IF NOT EXISTS idx_expenses_item ON expenses(item_id);
	`

	if _, err := db.Exec(sharesSchema); err != nil {
		return fmt.Errorf("migrate expense_shares: %w", err)
	}

	if err := seedPersons(db); err != nil {
		return err
	}

	return migrateFromInvoices(db)
}

func tableExists(db *sql.DB, name string) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'table' AND name = ?
	`, name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func columnExists(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if strings.EqualFold(name, column) {
			return true, nil
		}
	}
	return false, rows.Err()
}

func ensureExpensesTable(db *sql.DB) error {
	exists, err := tableExists(db, "expenses")
	if err != nil {
		return fmt.Errorf("check expenses table: %w", err)
	}
	if !exists {
		_, err := db.Exec(`
			CREATE TABLE expenses (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				person_id INTEGER NOT NULL,
				shop_id INTEGER NOT NULL,
				item_id INTEGER NOT NULL,
				date TEXT NOT NULL,
				amount REAL NOT NULL,
				FOREIGN KEY (person_id) REFERENCES persons(id),
				FOREIGN KEY (shop_id) REFERENCES shops(id),
				FOREIGN KEY (item_id) REFERENCES items(id)
			)
		`)
		if err != nil {
			return fmt.Errorf("create expenses: %w", err)
		}
		return nil
	}

	hasItemID, err := columnExists(db, "expenses", "item_id")
	if err != nil {
		return fmt.Errorf("check item_id column: %w", err)
	}
	if hasItemID {
		return nil
	}

	hasName, err := columnExists(db, "expenses", "name")
	if err != nil {
		return fmt.Errorf("check name column: %w", err)
	}
	if !hasName {
		return fmt.Errorf("expenses table has neither item_id nor name")
	}

	return migrateExpensesNameToItemID(db)
}

func migrateExpensesNameToItemID(db *sql.DB) error {
	hasShares, err := tableExists(db, "expense_shares")
	if err != nil {
		return fmt.Errorf("check expense_shares: %w", err)
	}

	// SQLite blocks DROP of a parent table while child FKs exist.
	if _, err := db.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
		return fmt.Errorf("disable foreign keys for migration: %w", err)
	}
	defer db.Exec(`PRAGMA foreign_keys = ON`)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin expenses item_id migration: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DROP TABLE IF EXISTS expenses_backup_name`); err != nil {
		return fmt.Errorf("drop stale expenses backup: %w", err)
	}
	if _, err := tx.Exec(`DROP TABLE IF EXISTS expense_shares_backup`); err != nil {
		return fmt.Errorf("drop stale shares backup: %w", err)
	}

	if _, err := tx.Exec(`
		CREATE TABLE expenses_backup_name (
			id INTEGER PRIMARY KEY,
			person_id INTEGER NOT NULL,
			shop_id INTEGER NOT NULL,
			date TEXT NOT NULL,
			name TEXT NOT NULL,
			amount REAL NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create expenses backup: %w", err)
	}

	if _, err := tx.Exec(`
		INSERT INTO expenses_backup_name (id, person_id, shop_id, date, name, amount)
		SELECT id, person_id, shop_id, date, name, amount FROM expenses
	`); err != nil {
		return fmt.Errorf("backup expenses: %w", err)
	}

	if _, err := tx.Exec(`
		CREATE TABLE expense_shares_backup (
			expense_id INTEGER NOT NULL,
			person_id INTEGER NOT NULL,
			share REAL NOT NULL,
			PRIMARY KEY (expense_id, person_id)
		)
	`); err != nil {
		return fmt.Errorf("create shares backup: %w", err)
	}

	if hasShares {
		if _, err := tx.Exec(`
			INSERT INTO expense_shares_backup (expense_id, person_id, share)
			SELECT expense_id, person_id, share FROM expense_shares
		`); err != nil {
			return fmt.Errorf("backup expense shares: %w", err)
		}
		if _, err := tx.Exec(`DROP TABLE expense_shares`); err != nil {
			return fmt.Errorf("drop expense shares: %w", err)
		}
	}

	if _, err := tx.Exec(`DROP TABLE expenses`); err != nil {
		return fmt.Errorf("drop old expenses: %w", err)
	}

	if _, err := tx.Exec(`
		CREATE TABLE expenses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			person_id INTEGER NOT NULL,
			shop_id INTEGER NOT NULL,
			item_id INTEGER NOT NULL,
			date TEXT NOT NULL,
			amount REAL NOT NULL,
			FOREIGN KEY (person_id) REFERENCES persons(id),
			FOREIGN KEY (shop_id) REFERENCES shops(id),
			FOREIGN KEY (item_id) REFERENCES items(id)
		)
	`); err != nil {
		return fmt.Errorf("recreate expenses: %w", err)
	}

	if _, err := tx.Exec(`
		CREATE TABLE expense_shares (
			expense_id INTEGER NOT NULL,
			person_id INTEGER NOT NULL,
			share REAL NOT NULL,
			PRIMARY KEY (expense_id, person_id),
			FOREIGN KEY (expense_id) REFERENCES expenses(id) ON DELETE CASCADE,
			FOREIGN KEY (person_id) REFERENCES persons(id)
		)
	`); err != nil {
		return fmt.Errorf("recreate expense shares: %w", err)
	}

	rows, err := tx.Query(`
		SELECT id, person_id, shop_id, date, name, amount
		FROM expenses_backup_name
		ORDER BY id
	`)
	if err != nil {
		return fmt.Errorf("select backup expenses: %w", err)
	}

	type backupRow struct {
		id       int
		personID int
		shopID   int
		date     string
		name     string
		amount   float64
	}
	var backups []backupRow
	for rows.Next() {
		var r backupRow
		if err := rows.Scan(&r.id, &r.personID, &r.shopID, &r.date, &r.name, &r.amount); err != nil {
			rows.Close()
			return fmt.Errorf("scan backup expense: %w", err)
		}
		backups = append(backups, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate backup expenses: %w", err)
	}

	for _, r := range backups {
		name := strings.TrimSpace(r.name)
		if name == "" {
			name = "Unknown"
		}
		itemID, err := upsertItemTx(tx, name)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(
			`INSERT INTO expenses (id, person_id, shop_id, item_id, date, amount) VALUES (?, ?, ?, ?, ?, ?)`,
			r.id, r.personID, r.shopID, itemID, r.date, r.amount,
		); err != nil {
			return fmt.Errorf("restore expense %d: %w", r.id, err)
		}
	}

	if _, err := tx.Exec(`
		INSERT INTO expense_shares (expense_id, person_id, share)
		SELECT expense_id, person_id, share FROM expense_shares_backup
	`); err != nil {
		return fmt.Errorf("restore expense shares: %w", err)
	}

	if _, err := tx.Exec(`DROP TABLE expenses_backup_name`); err != nil {
		return fmt.Errorf("drop expenses backup: %w", err)
	}
	if _, err := tx.Exec(`DROP TABLE expense_shares_backup`); err != nil {
		return fmt.Errorf("drop shares backup: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit expenses item_id migration: %w", err)
	}
	return nil
}

func upsertItemTx(tx *sql.Tx, name string) (int64, error) {
	var id int64
	err := tx.QueryRow(`SELECT id FROM items WHERE name = ? COLLATE NOCASE`, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("lookup item %q: %w", name, err)
	}

	result, err := tx.Exec(`INSERT INTO items (name) VALUES (?)`, name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			if err := tx.QueryRow(`SELECT id FROM items WHERE name = ? COLLATE NOCASE`, name).Scan(&id); err != nil {
				return 0, fmt.Errorf("resolve item after conflict %q: %w", name, err)
			}
			return id, nil
		}
		return 0, fmt.Errorf("insert item %q: %w", name, err)
	}
	return result.LastInsertId()
}

func migrateFromInvoices(db *sql.DB) error {
	var hasInvoices int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'table' AND name = 'invoices'
	`).Scan(&hasInvoices)
	if err != nil {
		return fmt.Errorf("check invoices table: %w", err)
	}
	if hasInvoices == 0 {
		return nil
	}

	var expenseCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM expenses`).Scan(&expenseCount); err != nil {
		return fmt.Errorf("count expenses: %w", err)
	}
	if expenseCount > 0 {
		return dropInvoiceTables(db)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin invoice migration: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
		SELECT ii.id, i.person_id, i.shop_id, i.date, ii.description, ii.amount
		FROM invoice_items ii
		JOIN invoices i ON i.id = ii.invoice_id
		ORDER BY ii.id
	`)
	if err != nil {
		if strings.Contains(err.Error(), "no such table") {
			return dropInvoiceTables(db)
		}
		return fmt.Errorf("select invoice items: %w", err)
	}

	type line struct {
		invoiceItemID int
		personID      int
		shopID        int
		date          string
		name          string
		amount        float64
	}
	var lines []line
	for rows.Next() {
		var l line
		if err := rows.Scan(&l.invoiceItemID, &l.personID, &l.shopID, &l.date, &l.name, &l.amount); err != nil {
			rows.Close()
			return fmt.Errorf("scan invoice item: %w", err)
		}
		lines = append(lines, l)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate invoice items: %w", err)
	}

	for _, l := range lines {
		name := strings.TrimSpace(l.name)
		if name == "" {
			name = "Unknown"
		}
		itemID, err := upsertItemTx(tx, name)
		if err != nil {
			return err
		}

		result, err := tx.Exec(
			`INSERT INTO expenses (person_id, shop_id, item_id, date, amount) VALUES (?, ?, ?, ?, ?)`,
			l.personID, l.shopID, itemID, l.date, l.amount,
		)
		if err != nil {
			return fmt.Errorf("insert expense: %w", err)
		}
		expenseID, _ := result.LastInsertId()

		shareRows, err := tx.Query(`
			SELECT person_id, share FROM invoice_shares
			WHERE invoice_id = (
				SELECT invoice_id FROM invoice_items WHERE id = ?
			)`, l.invoiceItemID)
		if err != nil {
			return fmt.Errorf("select invoice shares: %w", err)
		}

		shares := make(map[int]float64)
		for shareRows.Next() {
			var personID int
			var share float64
			if err := shareRows.Scan(&personID, &share); err != nil {
				shareRows.Close()
				return fmt.Errorf("scan invoice share: %w", err)
			}
			shares[personID] = share
		}
		shareRows.Close()

		if len(shares) == 0 {
			shares[1] = 0.5
			shares[2] = 0.5
		}
		for _, personID := range []int{1, 2} {
			share, ok := shares[personID]
			if !ok {
				share = 0.5
			}
			if _, err := tx.Exec(
				`INSERT INTO expense_shares (expense_id, person_id, share) VALUES (?, ?, ?)`,
				expenseID, personID, share,
			); err != nil {
				return fmt.Errorf("insert expense share: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit invoice migration: %w", err)
	}

	return dropInvoiceTables(db)
}

func dropInvoiceTables(db *sql.DB) error {
	for _, stmt := range []string{
		`DROP TABLE IF EXISTS invoice_shares`,
		`DROP TABLE IF EXISTS invoice_items`,
		`DROP TABLE IF EXISTS invoices`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("drop invoice tables: %w", err)
		}
	}
	return nil
}

func seedPersons(db *sql.DB) error {
	persons := []struct {
		id   int
		name string
	}{
		{1, "Armin"},
		{2, "Ramin"},
	}

	for _, p := range persons {
		_, err := db.Exec(
			`INSERT INTO persons (id, name) VALUES (?, ?)
			 ON CONFLICT(id) DO UPDATE SET name = excluded.name`,
			p.id, p.name,
		)
		if err != nil {
			return fmt.Errorf("seed person %d: %w", p.id, err)
		}
	}

	return nil
}

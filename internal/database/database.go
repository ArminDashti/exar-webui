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

	CREATE TABLE IF NOT EXISTS expenses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		person_id INTEGER NOT NULL,
		shop_id INTEGER NOT NULL,
		date TEXT NOT NULL,
		name TEXT NOT NULL,
		amount REAL NOT NULL,
		FOREIGN KEY (person_id) REFERENCES persons(id),
		FOREIGN KEY (shop_id) REFERENCES shops(id)
	);

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
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}

	if err := seedPersons(db); err != nil {
		return err
	}

	return migrateFromInvoices(db)
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
		// invoice_items may be missing on a half-created DB
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
		result, err := tx.Exec(
			`INSERT INTO expenses (person_id, shop_id, date, name, amount) VALUES (?, ?, ?, ?, ?)`,
			l.personID, l.shopID, l.date, l.name, l.amount,
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

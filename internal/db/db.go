package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

// New opens a SQLite database. The path can be overridden with DATABASE_PATH.
func New() *sql.DB {
	path := os.Getenv("DATABASE_PATH")
	if path == "" {
		path = "petstore.db"
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("unable to open database: %v", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		log.Fatalf("unable to enable foreign keys: %v", err)
	}
	if err := applyMigrations(db); err != nil {
		log.Fatalf("unable to apply migrations: %v", err)
	}
	return db
}

func applyMigrations(db *sql.DB) error {
	entries, err := os.ReadDir("migrations")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		path := filepath.Join("migrations", e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("%s: %w", e.Name(), err)
		}
	}
	return nil
}

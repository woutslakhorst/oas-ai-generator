package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/woutslakhorst/oas-ai-generator/assets"
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
	entries, err := fs.ReadDir(assets.Migrations, "migrations")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		data, err := assets.Migrations.ReadFile("migrations/" + e.Name())
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("%s: %w", e.Name(), err)
		}
	}
	return nil
}

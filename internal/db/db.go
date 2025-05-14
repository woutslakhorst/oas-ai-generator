package db

import (
	"database/sql"
	"log"
	"os"

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
	return db
}

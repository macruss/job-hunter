package storage

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func Connect(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("storage: open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("storage: ping: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return &DB{db}, nil
}

// RunMigrations executes the SQL file at path.
// For production, prefer golang-migrate or goose; this is enough for MVP.
func (db *DB) RunMigrations(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("migrations: read %s: %w", path, err)
	}
	if _, err := db.Exec(string(data)); err != nil {
		return fmt.Errorf("migrations: exec: %w", err)
	}
	return nil
}

package database

import (
	"database/sql"
	_ "github.com/marcboeker/go-duckdb"
)

// Connect establishes a connection to the DuckDB database
func Connect(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, err
	}
	
	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	
	return db, nil
}
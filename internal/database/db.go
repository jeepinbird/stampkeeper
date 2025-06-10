package database

import (
	"database/sql"
	_ "github.com/lib/pq"
)

// Connect establishes a connection to the PostgreSQL database
func Connect(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
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
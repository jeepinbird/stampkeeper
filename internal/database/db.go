package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// Connect establishes a connection to the SQLite database
func Connect(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	
	// Enable foreign key constraints (important for SQLite)
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		db.Close()
		return nil, err
	}
	
	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	
	return db, nil
}
package database

import (
	"database/sql"
	"log"
)

// Rollback handles transaction rollback with proper error logging
// Uses CRITICAL log level for rollback failures which may indicate database issues
func Rollback(tx *sql.Tx, context string) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		log.Printf("CRITICAL: Transaction rollback failed [%s]: %v", context, err)
		// TODO: Add alerting/monitoring here for production
		// A failed rollback could indicate serious database issues
	}
}

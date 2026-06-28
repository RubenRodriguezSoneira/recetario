// Package database centralizes the SQLite schema so that the server entrypoint
// and tests apply an identical set of tables. The schema is embedded from
// schema.sql, keeping a single source of truth for the database structure.
package database

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var schemaSQL string

// Schema returns the embedded SQLite DDL used to provision the database.
func Schema() string {
	return schemaSQL
}

// ApplySchema creates every table the application needs if it does not already
// exist. It is safe to call repeatedly. The modernc.org/sqlite driver executes
// the multi-statement script in a single Exec because the script carries no
// bound parameters.
func ApplySchema(db *sql.DB) error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}
	return nil
}

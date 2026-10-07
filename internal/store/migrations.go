package store

import (
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var initialSchema string

const schemaVersion = 2

func (s *Store) migrate() error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var version int
	if err = tx.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version > schemaVersion {
		return fmt.Errorf("database schema %d is newer than supported version %d", version, schemaVersion)
	}
	if version < 1 {
		if _, err = tx.Exec(initialSchema); err != nil {
			return err
		}
	}
	if version < 2 {
		if _, err = tx.Exec("ALTER TABLE settings ADD COLUMN revision INTEGER NOT NULL DEFAULT 0; PRAGMA user_version=2;"); err != nil {
			return err
		}
	}
	return tx.Commit()
}

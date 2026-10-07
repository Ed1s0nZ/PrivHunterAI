package store

import (
	"database/sql"
	"errors"
)

func (s *Store) BulkReview(ids []int64, status string) error {
	if len(ids) == 0 || len(ids) > 100 {
		return errors.New("batch size invalid")
	}
	tx, e := s.DB.Begin()
	if e != nil {
		return e
	}
	defer tx.Rollback()
	seen := map[int64]bool{}
	for _, id := range ids {
		if id < 1 || seen[id] {
			return errors.New("invalid ids")
		}
		seen[id] = true
		r, e := tx.Exec("UPDATE findings SET review=? WHERE id=?", status, id)
		if e != nil {
			return e
		}
		n, e := r.RowsAffected()
		if e != nil {
			return e
		}
		if n != 1 {
			return sql.ErrNoRows
		}
	}
	return tx.Commit()
}

package store

import (
	"encoding/json"
	"yuequanScan/internal/domain"
)

func (s *Store) Finding(id int64) (domain.Finding, error) {
	var v domain.Finding
	var raw, review, note string
	e := s.DB.QueryRow("SELECT payload,review,note FROM findings WHERE id=?", id).Scan(&raw, &review, &note)
	if e != nil {
		return v, e
	}
	if e = json.Unmarshal([]byte(raw), &v); e != nil {
		return v, e
	}
	v.ID = id
	v.Review = review
	v.Note = note
	return v, nil
}

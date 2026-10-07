package store

import (
	"encoding/json"
	"errors"

	"yuequanScan/internal/domain"
)

var ErrExportLimit = errors.New("export exceeds size limit")

// Export reads one SQLite snapshot and bounds both evidence size and row count.
func (s *Store) Export(f Filter) ([]domain.Finding, error) {
	rows, err := s.DB.Query(`SELECT id,payload,review,note FROM findings WHERE (?='' OR result=?) AND (?='' OR review=?) AND (?='' OR instr(lower(url),lower(?))>0) ORDER BY id DESC LIMIT 10001`, f.Result, f.Result, f.Review, f.Review, f.Search, f.Search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Finding{}
	size := 0
	for rows.Next() {
		var raw, note, review string
		var id int64
		if err = rows.Scan(&id, &raw, &review, &note); err != nil {
			return nil, err
		}
		size += len(raw) + len(note)
		if len(out) >= 10000 || size > 32<<20 {
			return nil, ErrExportLimit
		}
		var v domain.Finding
		if err = json.Unmarshal([]byte(raw), &v); err != nil {
			return nil, err
		}
		v.ID = id
		v.Review = review
		v.Note = note
		out = append(out, v)
	}
	return out, rows.Err()
}

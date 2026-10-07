package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"yuequanScan/internal/domain"
)

var ErrSettingsConflict = errors.New("settings revision conflict")

type storedSettings struct {
	Settings domain.Settings   `json:"settings"`
	Headers  map[string]string `json:"headers"`
	Key      string            `json:"key"`
}

func (s *Store) SaveSettings(v domain.Settings) error { return s.saveSettings(v, nil) }
func (s *Store) SaveSettingsRevision(v domain.Settings, revision int64) error {
	return s.saveSettings(v, &revision)
}
func (s *Store) saveSettings(v domain.Settings, revision *int64) error {
	// The local representation contains secrets; public JSON does not.
	b, e := json.Marshal(storedSettings{v, v.Headers, v.APIKey})
	if e != nil {
		return e
	}
	if revision == nil {
		_, e = s.DB.Exec("INSERT INTO settings(id,payload,revision) VALUES(1,?,1) ON CONFLICT(id) DO UPDATE SET payload=excluded.payload, revision=settings.revision+1", string(b))
		return e
	}
	r, e := s.DB.Exec(`INSERT INTO settings(id,payload,revision) SELECT 1,?,1 WHERE ?=0 OR EXISTS(SELECT 1 FROM settings WHERE id=1) ON CONFLICT(id) DO UPDATE SET payload=excluded.payload,revision=settings.revision+1 WHERE settings.revision=?`, string(b), *revision, *revision)
	if e != nil {
		return e
	}
	n, e := r.RowsAffected()
	if e == nil && n == 0 {
		return ErrSettingsConflict
	}
	return e
}
func (s *Store) Settings() (domain.Settings, error) {
	var raw string
	var revision int64
	var v storedSettings
	e := s.DB.QueryRow("SELECT payload,revision FROM settings WHERE id=1").Scan(&raw, &revision)
	if e == sql.ErrNoRows {
		return domain.Settings{Paused: true, Domains: []string{}, Headers: map[string]string{}, SkipSuffixes: []string{".js", ".css", ".png", ".jpg", ".ico"}}, nil
	}
	if e != nil {
		return v.Settings, e
	}
	e = json.Unmarshal([]byte(raw), &v)
	v.Settings.Headers = v.Headers
	v.Settings.APIKey = v.Key
	v.Settings.Revision = revision
	return v.Settings, e
}

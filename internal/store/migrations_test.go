package store

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"yuequanScan/internal/domain"
)

func TestMigrationAndRevision(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")
	db, e := sql.Open("sqlite3", path)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(initialSchema); e != nil {
		t.Fatal(e)
	}
	db.Close()
	s, e := Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	v, e := s.Settings()
	if e != nil {
		t.Fatal(e)
	}
	v.APIKey = "synthetic"
	if e = s.SaveSettingsRevision(v, 0); e != nil {
		t.Fatal(e)
	}
	v, e = s.Settings()
	if e != nil || v.Revision != 1 || v.APIKey != "synthetic" {
		t.Fatal("migration/save failed", e)
	}
	fresh := v
	fresh.Model = "new"
	if e = s.SaveSettingsRevision(fresh, v.Revision); e != nil {
		t.Fatal(e)
	}
	if e = s.SaveSettingsRevision(domain.Settings{Model: "stale"}, v.Revision); e != ErrSettingsConflict {
		t.Fatal("stale overwrite accepted", e)
	}
	got, _ := s.Settings()
	if got.Model != "new" || got.APIKey != "synthetic" {
		t.Fatal("stale overwrite changed config")
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Fatal("unsafe database permissions")
	}
}
func TestFutureSchemaRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "future.db")
	db, _ := sql.Open("sqlite3", path)
	db.Exec("PRAGMA user_version=999")
	db.Close()
	if s, e := Open(path); e == nil {
		s.Close()
		t.Fatal("newer schema opened")
	}
}

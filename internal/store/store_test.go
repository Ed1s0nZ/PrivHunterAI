package store

import (
	"path/filepath"
	"testing"
	"yuequanScan/internal/domain"
)

func TestPersistenceAndBootstrap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	s, e := Open(path)
	if e != nil {
		t.Fatal(e)
	}
	ok, e := s.Bootstrap("admin", "hash")
	if e != nil || !ok {
		t.Fatal("bootstrap failed", e)
	}
	ok, e = s.Bootstrap("other", "hash")
	if e != nil || ok {
		t.Fatal("bootstrap accepted twice", e)
	}
	_, e = s.AddFinding(domain.Finding{URL: "https://example.org/order", Method: "GET", Result: "unknown"})
	if e != nil {
		t.Fatal(e)
	}
	v := domain.Settings{Paused: true, APIKey: "fixture", Headers: map[string]string{"Cookie": "fixture"}}
	if e = s.SaveSettings(v); e != nil {
		t.Fatal(e)
	}
	s.Close()
	s, e = Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	got, total, e := s.Findings(Filter{Search: "order"})
	if e != nil || total != 1 || len(got) != 1 {
		t.Fatal("persistence failed", e)
	}
	if e = s.Review(got[0].ID, "confirmed", "checked"); e != nil {
		t.Fatal(e)
	}
	got, total, e = s.Findings(Filter{Review: "confirmed"})
	if e != nil || total != 1 || got[0].Note != "checked" {
		t.Fatal("review filter failed", e)
	}
	restored, e := s.Settings()
	if e != nil || restored.APIKey != "fixture" || restored.Headers["Cookie"] != "fixture" {
		t.Fatal("settings persistence failed", e)
	}
}

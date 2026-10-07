package scanner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"yuequanScan/internal/domain"
	"yuequanScan/internal/store"
)

func TestReplayPersistsRedactedEvidence(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fixture-b" {
			t.Error("replacement missing")
		}
		w.WriteHeader(403)
		w.Write([]byte(`{"token":"fixture-secret"}`))
	}))
	defer target.Close()
	db, err := store.Open(filepath.Join(t.TempDir(), "scan.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.SaveSettings(domain.Settings{Domains: []string{"127.0.0.1"}, Headers: map[string]string{"Authorization": "Bearer fixture-b"}}); err != nil {
		t.Fatal(err)
	}
	engine := New(db)
	original := http.Header{"Authorization": {"Bearer fixture-a"}}
	capture := Capture{Method: "GET", URL: target.URL + "/?token=fixture-secret", Header: original, Response: []byte(`{"email":"fixture-secret"}`)}
	engine.process(context.Background(), capture)
	if original.Get("Authorization") != "Bearer fixture-a" {
		t.Fatal("original headers mutated")
	}
	v, total, err := db.Findings(store.Filter{})
	if err != nil || total != 1 {
		t.Fatal("result not saved", err)
	}
	if v[0].Result != "false" {
		t.Fatal("403 not denied")
	}
	detail, err := db.Finding(v[0].ID)
	if err != nil || detail.RequestA == "" || detail.ResponseB == "" {
		t.Fatal("evidence detail missing", err)
	}
	for _, text := range []string{detail.RequestA, detail.RequestB, detail.ResponseA, detail.ResponseB, detail.URL} {
		if containsFixture(text) {
			t.Fatal("secret persisted")
		}
	}
	engine.process(context.Background(), capture)
	_, total, _ = db.Findings(store.Filter{})
	if total != 1 {
		t.Fatal("duplicate not suppressed")
	}
}
func containsFixture(v string) bool {
	for i := 0; i+7 <= len(v); i++ {
		if v[i:i+7] == "fixture" {
			return true
		}
	}
	return false
}

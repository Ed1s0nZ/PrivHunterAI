package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"yuequanScan/internal/auth"
	"yuequanScan/internal/store"
)

func TestSettingsDisclosureAndViewerPermissions(t *testing.T) {
	db, e := store.Open(filepath.Join(t.TempDir(), "settings.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	hash, _ := auth.Hash("fictional-password-123")
	db.Bootstrap("admin", hash)
	db.CreateUser("viewer", hash, "viewer")
	server := New(db, false)
	router := server.Router()
	token, session, e := server.Auth.Login("admin", "fictional-password-123")
	if e != nil {
		t.Fatal(e)
	}
	call := func(method, path, body, token, csrf string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Cookie", "privhunter_session="+token)
		r.Header.Set("X-CSRF-Token", csrf)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	body := `{"revision":0,"paused":true,"domains":["example.test"],"apiKey":"synthetic-key","headers":{"Cookie":"synthetic-cookie"},"endpoint":"https://example.test/chat","model":"fixture"}`
	if w := call("PUT", "/api/settings", body, token, session.CSRF); w.Code != 200 {
		t.Fatal("save failed", w.Code)
	}
	w := call("GET", "/api/settings", "", token, "")
	if w.Code != 200 || strings.Contains(w.Body.String(), "synthetic") {
		t.Fatal("secret disclosed")
	}
	var v struct {
		HasKey     bool `json:"hasAPIKey"`
		HasHeaders bool `json:"hasHeaders"`
	}
	json.Unmarshal(w.Body.Bytes(), &v)
	if !v.HasKey || !v.HasHeaders {
		t.Fatal("missing configured flags")
	}
	viewer, vs, e := server.Auth.Login("viewer", "fictional-password-123")
	if e != nil {
		t.Fatal(e)
	}
	for _, path := range []string{"/api/settings", "/api/users", "/api/audit"} {
		if w := call("GET", path, "", viewer, vs.CSRF); w.Code != 403 {
			t.Fatal("viewer access", path, w.Code)
		}
	}
	if w := call("PUT", "/api/settings", body, viewer, vs.CSRF); w.Code != 403 {
		t.Fatal("viewer mutated settings")
	}
}
func TestCSVFormulaGuard(t *testing.T) {
	for _, value := range []string{"=1+1", "+cmd", "-cmd", "@cmd", "\tformula", "\rformula"} {
		if !strings.HasPrefix(safeCell(value), "'") {
			t.Fatal("formula not escaped")
		}
	}
}

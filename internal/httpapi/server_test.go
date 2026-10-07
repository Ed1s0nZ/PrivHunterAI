package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"yuequanScan/internal/store"
)

func TestAuthAndCSRF(t *testing.T) {
	db, e := store.Open(filepath.Join(t.TempDir(), "api.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	router := New(db, false).Router()
	request := func(method, path, body string, cookie *http.Cookie, csrf string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		r.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			r.AddCookie(cookie)
		}
		r.Header.Set("X-CSRF-Token", csrf)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	if w := request("GET", "/api/findings", "", nil, ""); w.Code != 401 {
		t.Fatal("unprotected findings", w.Code)
	}
	if w := request("GET", "/api/findings/1", "", nil, ""); w.Code != 401 {
		t.Fatal("unprotected evidence", w.Code)
	}
	if w := request("POST", "/api/auth/setup", `{"username":"admin","password":"test-password-123"}`, nil, ""); w.Code != 201 {
		t.Fatal("setup failed", w.Code)
	}
	w := request("POST", "/api/auth/login", `{"username":"admin","password":"test-password-123"}`, nil, "")
	if w.Code != 200 {
		t.Fatal("login failed", w.Code)
	}
	cookies := w.Result().Cookies()
	if w := request("GET", "/api/findings/999", "", cookies[0], ""); w.Code != 404 {
		t.Fatal("missing evidence status", w.Code)
	}
	if len(cookies) != 1 || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatal("unsafe cookie")
	}
	var payload struct {
		CSRF string `json:"csrf"`
	}
	if e = json.Unmarshal(w.Body.Bytes(), &payload); e != nil {
		t.Fatal(e)
	}
	if w = request("POST", "/api/auth/logout", "{}", cookies[0], ""); w.Code != 403 {
		t.Fatal("CSRF bypass", w.Code)
	}
	if w = request("POST", "/api/auth/logout", "{}", cookies[0], payload.CSRF); w.Code != 200 {
		t.Fatal("logout failed", w.Code)
	}
	if w = request("GET", "/api/auth/me", "", cookies[0], ""); w.Code != 401 {
		t.Fatal("logout did not revoke", w.Code)
	}
}

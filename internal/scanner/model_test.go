package scanner

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yuequanScan/internal/domain"
)

func TestModelProtocol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fixture" {
			t.Error("authorization missing")
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "responseA") {
			t.Error("evidence missing")
		}
		io.WriteString(w, `{"choices":[{"message":{"content":"{\"res\":\"unknown\",\"reason\":\"Insufficient identity evidence\",\"confidence\":\"50%\"}"}}]}`)
	}))
	defer server.Close()
	d, e := Analyze(context.Background(), domain.Settings{Endpoint: server.URL, Model: "fixture", APIKey: "fixture"}, domain.Finding{ResponseA: "{}", ResponseB: "{}"}, 200)
	if e != nil || d.Result != "unknown" {
		t.Fatal("model protocol failed", e)
	}
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, `{"choices":[]}`) }))
	defer bad.Close()
	if _, e = Analyze(context.Background(), domain.Settings{Endpoint: bad.URL}, domain.Finding{}, 200); e == nil {
		t.Fatal("empty choices accepted")
	}
}

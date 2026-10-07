package scanner

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/lqqyt2423/go-mitmproxy/proxy"
	"yuequanScan/internal/domain"
	"yuequanScan/internal/store"
)

func TestPassiveProxyPipeline(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") == "Bearer account-b" {
			w.WriteHeader(403)
		}
		io.WriteString(w, `{"token":"synthetic-secret","id":1}`)
	}))
	defer target.Close()
	db, e := store.Open(filepath.Join(t.TempDir(), "proxy.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = db.SaveSettings(domain.Settings{Domains: []string{"127.0.0.1"}, Headers: map[string]string{"Authorization": "Bearer account-b"}}); e != nil {
		t.Fatal(e)
	}
	listener, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	addr := listener.Addr().String()
	listener.Close()
	engine := New(db)
	ctx, cancel := context.WithCancel(context.Background())
	engine.Run(ctx)
	defer func() { cancel(); engine.Wait() }()
	p, e := proxy.NewProxy(&proxy.Options{Addr: addr, StreamLargeBodies: maxBody})
	if e != nil {
		t.Fatal(e)
	}
	p.AddAddon(engine)
	done := make(chan error, 1)
	go func() { done <- p.Start() }()
	defer func() { p.Close(); <-done }()
	proxyURL, _ := url.Parse("http://" + addr)
	transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: time.Second}
	var resp *http.Response
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		req, _ := http.NewRequest("GET", target.URL+"/orders", nil)
		req.Header.Set("Authorization", "Bearer account-a")
		resp, e = client.Do(req)
		if e == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if e != nil {
		t.Fatal("proxy request failed", e)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	for time.Now().Before(deadline) {
		records, total, e := db.Findings(store.Filter{})
		if e != nil {
			t.Fatal(e)
		}
		if total == 1 {
			if records[0].Result != "false" {
				t.Fatal("wrong verdict")
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("proxy did not persist result")
}

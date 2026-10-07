package scanner

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lqqyt2423/go-mitmproxy/proxy"
	"yuequanScan/internal/domain"
	"yuequanScan/internal/security"
	"yuequanScan/internal/store"
)

const maxBody = 1 << 20

type Capture struct {
	Method, URL    string
	Header         http.Header
	Body, Response []byte
	Status         int
}
type Engine struct {
	proxy.BaseAddon
	Store     *store.Store
	queue     chan Capture
	mu        sync.Mutex
	seen      map[[32]byte]time.Time
	Dropped   atomic.Int64
	Processed atomic.Int64
	wg        sync.WaitGroup
}

func New(db *store.Store) *Engine {
	return &Engine{Store: db, queue: make(chan Capture, 64), seen: map[[32]byte]time.Time{}}
}
func (e *Engine) Response(f *proxy.Flow) {
	if f.Request == nil || f.Response == nil || f.Request.URL == nil || len(f.Request.Body) > maxBody || len(f.Response.Body) > maxBody {
		return
	}
	settings, err := e.Store.Settings()
	if err != nil || settings.Paused || !inScope(f.Request.URL.Hostname(), settings.Domains) {
		return
	}
	for _, suffix := range settings.SkipSuffixes {
		if suffix != "" && strings.HasSuffix(strings.ToLower(f.Request.URL.Path), strings.ToLower(suffix)) {
			return
		}
	}
	ct := strings.ToLower(f.Response.Header.Get("Content-Type"))
	if strings.HasPrefix(ct, "image/") || strings.HasPrefix(ct, "video/") || strings.HasPrefix(ct, "audio/") {
		return
	}
	decoded, err := decodeBody(f.Response.Body, f.Response.Header.Get("Content-Encoding"))
	if err != nil {
		return
	}
	c := Capture{f.Request.Method, f.Request.URL.String(), f.Request.Header.Clone(), bytes.Clone(f.Request.Body), bytes.Clone(decoded), f.Response.StatusCode}
	select {
	case e.queue <- c:
	default:
		e.Dropped.Add(1)
	}
}
func inScope(host string, domains []string) bool {
	for _, d := range domains {
		if strings.EqualFold(host, d) {
			return true
		}
	}
	return false
}
func (e *Engine) Run(ctx context.Context) {
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case c := <-e.queue:
				e.process(ctx, c)
			}
		}
	}()
}
func (e *Engine) Wait() { e.wg.Wait() }
func (e *Engine) Status() map[string]any {
	return map[string]any{"queued": len(e.queue), "dropped": e.Dropped.Load(), "processed": e.Processed.Load()}
}
func (e *Engine) process(ctx context.Context, c Capture) {
	settings, err := e.Store.Settings()
	if err != nil || settings.Paused {
		return
	}
	target, parseErr := url.Parse(c.URL)
	if parseErr != nil || !inScope(target.Hostname(), settings.Domains) {
		return
	}
	if key := os.Getenv("PRIVHUNTER_API_KEY"); key != "" {
		settings.APIKey = key
	}
	digest := sha256.Sum256([]byte(c.Method + c.URL + string(c.Body) + fmt.Sprint(c.Header) + fmt.Sprint(settings.Headers)))
	e.mu.Lock()
	now := time.Now()
	for k, t := range e.seen {
		if now.Sub(t) > 5*time.Minute {
			delete(e.seen, k)
		}
	}
	if _, ok := e.seen[digest]; ok {
		e.mu.Unlock()
		return
	}
	e.seen[digest] = now
	e.mu.Unlock()
	v := domain.Finding{Method: c.Method, URL: security.URL(c.URL), RequestA: requestEvidence(c.Method, c.URL, c.Header, c.Body), ResponseA: security.Text(string(c.Response)), Result: "unknown", Confidence: "—"}
	req, err := http.NewRequestWithContext(ctx, c.Method, c.URL, bytes.NewReader(c.Body))
	if err != nil {
		return
	}
	req.Header = c.Header.Clone()
	req.Header.Del("Accept-Encoding")
	for key, value := range settings.Headers {
		req.Header.Set(key, value)
	}
	v.RequestB = requestEvidence(c.Method, c.URL, req.Header, c.Body)
	client := &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		v.Result = "error"
		v.Reason = "账号 B 请求失败或超时"
	} else {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
		resp.Body.Close()
		if readErr == nil {
			body, readErr = decodeBody(body, resp.Header.Get("Content-Encoding"))
		}
		if readErr != nil || len(body) > maxBody {
			v.Result = "error"
			v.Reason = "账号 B 响应超过上限或读取失败"
		} else {
			v.ResponseB = security.Text(string(body))
			switch {
			case resp.StatusCode == 401 || resp.StatusCode == 403:
				v.Result = "false"
				v.Reason = "账号 B 被认证或权限检查拒绝"
				v.Confidence = "100%"
			case resp.StatusCode >= 500:
				v.Reason = "账号 B 返回服务器错误，无法判断"
			default:
				denied := false
				for _, word := range settings.DenyKeywords {
					if word != "" && strings.Contains(string(body), word) {
						denied = true
						break
					}
				}
				if denied {
					v.Reason = "响应包含拒绝关键词，需要人工核实"
				} else if settings.Endpoint != "" && settings.Model != "" && settings.APIKey != "" {
					decision, aiErr := Analyze(ctx, settings, v, resp.StatusCode)
					if aiErr != nil {
						v.Result = "error"
						v.Reason = "模型分析失败，请检查模型设置或服务状态"
					} else {
						v.Result = decision.Result
						v.Reason = security.Text(decision.Reason)
						v.Confidence = decision.Confidence
					}
				} else {
					v.Reason = "未配置模型，已保留对比证据供人工复核"
				}
			}
		}
	}
	if _, err = e.Store.AddFinding(v); err == nil {
		e.Processed.Add(1)
	}
}
func requestEvidence(method, raw string, headers http.Header, body []byte) string {
	h := security.Headers(headers)
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(method + " " + security.URL(raw) + "\n")
	for _, k := range keys {
		b.WriteString(k + ": " + strings.Join(h[k], ", ") + "\n")
	}
	b.WriteString("\n" + security.Text(string(body)))
	return b.String()
}

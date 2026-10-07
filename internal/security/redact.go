// Package security centralizes disclosure rules for stored and exported evidence.
package security

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const Mask = "[REDACTED]"

func Sensitive(name string) bool {
	n := strings.ToLower(strings.ReplaceAll(name, "-", "_"))
	for _, word := range []string{"authorization", "cookie", "password", "passwd", "secret", "token", "api_key", "apikey", "credential", "session", "phone", "email", "id_card"} {
		if strings.Contains(n, word) {
			return true
		}
	}
	return false
}

func Headers(src http.Header) http.Header {
	dst := src.Clone()
	for k := range dst {
		if Sensitive(k) {
			dst[k] = []string{Mask}
		}
	}
	return dst
}

func URL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return Mask
	}
	if u.User != nil {
		u.User = url.User(Mask)
	}
	q := u.Query()
	for k := range q {
		if Sensitive(k) {
			q.Set(k, Mask)
		}
	}
	u.RawQuery = q.Encode()
	u.Fragment = ""
	return u.String()
}

var bearer = regexp.MustCompile(`(?i)\bBearer\s+[^\s",;]+`)
var assignments = regexp.MustCompile(`(?i)(["']?(?:password|passwd|secret|token|api[_-]?key|authorization|cookie|email|phone)["']?\s*[:=]\s*)(?:"[^"\r\n]*"|'[^'\r\n]*'|[^\s,;&\r\n]+)`)

func Text(s string) string {
	var v any
	if json.Unmarshal([]byte(s), &v) == nil {
		scrub(v)
		if b, err := json.MarshalIndent(v, "", "  "); err == nil {
			return string(b)
		}
	}
	s = bearer.ReplaceAllString(s, "Bearer "+Mask)
	return assignments.ReplaceAllString(s, "${1}"+Mask)
}

func scrub(v any) {
	switch x := v.(type) {
	case map[string]any:
		for k, value := range x {
			if Sensitive(k) {
				x[k] = Mask
			} else {
				scrub(value)
				if s, ok := value.(string); ok {
					x[k] = bearer.ReplaceAllString(s, "Bearer "+Mask)
				}
			}
		}
	case []any:
		for _, value := range x {
			scrub(value)
		}
	}
}

package security

import (
	"net/http"
	"strings"
	"testing"
)

func TestDisclosure(t *testing.T) {
	h := http.Header{"Authorization": {"Bearer private"}, "Cookie": {"sid=private"}, "Content-Type": {"application/json"}}
	redacted := Headers(h)
	if redacted.Get("Authorization") != Mask || h.Get("Authorization") != "Bearer private" {
		t.Fatal("header clone or redaction failed")
	}
	for _, sample := range []string{`{"nested":[{"access_token":"private","email":"private"}],"ok":1}`, "password=private&ok=1", "Authorization: Bearer private"} {
		if strings.Contains(Text(sample), "private") {
			t.Fatal("sensitive value survived")
		}
	}
	if strings.Contains(URL("https://user:private@example.org/a?access_token=private&ok=1"), "private") {
		t.Fatal("URL credential survived")
	}
}

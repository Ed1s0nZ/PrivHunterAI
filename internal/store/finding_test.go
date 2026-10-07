package store

import (
	"path/filepath"
	"strings"
	"testing"
	"yuequanScan/internal/domain"
)

func TestSummaryAndEvidence(t *testing.T) {
	s, e := Open(filepath.Join(t.TempDir(), "findings.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	body := strings.Repeat("evidence", 10000)
	id, e := s.AddFinding(domain.Finding{Method: "GET", URL: "https://example.test", Result: "unknown", Reason: "reason", ResponseA: body})
	if e != nil {
		t.Fatal(e)
	}
	rows, total, e := s.Findings(Filter{})
	if e != nil || total != 1 || rows[0].ResponseA != "" || rows[0].Reason != "reason" {
		t.Fatal("summary projection incorrect", e)
	}
	v, e := s.Finding(id)
	if e != nil || v.ResponseA != body {
		t.Fatal("evidence detail lost", e)
	}
	out, e := s.Export(Filter{})
	if e != nil || out[0].ResponseA != body {
		t.Fatal("export lost evidence", e)
	}
}

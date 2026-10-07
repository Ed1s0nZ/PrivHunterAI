package store

import (
	"path/filepath"
	"testing"
	"yuequanScan/internal/domain"
)

func TestDashboardAndAtomicBulk(t *testing.T) {
	s, e := Open(filepath.Join(t.TempDir(), "dashboard.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	a, _ := s.AddFinding(domain.Finding{URL: "https://example.test/a", Result: "true"})
	b, _ := s.AddFinding(domain.Finding{URL: "https://example.test/b", Result: "false"})
	if e = s.BulkReview([]int64{a, 99999}, "confirmed"); e == nil {
		t.Fatal("missing id accepted")
	}
	d, e := s.Dashboard()
	if e != nil || d.Reviews["pending"] != 2 {
		t.Fatal("partial update or wrong count", e)
	}
	if e = s.BulkReview([]int64{a, b}, "confirmed"); e != nil {
		t.Fatal(e)
	}
	d, e = s.Dashboard()
	if e != nil || d.Reviews["confirmed"] != 2 || len(d.Trend) != 7 || d.Trend[6].Total != 2 || len(d.Targets) != 1 || d.Targets[0].Pending != 0 {
		t.Fatal("dashboard aggregation failed", e)
	}
}

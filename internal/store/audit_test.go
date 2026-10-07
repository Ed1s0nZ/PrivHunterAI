package store

import (
	"path/filepath"
	"testing"
)

func TestAuditFilter(t *testing.T) {
	s, e := Open(filepath.Join(t.TempDir(), "audit.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	s.Log("alice", "review", "42")
	s.Log("bob", "login", "session")
	s.Log("alice", "login", "session")
	rows, total, e := s.Audit(AuditFilter{Search: "alice", Action: "review"})
	if e != nil || total != 1 || len(rows) != 1 || rows[0].Target != "42" {
		t.Fatal("audit filter incorrect", e)
	}
	rows, total, e = s.Audit(AuditFilter{Search: "' OR 1=1 --"})
	if e != nil || total != 0 || len(rows) != 0 {
		t.Fatal("search injection")
	}
}

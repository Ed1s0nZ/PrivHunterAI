package auth

import (
	"path/filepath"
	"testing"
	"yuequanScan/internal/store"
)

func TestSessionRevocationAndLastAdmin(t *testing.T) {
	db, e := store.Open(filepath.Join(t.TempDir(), "auth.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	s := Service{Store: db}
	hash, e := Hash("test-password-123")
	if e != nil {
		t.Fatal(e)
	}
	ok, e := db.Bootstrap("admin", hash)
	if e != nil || !ok {
		t.Fatal(e)
	}
	token, session, e := s.Login("admin", "test-password-123")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Authenticate(token); e != nil {
		t.Fatal(e)
	}
	if e = s.SetUser(session.User.ID, "viewer", false); e == nil {
		t.Fatal("last admin demoted")
	}
	if e = s.ChangePassword(session.User.ID, "test-password-123", "changed-password-123"); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Authenticate(token); e == nil {
		t.Fatal("old session survived password change")
	}
	if _, _, e = s.Login("admin", "test-password-123"); e == nil {
		t.Fatal("old password accepted")
	}
	token, _, e = s.Login("admin", "changed-password-123")
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Logout(token); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Authenticate(token); e == nil {
		t.Fatal("logged-out session accepted")
	}
}

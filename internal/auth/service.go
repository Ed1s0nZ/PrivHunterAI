package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
	"yuequanScan/internal/domain"
	"yuequanScan/internal/store"
)

var ErrCredentials = errors.New("用户名或密码错误")

type Service struct{ Store *store.Store }
type Session struct {
	User    domain.User
	CSRF    string
	Expires time.Time
}

func Validate(username, password string) error {
	if len(username) < 3 || len(username) > 64 {
		return errors.New("用户名需为 3–64 个字符")
	}
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' && r != '.' {
			return errors.New("用户名仅支持字母、数字、点、下划线和连字符")
		}
	}
	if len(password) < 12 || len(password) > 72 {
		return errors.New("密码需为 12–72 字节")
	}
	if strings.TrimSpace(password) == "" {
		return errors.New("密码不能为空")
	}
	return nil
}
func Hash(password string) (string, error) {
	b, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), e
}
func digest(token string) string { h := sha256.Sum256([]byte(token)); return hex.EncodeToString(h[:]) }
func random() (string, error) {
	b := make([]byte, 32)
	_, e := rand.Read(b)
	return hex.EncodeToString(b), e
}
func (s *Service) Login(name, password string) (string, Session, error) {
	u, h, e := s.Store.UserByName(name)
	if e != nil || u.Disabled { // Also spend password-hash time for nonexistent users.
		bcrypt.CompareHashAndPassword([]byte("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"), []byte(password))
		return "", Session{}, ErrCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(h), []byte(password)) != nil {
		return "", Session{}, ErrCredentials
	}
	token, e := random()
	if e != nil {
		return "", Session{}, e
	}
	csrf, e := random()
	if e != nil {
		return "", Session{}, e
	}
	expires := time.Now().Add(12 * time.Hour)
	_, e = s.Store.DB.Exec("INSERT INTO sessions(token_hash,user_id,csrf,expires_at) VALUES(?,?,?,?)", digest(token), u.ID, csrf, expires.Unix())
	if e != nil {
		return "", Session{}, e
	}
	s.Store.DB.Exec("DELETE FROM sessions WHERE expires_at<=?", time.Now().Unix())
	return token, Session{u, csrf, expires}, nil
}
func (s *Service) Authenticate(token string) (Session, error) {
	if len(token) != 64 {
		return Session{}, ErrCredentials
	}
	var session Session
	var expiry int64
	e := s.Store.DB.QueryRow(`SELECT u.id,u.username,u.role,u.disabled,s.csrf,s.expires_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE token_hash=? AND expires_at>? AND u.disabled=0`, digest(token), time.Now().Unix()).Scan(&session.User.ID, &session.User.Username, &session.User.Role, &session.User.Disabled, &session.CSRF, &expiry)
	session.Expires = time.Unix(expiry, 0)
	return session, e
}
func (s *Service) Logout(token string) error {
	_, e := s.Store.DB.Exec("DELETE FROM sessions WHERE token_hash=?", digest(token))
	return e
}
func (s *Service) ChangePassword(userID int64, current, next string) error {
	var hash, name string
	e := s.Store.DB.QueryRow("SELECT username,password_hash FROM users WHERE id=?", userID).Scan(&name, &hash)
	if e != nil {
		return e
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(current)) != nil {
		return ErrCredentials
	}
	if e = Validate(name, next); e != nil {
		return e
	}
	hash, e = Hash(next)
	if e != nil {
		return e
	}
	tx, e := s.Store.DB.Begin()
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if _, e = tx.Exec("UPDATE users SET password_hash=? WHERE id=?", hash, userID); e != nil {
		return e
	}
	if _, e = tx.Exec("DELETE FROM sessions WHERE user_id=?", userID); e != nil {
		return e
	}
	return tx.Commit()
}
func (s *Service) SetUser(id int64, role string, disabled bool) error {
	if role != "admin" && role != "viewer" {
		return errors.New("无效角色")
	}
	tx, e := s.Store.DB.Begin()
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var oldRole string
	var oldDisabled bool
	if e = tx.QueryRow("SELECT role,disabled FROM users WHERE id=?", id).Scan(&oldRole, &oldDisabled); e != nil {
		return e
	}
	if oldRole == "admin" && !oldDisabled && (role != "admin" || disabled) {
		var count int
		if e = tx.QueryRow("SELECT count(*) FROM users WHERE role='admin' AND disabled=0").Scan(&count); e != nil {
			return e
		}
		if count <= 1 {
			return errors.New("必须保留至少一名启用的管理员")
		}
	}
	if _, e = tx.Exec("UPDATE users SET role=?,disabled=? WHERE id=?", role, disabled, id); e != nil {
		return e
	}
	if _, e = tx.Exec("DELETE FROM sessions WHERE user_id=?", id); e != nil {
		return e
	}
	return tx.Commit()
}

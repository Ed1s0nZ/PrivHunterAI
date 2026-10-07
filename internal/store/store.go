package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"yuequanScan/internal/domain"
)

type Store struct{ DB *sql.DB }

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	f.Close()
	if err = os.Chmod(path, 0600); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", "file:"+path+"?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{DB: db}
	if err = s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error { return s.DB.Close() }
func (s *Store) UserCount() (int, error) {
	var n int
	err := s.DB.QueryRow("SELECT count(*) FROM users").Scan(&n)
	return n, err
}
func (s *Store) CreateUser(name, hash, role string) (int64, error) {
	r, e := s.DB.Exec("INSERT INTO users(username,password_hash,role) VALUES(?,?,?)", name, hash, role)
	if e != nil {
		return 0, e
	}
	return r.LastInsertId()
}

// Bootstrap atomically permits exactly one initial administrator.
func (s *Store) Bootstrap(name, hash string) (bool, error) {
	r, e := s.DB.Exec("INSERT INTO users(username,password_hash,role) SELECT ?,?,'admin' WHERE NOT EXISTS(SELECT 1 FROM users)", name, hash)
	if e != nil {
		return false, e
	}
	n, e := r.RowsAffected()
	return n == 1, e
}
func (s *Store) UserByName(name string) (domain.User, string, error) {
	var u domain.User
	var h string
	e := s.DB.QueryRow("SELECT id,username,role,disabled,password_hash FROM users WHERE username=?", name).Scan(&u.ID, &u.Username, &u.Role, &u.Disabled, &h)
	return u, h, e
}
func (s *Store) Users() ([]domain.User, error) {
	rows, e := s.DB.Query("SELECT id,username,role,disabled FROM users ORDER BY id")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.User{}
	for rows.Next() {
		var u domain.User
		if e = rows.Scan(&u.ID, &u.Username, &u.Role, &u.Disabled); e != nil {
			return nil, e
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
func (s *Store) AddFinding(v domain.Finding) (int64, error) {
	v.CreatedAt = time.Now().UTC()
	v.Review = "pending"
	b, e := json.Marshal(v)
	if e != nil {
		return 0, e
	}
	r, e := s.DB.Exec("INSERT INTO findings(payload,result,method,url,created_at) VALUES(?,?,?,?,?)", string(b), v.Result, v.Method, v.URL, v.CreatedAt.Format(time.RFC3339Nano))
	if e != nil {
		return 0, e
	}
	return r.LastInsertId()
}

type Filter struct {
	Search, Result, Review string
	Page, Size             int
}

func (s *Store) Findings(f Filter) ([]domain.Finding, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 || f.Size > 100 {
		f.Size = 20
	}
	where := ` WHERE (?='' OR result=?) AND (?='' OR review=?) AND (?='' OR instr(lower(url),lower(?))>0)`
	args := []any{f.Result, f.Result, f.Review, f.Review, f.Search, f.Search}
	var total int
	if e := s.DB.QueryRow("SELECT count(*) FROM findings"+where, args...).Scan(&total); e != nil {
		return nil, 0, e
	}
	args = append(args, f.Size, (f.Page-1)*f.Size)
	rows, e := s.DB.Query("SELECT id,json_object('method',method,'url',url,'result',result,'reason',json_extract(payload,'$.reason'),'confidence',json_extract(payload,'$.confidence'),'createdAt',created_at),review,note FROM findings"+where+" ORDER BY id DESC LIMIT ? OFFSET ?", args...)
	if e != nil {
		return nil, 0, e
	}
	defer rows.Close()
	out := []domain.Finding{}
	for rows.Next() {
		var v domain.Finding
		var raw, review, note string
		var id int64
		if e = rows.Scan(&id, &raw, &review, &note); e != nil {
			return nil, 0, e
		}
		if e = json.Unmarshal([]byte(raw), &v); e != nil {
			return nil, 0, e
		}
		v.ID = id
		v.Review = review
		v.Note = note
		out = append(out, v)
	}
	return out, total, rows.Err()
}
func (s *Store) Review(id int64, status, note string) error {
	r, e := s.DB.Exec("UPDATE findings SET review=?,note=? WHERE id=?", status, note, id)
	if e != nil {
		return e
	}
	n, e := r.RowsAffected()
	if e == nil && n == 0 {
		return sql.ErrNoRows
	}
	return e
}
func (s *Store) DeleteFinding(id int64) error {
	_, e := s.DB.Exec("DELETE FROM findings WHERE id=?", id)
	return e
}
func (s *Store) Stats() (map[string]int, error) {
	rows, e := s.DB.Query("SELECT result,count(*) FROM findings GROUP BY result")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := map[string]int{"total": 0, "true": 0, "false": 0, "unknown": 0, "error": 0}
	for rows.Next() {
		var k string
		var n int
		if e = rows.Scan(&k, &n); e != nil {
			return nil, e
		}
		out[k] = n
		out["total"] += n
	}
	return out, rows.Err()
}
func (s *Store) Log(actor, action, target string) error {
	_, e := s.DB.Exec("INSERT INTO audit(actor,action,target,created_at) VALUES(?,?,?,?)", actor, action, target, time.Now().UTC().Format(time.RFC3339))
	return e
}
func (s *Store) Check() error {
	if e := s.DB.Ping(); e != nil {
		return fmt.Errorf("database unavailable: %w", e)
	}
	return nil
}

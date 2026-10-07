package store

import "yuequanScan/internal/domain"

type AuditFilter struct {
	Page           int
	Search, Action string
}

func (s *Store) Audit(f AuditFilter) ([]domain.Audit, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	where := ` WHERE (?='' OR action=?) AND (?='' OR instr(lower(actor||' '||target),lower(?))>0)`
	args := []any{f.Action, f.Action, f.Search, f.Search}
	var total int
	if e := s.DB.QueryRow("SELECT count(*) FROM audit"+where, args...).Scan(&total); e != nil {
		return nil, 0, e
	}
	rows, e := s.DB.Query("SELECT id,actor,action,target,created_at FROM audit"+where+" ORDER BY id DESC LIMIT 50 OFFSET ?", append(args, (f.Page-1)*50)...)
	if e != nil {
		return nil, 0, e
	}
	defer rows.Close()
	out := []domain.Audit{}
	for rows.Next() {
		var v domain.Audit
		if e = rows.Scan(&v.ID, &v.Actor, &v.Action, &v.Target, &v.CreatedAt); e != nil {
			return nil, 0, e
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
}

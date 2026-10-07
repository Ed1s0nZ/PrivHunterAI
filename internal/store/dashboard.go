package store

import (
	"net/url"
	"sort"
	"time"
)

type Trend struct {
	Date      string `json:"date"`
	Total     int    `json:"total"`
	Suspected int    `json:"suspected"`
}
type Target struct {
	Host    string `json:"host"`
	Total   int    `json:"total"`
	Pending int    `json:"pending"`
}
type Dashboard struct {
	Reviews map[string]int `json:"reviews"`
	Trend   []Trend        `json:"trend"`
	Targets []Target       `json:"targets"`
}

func (s *Store) Dashboard() (Dashboard, error) {
	out := Dashboard{Reviews: map[string]int{"pending": 0, "confirmed": 0, "false_positive": 0, "resolved": 0}, Trend: []Trend{}, Targets: []Target{}}
	location := time.FixedZone("Asia/Shanghai", 8*3600)
	now := time.Now().In(location)
	indices := map[string]int{}
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format("2006-01-02")
		indices[day] = len(out.Trend)
		out.Trend = append(out.Trend, Trend{Date: day})
	}
	rows, e := s.DB.Query("SELECT url,result,review,created_at FROM findings")
	if e != nil {
		return out, e
	}
	defer rows.Close()
	hosts := map[string]*Target{}
	for rows.Next() {
		var raw, result, review, created string
		if e = rows.Scan(&raw, &result, &review, &created); e != nil {
			return out, e
		}
		out.Reviews[review]++
		if ts, err := time.Parse(time.RFC3339Nano, created); err == nil {
			if index, ok := indices[ts.In(location).Format("2006-01-02")]; ok {
				out.Trend[index].Total++
				if result == "true" {
					out.Trend[index].Suspected++
				}
			}
		}
		if u, err := url.Parse(raw); err == nil && u.Hostname() != "" {
			host := u.Hostname()
			if hosts[host] == nil {
				hosts[host] = &Target{Host: host}
			}
			hosts[host].Total++
			if review == "pending" {
				hosts[host].Pending++
			}
		}
	}
	if e = rows.Err(); e != nil {
		return out, e
	}
	for _, v := range hosts {
		out.Targets = append(out.Targets, *v)
	}
	sort.Slice(out.Targets, func(i, j int) bool {
		if out.Targets[i].Pending == out.Targets[j].Pending {
			return out.Targets[i].Host < out.Targets[j].Host
		}
		return out.Targets[i].Pending > out.Targets[j].Pending
	})
	if len(out.Targets) > 5 {
		out.Targets = out.Targets[:5]
	}
	return out, nil
}

package httpapi

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"yuequanScan/internal/store"

	"github.com/gin-gonic/gin"
)

func safeCell(v string) string {
	if strings.HasPrefix(v, "=") || strings.HasPrefix(v, "+") || strings.HasPrefix(v, "-") || strings.HasPrefix(v, "@") || strings.HasPrefix(v, "\t") || strings.HasPrefix(v, "\r") {
		return "'" + v
	}
	return v
}
func (s *Server) export(c *gin.Context) {
	f := filter(c)
	f.Size = 100
	f.Page = 1
	format := c.DefaultQuery("format", "csv")
	if format != "csv" && format != "json" {
		fail(c, 400, "不支持的导出格式")
		return
	}
	all, err := s.Store.Export(f)
	if errors.Is(err, store.ErrExportLimit) {
		fail(c, 400, "导出限制为 10000 条 / 32 MiB，请缩小筛选范围")
		return
	}
	if err != nil {
		fail(c, 500, "导出失败")
		return
	}

	s.Store.Log(session(c).User.Username, "export", strconv.Itoa(len(all)))
	c.Header("Content-Disposition", "attachment; filename=privhunter-results."+format)
	if format == "json" {
		c.Header("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(c.Writer).Encode(all)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Writer.Write([]byte{0xef, 0xbb, 0xbf})
	w := csv.NewWriter(c.Writer)
	defer w.Flush()
	w.Write([]string{"ID", "Method", "URL", "Result", "Confidence", "Review", "Reason", "Note", "Created At"})
	for _, v := range all {
		row := []string{strconv.FormatInt(v.ID, 10), v.Method, v.URL, v.Result, v.Confidence, v.Review, v.Reason, v.Note, v.CreatedAt.Format("2006-01-02T15:04:05Z")}
		for i := range row {
			row[i] = safeCell(row[i])
		}
		if w.Write(row) != nil {
			return
		}
	}
}

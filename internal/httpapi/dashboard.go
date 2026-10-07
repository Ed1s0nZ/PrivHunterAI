package httpapi

import (
	"github.com/gin-gonic/gin"
	"os"
	"yuequanScan/internal/store"
)

func (s *Server) dashboard(c *gin.Context) {
	summary, e := s.Store.Dashboard()
	if e != nil {
		fail(c, 500, "汇总查询失败")
		return
	}
	counts, e := s.Store.Stats()
	if e != nil {
		fail(c, 500, "统计读取失败")
		return
	}
	recent, _, e := s.Store.Findings(store.Filter{Page: 1, Size: 5})
	if e != nil {
		fail(c, 500, "最近记录读取失败")
		return
	}
	result := gin.H{"summary": summary, "counts": counts, "recent": recent}
	if session(c).User.Role == "admin" {
		settings, e := s.Store.Settings()
		if e != nil {
			fail(c, 500, "就绪状态读取失败")
			return
		}
		result["readiness"] = gin.H{"scope": len(settings.Domains) > 0, "identity": len(settings.Headers) > 0, "model": settings.Endpoint != "" && settings.Model != "" && (settings.APIKey != "" || os.Getenv("PRIVHUNTER_API_KEY") != ""), "paused": settings.Paused}
		if s.ScannerStatus != nil {
			result["runtime"] = s.ScannerStatus()
		}
	}
	c.JSON(200, result)
}

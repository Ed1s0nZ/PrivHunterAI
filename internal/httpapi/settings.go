package httpapi

import (
	"errors"
	"net/url"
	"strings"
	"yuequanScan/internal/store"

	"github.com/gin-gonic/gin"
)

func (s *Server) settings(c *gin.Context) {
	v, e := s.Store.Settings()
	if e != nil {
		fail(c, 500, "设置读取失败")
		return
	}
	c.JSON(200, gin.H{"settings": v, "hasAPIKey": v.APIKey != "", "hasHeaders": len(v.Headers) > 0})
}
func (s *Server) saveSettings(c *gin.Context) {
	var input struct {
		Revision     *int64             `json:"revision"`
		Paused       bool               `json:"paused"`
		Domains      []string           `json:"domains"`
		Endpoint     string             `json:"endpoint"`
		Model        string             `json:"model"`
		APIKey       *string            `json:"apiKey"`
		Headers      *map[string]string `json:"headers"`
		SkipSuffixes []string           `json:"skipSuffixes"`
		DenyKeywords []string           `json:"denyKeywords"`
	}
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "设置格式无效")
		return
	}
	if input.Revision == nil || *input.Revision < 0 {
		fail(c, 400, "配置版本缺失，请重新加载设置")
		return
	}
	if len(input.Domains) > 100 || len(input.Model) > 200 || len(input.DenyKeywords) > 100 {
		fail(c, 400, "设置超过限制")
		return
	}
	if input.Endpoint != "" {
		u, e := url.Parse(input.Endpoint)
		if e != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || (u.Scheme != "https" && u.Scheme != "http") {
			fail(c, 400, "模型端点必须是有效 HTTP(S) URL")
			return
		}
	}
	for i, d := range input.Domains {
		d = strings.ToLower(strings.TrimSpace(d))
		if d == "" || strings.ContainsAny(d, "/:?@ ") || strings.Contains(d, "*") {
			fail(c, 400, "目标范围应为精确域名或 IP，不含协议和通配符")
			return
		}
		input.Domains[i] = d
	}
	if input.Headers != nil {
		for k, v := range *input.Headers {
			if k == "" || strings.ContainsAny(k+v, "\r\n") {
				fail(c, 400, "请求头格式无效")
				return
			}
		}
	}
	old, e := s.Store.Settings()
	if e != nil {
		fail(c, 500, "设置读取失败")
		return
	}
	old.Paused = input.Paused
	old.Domains = input.Domains
	old.Endpoint = input.Endpoint
	old.Model = input.Model
	old.SkipSuffixes = input.SkipSuffixes
	old.DenyKeywords = input.DenyKeywords
	if input.APIKey != nil {
		old.APIKey = *input.APIKey
	}
	if input.Headers != nil {
		old.Headers = *input.Headers
	}
	if !old.Paused && (len(old.Domains) == 0 || len(old.Headers) == 0) {
		fail(c, 400, "启用扫描前请设置目标范围和账号 B 请求头")
		return
	}
	if e := s.Store.SaveSettingsRevision(old, *input.Revision); e != nil {
		if errors.Is(e, store.ErrSettingsConflict) {
			fail(c, 409, "设置已被其他操作修改。请重新加载后再保存；当前草稿仍保留。")
			return
		}
		fail(c, 500, "设置保存失败")
		return
	}
	s.Store.Log(session(c).User.Username, "update_settings", "scanner")
	c.JSON(200, gin.H{"message": "设置已保存"})
}

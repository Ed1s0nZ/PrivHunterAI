package httpapi

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"yuequanScan/internal/auth"
	"yuequanScan/internal/security"
	"yuequanScan/internal/store"
)

func (s *Server) stats(c *gin.Context) {
	v, e := s.Store.Stats()
	if e != nil {
		fail(c, 500, "查询失败")
		return
	}
	c.JSON(200, v)
}
func filter(c *gin.Context) store.Filter {
	size, _ := strconv.Atoi(c.Query("size"))
	return store.Filter{Search: c.Query("search"), Result: c.Query("result"), Review: c.Query("review"), Page: page(c), Size: size}
}
func (s *Server) findings(c *gin.Context) {
	f := filter(c)
	v, total, e := s.Store.Findings(f)
	if e != nil {
		fail(c, 500, "查询失败")
		return
	}
	c.JSON(200, gin.H{"data": v, "total": total, "page": f.Page})
}
func (s *Server) review(c *gin.Context) {
	n, ok := id(c)
	if !ok {
		return
	}
	var v struct {
		Review string `json:"review"`
		Note   string `json:"note"`
	}
	if c.ShouldBindJSON(&v) != nil || len(v.Note) > 4000 {
		fail(c, 400, "无效复核内容")
		return
	}
	switch v.Review {
	case "pending", "confirmed", "false_positive", "resolved":
	default:
		fail(c, 400, "无效复核状态")
		return
	}
	if s.Store.Review(n, v.Review, security.Text(v.Note)) != nil {
		fail(c, 404, "记录不存在")
		return
	}
	s.Store.Log(session(c).User.Username, "review", strconv.FormatInt(n, 10))
	c.JSON(200, gin.H{"message": "已保存"})
}
func (s *Server) deleteFinding(c *gin.Context) {
	n, ok := id(c)
	if !ok {
		return
	}
	if s.Store.DeleteFinding(n) != nil {
		fail(c, 500, "删除失败")
		return
	}
	s.Store.Log(session(c).User.Username, "delete_finding", strconv.FormatInt(n, 10))
	c.JSON(200, gin.H{"message": "已删除"})
}
func (s *Server) users(c *gin.Context) {
	v, e := s.Store.Users()
	if e != nil {
		fail(c, 500, "查询失败")
		return
	}
	c.JSON(200, gin.H{"data": v})
}
func (s *Server) createUser(c *gin.Context) {
	var v struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if c.ShouldBindJSON(&v) != nil {
		fail(c, 400, "无效请求")
		return
	}
	if e := auth.Validate(v.Username, v.Password); e != nil {
		fail(c, 400, e.Error())
		return
	}
	if v.Role != "admin" && v.Role != "viewer" {
		fail(c, 400, "无效角色")
		return
	}
	h, e := auth.Hash(v.Password)
	if e != nil {
		fail(c, 500, "创建失败")
		return
	}
	n, e := s.Store.CreateUser(v.Username, h, v.Role)
	if e != nil {
		fail(c, 409, "用户名已存在或创建失败")
		return
	}
	s.Store.Log(session(c).User.Username, "create_user", strconv.FormatInt(n, 10))
	c.JSON(201, gin.H{"id": n})
}
func (s *Server) updateUser(c *gin.Context) {
	n, ok := id(c)
	if !ok {
		return
	}
	var v struct {
		Role     string `json:"role"`
		Disabled bool   `json:"disabled"`
	}
	if c.ShouldBindJSON(&v) != nil {
		fail(c, 400, "无效请求")
		return
	}
	if e := s.Auth.SetUser(n, v.Role, v.Disabled); e != nil {
		fail(c, 400, e.Error())
		return
	}
	s.Store.Log(session(c).User.Username, "update_user", strconv.FormatInt(n, 10))
	c.JSON(200, gin.H{"message": "已更新，会话已失效"})
}
func (s *Server) audit(c *gin.Context) {
	v, total, e := s.Store.Audit(store.AuditFilter{Page: page(c), Search: c.Query("search"), Action: c.Query("action")})
	if e != nil {
		fail(c, 500, "查询失败")
		return
	}
	c.JSON(200, gin.H{"data": v, "total": total})
}

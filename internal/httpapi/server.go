package httpapi

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"yuequanScan/internal/auth"
	"yuequanScan/internal/store"
)

type Server struct {
	ScannerStatus func() map[string]any
	Store         *store.Store
	Auth          *auth.Service
	SecureCookie  bool
	attemptsMu    sync.Mutex
	attempts      map[string]attempt
}
type attempt struct {
	Count int
	Until time.Time
}

func New(db *store.Store, secure bool) *Server {
	return &Server{Store: db, Auth: &auth.Service{Store: db}, SecureCookie: secure, attempts: map[string]attempt{}}
}
func fail(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
func (s *Server) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'")
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	})
	r.SetTrustedProxies(nil)
	a := r.Group("/api")
	a.Use(func(c *gin.Context) { c.Header("Cache-Control", "no-store"); c.Next() })
	a.GET("/auth/status", s.status)
	a.POST("/auth/setup", s.sameOrigin, s.setup)
	a.POST("/auth/login", s.sameOrigin, s.login)
	p := a.Group("")
	p.Use(s.requireSession)
	p.GET("/auth/me", s.me)
	p.POST("/auth/logout", s.logout)
	p.POST("/auth/password", s.password)
	p.GET("/stats", s.stats)
	p.GET("/dashboard", s.dashboard)
	p.GET("/findings", s.findings)
	p.GET("/findings/:id", s.finding)
	p.GET("/export", s.export)
	admin := p.Group("")
	admin.Use(s.requireAdmin)
	admin.POST("/findings/bulk-review", s.bulkReview)
	admin.PATCH("/findings/:id", s.review)
	admin.DELETE("/findings/:id", s.deleteFinding)
	admin.GET("/users", s.users)
	admin.POST("/users", s.createUser)
	admin.PATCH("/users/:id", s.updateUser)
	admin.GET("/audit", s.audit)
	admin.GET("/scanner/status", func(c *gin.Context) {
		if s.ScannerStatus == nil {
			c.JSON(200, gin.H{})
			return
		}
		c.JSON(200, s.ScannerStatus())
	})
	admin.GET("/settings", s.settings)
	admin.PUT("/settings", s.saveSettings)
	return r
}
func (s *Server) sameOrigin(c *gin.Context) {
	if origin := c.GetHeader("Origin"); origin != "" {
		u, e := url.Parse(origin)
		if e != nil || u.Host != c.Request.Host || (u.Scheme != "http" && u.Scheme != "https") {
			fail(c, 403, "请求来源不允许")
			return
		}
	}
	if site := c.GetHeader("Sec-Fetch-Site"); site == "cross-site" {
		fail(c, 403, "请求来源不允许")
		return
	}
}
func (s *Server) requireSession(c *gin.Context) {
	token, e := c.Cookie("privhunter_session")
	if e != nil {
		fail(c, 401, "请先登录")
		return
	}
	v, e := s.Auth.Authenticate(token)
	if e != nil {
		fail(c, 401, "会话已过期，请重新登录")
		return
	}
	c.Set("session", v)
	if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
		s.sameOrigin(c)
		if c.IsAborted() {
			return
		}
		if subtle.ConstantTimeCompare([]byte(c.GetHeader("X-CSRF-Token")), []byte(v.CSRF)) != 1 {
			fail(c, 403, "请求校验失败")
			return
		}
	}
	c.Next()
}
func session(c *gin.Context) auth.Session { return c.MustGet("session").(auth.Session) }
func (s *Server) requireAdmin(c *gin.Context) {
	if session(c).User.Role != "admin" {
		fail(c, 403, "需要管理员权限")
		return
	}
	c.Next()
}
func id(c *gin.Context) (int64, bool) {
	n, e := strconv.ParseInt(c.Param("id"), 10, 64)
	if e != nil || n < 1 {
		fail(c, 400, "无效记录编号")
		return 0, false
	}
	return n, true
}
func page(c *gin.Context) int {
	n, _ := strconv.Atoi(c.Query("page"))
	if n < 1 {
		n = 1
	}
	return n
}
func (s *Server) status(c *gin.Context) {
	n, e := s.Store.UserCount()
	if e != nil {
		fail(c, 500, "数据库不可用")
		return
	}
	c.JSON(200, gin.H{"setupRequired": n == 0})
}
func (s *Server) me(c *gin.Context) {
	v := session(c)
	c.JSON(200, gin.H{"user": v.User, "csrf": v.CSRF})
}
func (s *Server) setup(c *gin.Context) {
	var v struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&v) != nil {
		fail(c, 400, "无效请求")
		return
	}
	if e := auth.Validate(v.Username, v.Password); e != nil {
		fail(c, 400, e.Error())
		return
	}
	h, e := auth.Hash(v.Password)
	if e != nil {
		fail(c, 500, "初始化失败")
		return
	}
	ok, e := s.Store.Bootstrap(v.Username, h)
	if e != nil {
		fail(c, 500, "初始化失败")
		return
	}
	if !ok {
		fail(c, 409, "系统已初始化")
		return
	}
	s.Store.Log(v.Username, "setup", "administrator")
	c.JSON(201, gin.H{"message": "初始化完成，请登录"})
}
func (s *Server) login(c *gin.Context) {
	var v struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&v) != nil || len(v.Password) > 72 {
		fail(c, 400, "无效请求")
		return
	}
	key := c.ClientIP()
	now := time.Now()
	s.attemptsMu.Lock()
	a := s.attempts[key]
	if now.Before(a.Until) && a.Count >= 10 {
		s.attemptsMu.Unlock()
		fail(c, 429, "尝试次数过多，请稍后再试")
		return
	}
	if !now.Before(a.Until) {
		a = attempt{Until: now.Add(5 * time.Minute)}
	}
	a.Count++
	s.attempts[key] = a
	if len(s.attempts) > 1024 {
		for k, v := range s.attempts {
			if now.After(v.Until) {
				delete(s.attempts, k)
			}
		}
	}
	s.attemptsMu.Unlock()
	token, vsession, e := s.Auth.Login(v.Username, v.Password)
	if e != nil {
		fail(c, 401, "用户名或密码错误")
		return
	}
	s.attemptsMu.Lock()
	delete(s.attempts, key)
	s.attemptsMu.Unlock()
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("privhunter_session", token, 43200, "/", "", s.SecureCookie, true)
	s.Store.Log(vsession.User.Username, "login", "session")
	c.JSON(200, gin.H{"user": vsession.User, "csrf": vsession.CSRF})
}
func (s *Server) logout(c *gin.Context) {
	token, _ := c.Cookie("privhunter_session")
	if s.Auth.Logout(token) != nil {
		fail(c, 500, "退出失败")
		return
	}
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("privhunter_session", "", -1, "/", "", s.SecureCookie, true)
	c.JSON(200, gin.H{"message": "已退出"})
}
func (s *Server) password(c *gin.Context) {
	var v struct {
		Current  string `json:"current"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&v) != nil {
		fail(c, 400, "无效请求")
		return
	}
	if e := s.Auth.ChangePassword(session(c).User.ID, v.Current, v.Password); e != nil {
		fail(c, 400, e.Error())
		return
	}
	s.Store.Log(session(c).User.Username, "password_changed", "self")
	c.JSON(200, gin.H{"message": "密码已修改，请重新登录"})
}

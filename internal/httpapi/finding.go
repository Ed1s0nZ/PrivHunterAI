package httpapi

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
)

func (s *Server) finding(c *gin.Context) {
	n, ok := id(c)
	if !ok {
		return
	}
	v, e := s.Store.Finding(n)
	if errors.Is(e, sql.ErrNoRows) {
		fail(c, 404, "记录不存在或已删除")
		return
	}
	if e != nil {
		fail(c, 500, "证据加载失败")
		return
	}
	c.JSON(200, v)
}

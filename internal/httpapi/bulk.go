package httpapi

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

func (s *Server) bulkReview(c *gin.Context) {
	var v struct {
		IDs    []int64 `json:"ids"`
		Review string  `json:"review"`
	}
	if c.ShouldBindJSON(&v) != nil || len(v.IDs) == 0 || len(v.IDs) > 100 {
		fail(c, 400, "请选择 1–100 条记录")
		return
	}
	switch v.Review {
	case "pending", "confirmed", "false_positive", "resolved":
	default:
		fail(c, 400, "无效复核状态")
		return
	}
	if e := s.Store.BulkReview(v.IDs, v.Review); e != nil {
		fail(c, 400, "批量更新失败，请刷新后重试")
		return
	}
	s.Store.Log(session(c).User.Username, "bulk_review", v.Review+":"+strconv.Itoa(len(v.IDs)))
	c.JSON(200, gin.H{"message": "批量复核已保存", "count": len(v.IDs)})
}

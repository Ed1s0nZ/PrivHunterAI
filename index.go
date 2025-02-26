package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 定义数据结构
type DataItem struct {
	Method    string `json:"method"`
	Host      string `json:"host"`
	Path      string `json:"path"`
	RespBodyA string `json:"respBodyA"`
	RespBodyB string `json:"respBodyB"`
	Result    string `json:"result"`
	Reason    string `json:"reason"`
}

var Resp []DataItem // 数据存储在全局切片中

func Index() {
	r := gin.Default()

	// 初始化数据
	// initData()

	// 提供前端静态文件服务
	r.LoadHTMLFiles("index.html") // 加载前端页面
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// 提供 API 接口
	r.GET("/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, Resp)
	})

	r.POST("/update", func(c *gin.Context) {
		var newData DataItem
		if err := c.ShouldBindJSON(&newData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		Resp = append(Resp, newData)
		c.JSON(http.StatusOK, gin.H{"message": "Data updated successfully"})
	})

	r.POST("/filter", func(c *gin.Context) {
		var filterData struct {
			Result string `json:"result"`
		}
		if err := c.ShouldBindJSON(&filterData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		filteredData := []DataItem{}
		for _, item := range Resp {
			if item.Result == filterData.Result {
				filteredData = append(filteredData, item)
			}
		}
		c.JSON(http.StatusOK, filteredData)
	})

	// 启动服务
	r.Run(":8222")
}

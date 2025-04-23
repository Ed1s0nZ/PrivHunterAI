package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var Resp []Result // 数据存储在全局切片中

func Index() {
	r := gin.Default()

	// 提供前端静态文件服务
	r.LoadHTMLFiles("index.html")   // 加载前端页面
	r.Static("/static", "./static") // 为前端静态资源提供服务

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// 分页数据接口
	r.GET("/data", func(c *gin.Context) {
		// 获取分页参数
		page, err1 := strconv.Atoi(c.Query("page"))
		pageSize, err2 := strconv.Atoi(c.Query("pageSize"))
		resultFilter := c.Query("result")

		if err1 != nil || page < 1 {
			page = 1
		}
		if err2 != nil || pageSize < 1 {
			pageSize = 10
		}

		// 应用筛选条件
		var filteredData []Result
		for _, item := range Resp {
			if resultFilter == "" || item.Result == resultFilter {
				filteredData = append(filteredData, item)
			}
		}

		// 计算分页数据
		total := len(filteredData)
		offset := (page - 1) * pageSize
		var data []Result
		if offset < total {
			if offset+pageSize > total {
				data = filteredData[offset:]
			} else {
				data = filteredData[offset : offset+pageSize]
			}
		}

		// 返回响应
		c.JSON(http.StatusOK, gin.H{
			"data":        data,
			"total":       total,
			"currentPage": page,
			"pageSize":    pageSize,
			"totalPages":  (total + pageSize - 1) / pageSize,
		})
	})

	// 统计数据接口
	r.GET("/stats", func(c *gin.Context) {
		total := len(Resp)
		vulnerable := 0
		unknown := 0
		safe := 0

		for _, item := range Resp {
			switch item.Result {
			case "true":
				vulnerable++
			case "unknown":
				unknown++
			case "false":
				safe++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"total":      total,
			"vulnerable": vulnerable,
			"unknown":    unknown,
			"safe":       safe,
		})
	})

	// 添加数据接口
	r.POST("/update", func(c *gin.Context) {
		var newData Result
		if err := c.ShouldBindJSON(&newData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		Resp = append(Resp, newData)
		c.JSON(http.StatusOK, gin.H{"message": "Data updated successfully"})
	})

	// 启动服务
	r.Run(":8222")
}

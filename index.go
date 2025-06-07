package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
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

	// 导出Excel接口
	r.GET("/export", func(c *gin.Context) {
		resultFilter := c.Query("result")

		// 应用筛选条件
		var filteredData []Result
		for _, item := range Resp {
			if resultFilter == "" || item.Result == resultFilter {
				filteredData = append(filteredData, item)
			}
		}

		// 创建Excel文件
		f := excelize.NewFile()
		defer func() {
			if err := f.Close(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
		}()

		// 创建基础单元格样式（仅包含边框）
		baseCellStyle, _ := f.NewStyle(&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "#000000", Style: 1},
				{Type: "top", Color: "#000000", Style: 1},
				{Type: "bottom", Color: "#000000", Style: 1},
				{Type: "right", Color: "#000000", Style: 1},
			},
		})

		// 创建漏洞状态的单元格样式
		vulnerableStyle, _ := f.NewStyle(&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "#000000", Style: 1},
				{Type: "top", Color: "#000000", Style: 1},
				{Type: "bottom", Color: "#000000", Style: 1},
				{Type: "right", Color: "#000000", Style: 1},
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{"#ffebee"},
				Pattern: 1,
			},
		})

		// 创建未知状态的单元格样式
		unknownStyle, _ := f.NewStyle(&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "#000000", Style: 1},
				{Type: "top", Color: "#000000", Style: 1},
				{Type: "bottom", Color: "#000000", Style: 1},
				{Type: "right", Color: "#000000", Style: 1},
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{"#fff8e1"},
				Pattern: 1,
			},
		})

		// 创建安全状态的单元格样式
		safeStyle, _ := f.NewStyle(&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "#000000", Style: 1},
				{Type: "top", Color: "#000000", Style: 1},
				{Type: "bottom", Color: "#000000", Style: 1},
				{Type: "right", Color: "#000000", Style: 1},
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{"#e8f5e9"},
				Pattern: 1,
			},
		})

		// 设置单元格样式
		headerStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Bold:  true,
				Color: "#FFFFFF",
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{"#4a90e2"},
				Pattern: 1,
			},
			Border: []excelize.Border{
				{Type: "left", Color: "#000000", Style: 1},
				{Type: "top", Color: "#000000", Style: 1},
				{Type: "bottom", Color: "#000000", Style: 1},
				{Type: "right", Color: "#000000", Style: 1},
			},
			Alignment: &excelize.Alignment{
				Horizontal: "center",
				Vertical:   "center",
			},
		})

		// 创建漏洞状态的文本换行样式
		vulnerableWrapStyle, _ := f.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{
				WrapText: true,
				Vertical: "top",
			},
			Border: []excelize.Border{
				{Type: "left", Color: "#000000", Style: 1},
				{Type: "top", Color: "#000000", Style: 1},
				{Type: "bottom", Color: "#000000", Style: 1},
				{Type: "right", Color: "#000000", Style: 1},
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{"#fff0f0"},
				Pattern: 1,
			},
		})

		// 创建未知状态的文本换行样式
		unknownWrapStyle, _ := f.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{
				WrapText: true,
				Vertical: "top",
			},
			Border: []excelize.Border{
				{Type: "left", Color: "#000000", Style: 1},
				{Type: "top", Color: "#000000", Style: 1},
				{Type: "bottom", Color: "#000000", Style: 1},
				{Type: "right", Color: "#000000", Style: 1},
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{"#fffaec"},
				Pattern: 1,
			},
		})

		// 创建安全状态的文本换行样式
		safeWrapStyle, _ := f.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{
				WrapText: true,
				Vertical: "top",
			},
			Border: []excelize.Border{
				{Type: "left", Color: "#000000", Style: 1},
				{Type: "top", Color: "#000000", Style: 1},
				{Type: "bottom", Color: "#000000", Style: 1},
				{Type: "right", Color: "#000000", Style: 1},
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{"#f0fff0"},
				Pattern: 1,
			},
		})

		// 设置表头
		sheetName := "扫描结果"
		f.SetSheetName("Sheet1", sheetName)
		headers := []string{"API地址", "请求方式", "状态", "置信度", "原因", "原始请求", "原始响应", "重放请求", "重放响应", "时间戳"}
		for i, header := range headers {
			col := string(rune('A' + i))
			f.SetCellValue(sheetName, col+"1", header)
			f.SetCellStyle(sheetName, col+"1", col+"1", headerStyle)
		}

		// 设置列宽
		f.SetColWidth(sheetName, "A", "A", 50) // API地址
		f.SetColWidth(sheetName, "B", "B", 10) // 请求方式
		f.SetColWidth(sheetName, "C", "C", 15) // 状态
		f.SetColWidth(sheetName, "D", "D", 15) // 置信度
		f.SetColWidth(sheetName, "E", "E", 40) // 原因
		f.SetColWidth(sheetName, "F", "F", 50) // 原始请求
		f.SetColWidth(sheetName, "G", "G", 50) // 原始响应
		f.SetColWidth(sheetName, "H", "H", 50) // 重放请求
		f.SetColWidth(sheetName, "I", "I", 50) // 重放响应
		f.SetColWidth(sheetName, "J", "J", 20) // 时间戳

		// 填充数据
		for i, item := range filteredData {
			row := i + 2 // 从第2行开始(表头是第1行)

			// 状态显示文本与样式
			var statusText string
			var styleID int
			if item.Result == "true" {
				statusText = "漏洞"
				styleID = vulnerableStyle
			} else if item.Result == "unknown" {
				statusText = "未知"
				styleID = unknownStyle
			} else {
				statusText = "安全"
				styleID = safeStyle
			}

			f.SetCellValue(sheetName, "A"+strconv.Itoa(row), item.Url)
			f.SetCellValue(sheetName, "B"+strconv.Itoa(row), item.Method)
			f.SetCellValue(sheetName, "C"+strconv.Itoa(row), statusText)
			f.SetCellValue(sheetName, "D"+strconv.Itoa(row), item.Confidence)
			f.SetCellValue(sheetName, "E"+strconv.Itoa(row), item.Reason)
			f.SetCellValue(sheetName, "F"+strconv.Itoa(row), item.RequestA)
			f.SetCellValue(sheetName, "G"+strconv.Itoa(row), item.RespBodyA)
			f.SetCellValue(sheetName, "H"+strconv.Itoa(row), item.RequestB)
			f.SetCellValue(sheetName, "I"+strconv.Itoa(row), item.RespBodyB)
			f.SetCellValue(sheetName, "J"+strconv.Itoa(row), item.Timestamp)

			// 设置普通单元格的基础样式
			f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "A"+strconv.Itoa(row), baseCellStyle)
			f.SetCellStyle(sheetName, "B"+strconv.Itoa(row), "B"+strconv.Itoa(row), baseCellStyle)
			f.SetCellStyle(sheetName, "C"+strconv.Itoa(row), "C"+strconv.Itoa(row), styleID)
			f.SetCellStyle(sheetName, "D"+strconv.Itoa(row), "D"+strconv.Itoa(row), baseCellStyle)
			f.SetCellStyle(sheetName, "J"+strconv.Itoa(row), "J"+strconv.Itoa(row), baseCellStyle)

			// 根据状态，为文本换行的单元格设置对应的背景色
			if item.Result == "true" {
				f.SetCellStyle(sheetName, "E"+strconv.Itoa(row), "E"+strconv.Itoa(row), vulnerableWrapStyle)
				f.SetCellStyle(sheetName, "F"+strconv.Itoa(row), "F"+strconv.Itoa(row), vulnerableWrapStyle)
				f.SetCellStyle(sheetName, "G"+strconv.Itoa(row), "G"+strconv.Itoa(row), vulnerableWrapStyle)
				f.SetCellStyle(sheetName, "H"+strconv.Itoa(row), "H"+strconv.Itoa(row), vulnerableWrapStyle)
				f.SetCellStyle(sheetName, "I"+strconv.Itoa(row), "I"+strconv.Itoa(row), vulnerableWrapStyle)
			} else if item.Result == "unknown" {
				f.SetCellStyle(sheetName, "E"+strconv.Itoa(row), "E"+strconv.Itoa(row), unknownWrapStyle)
				f.SetCellStyle(sheetName, "F"+strconv.Itoa(row), "F"+strconv.Itoa(row), unknownWrapStyle)
				f.SetCellStyle(sheetName, "G"+strconv.Itoa(row), "G"+strconv.Itoa(row), unknownWrapStyle)
				f.SetCellStyle(sheetName, "H"+strconv.Itoa(row), "H"+strconv.Itoa(row), unknownWrapStyle)
				f.SetCellStyle(sheetName, "I"+strconv.Itoa(row), "I"+strconv.Itoa(row), unknownWrapStyle)
			} else {
				f.SetCellStyle(sheetName, "E"+strconv.Itoa(row), "E"+strconv.Itoa(row), safeWrapStyle)
				f.SetCellStyle(sheetName, "F"+strconv.Itoa(row), "F"+strconv.Itoa(row), safeWrapStyle)
				f.SetCellStyle(sheetName, "G"+strconv.Itoa(row), "G"+strconv.Itoa(row), safeWrapStyle)
				f.SetCellStyle(sheetName, "H"+strconv.Itoa(row), "H"+strconv.Itoa(row), safeWrapStyle)
				f.SetCellStyle(sheetName, "I"+strconv.Itoa(row), "I"+strconv.Itoa(row), safeWrapStyle)
			}

			// 设置行高，确保内容能够正常显示
			f.SetRowHeight(sheetName, row, 100)
		}

		// 生成Excel文件
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=PrivHunterAI-扫描结果-"+time.Now().Format("2006-01-02-15-04-05")+".xlsx")
		c.Header("Content-Transfer-Encoding", "binary")

		// 将Excel文件写入响应
		if err := f.Write(c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	})

	// 启动服务
	r.Run(":8222")
}

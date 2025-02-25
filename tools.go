package main

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
	colorBlue   = "\033[34m" // 蓝色：白名单接口
	colorCyan   = "\033[36m"
)

// 输出
func PrintYuequan(yuequan string, method string, url string, reason string) string {
	switch yuequan {
	case "true":
		return fmt.Sprintf("%s[+] %s %s %s  %s 原因:%s %s%s%s\n", colorRed, colorReset, method, url, colorCyan, reason, colorRed, "[可能存在越权/未授权漏洞]", colorReset)
	case "false":
		return fmt.Sprintf("%s[-] %s %s %s  %s 原因:%s %s%s%s\n", colorGreen, colorReset, method, url, colorCyan, reason, colorGreen, "[不存在越权/未授权漏洞]", colorReset)
	case "unknown":
		return fmt.Sprintf("%s[*] %s %s %s  %s 原因:%s %s%s%s\n", colorYellow, colorReset, method, url, colorCyan, reason, colorYellow, "[不确定是否存在漏洞]", colorReset)
	default:
		return fmt.Sprintf("%s[-] %s %s %s  %s 原因:%s %s%s%s\n", colorBlue, colorReset, method, url, colorCyan, reason, colorBlue, "[未进行扫描]", colorReset)
	}
}

// 解析数据的函数
func parseResponse(data string) (string, error) {
	var jsonData string

	// 检查数据是否包含 Markdown 代码块
	if strings.Contains(data, "```json") {
		// 使用正则表达式提取 JSON 数据
		re := regexp.MustCompile("(?s)```json\\s*(\\{.*?\\})\\s*```")
		matches := re.FindStringSubmatch(data)
		if len(matches) < 2 {
			return "", fmt.Errorf("未找到 JSON 数据")
		}
		jsonData = matches[1] // 提取 JSON 数据
	} else {
		// 数据是普通 JSON 格式
		jsonData = data
	}
	return jsonData, nil

}

func isNotSuffix(s string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return false
		}
	}
	return true
}

// 扫描白名单
func containsString(target string, slice []string) bool {
	for _, s := range slice {
		if strings.Contains(strings.ToLower(target), strings.ToLower(s)) {
			// log.Println(target)
			return true
		}
	}

	return false
}

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
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

// 字符串大于600 会被截断
func TruncateString(s string) string {
	// 将字符串转换为 rune 切片
	runeSlice := []rune(s)

	// 获取 rune 切片的长度
	length := len(runeSlice)

	// 如果长度小于或等于600 runes，直接返回原字符串
	if length <= 600 {
		return s
	}

	// 截取前300 runes 和后300 runes
	start := runeSlice[:300]
	end := runeSlice[length-300:]

	// 将截取的部分和省略号拼接起来
	return fmt.Sprintf("%s...%s", string(start), string(end))
}

// 扫描接口白名单、匹配相应包关键字
func MatchString(keywords []string, str string) bool {
	switch len(keywords) {
	case 0:
		return false
	case 1:
		return strings.Contains(str, keywords[0])
	default:
		pattern := GeneratePattern(keywords)
		matched, err := regexp.MatchString(pattern, str)
		if err != nil {
			panic(err)
		}
		return matched
	}
}
func GeneratePattern(keywords []string) string {
	var pattern strings.Builder
	pattern.WriteString("(")
	pattern.WriteString(strings.Join(keywords, "|"))
	pattern.WriteString(")")
	return pattern.String()
}

type HTTPRequest struct {
	Method string `json:"Method"`
	URL    struct {
		Path     string `json:"Path"`
		RawQuery string `json:"RawQuery"`
	} `json:"URL"`
	Proto  string              `json:"Proto"`
	Header map[string][]string `json:"Header"`
	Body   string              `json:"Body"`
}

func generateHTTPRequest(input string) (string, error) {
	var req HTTPRequest
	err := json.Unmarshal([]byte(input), &req)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	// 构建 URL
	var urlString string
	if req.URL.Path != "" {
		urlString = req.URL.Path
	}
	if req.URL.RawQuery != "" {
		urlString += "?" + req.URL.RawQuery
	}

	// 构建请求行
	requestLine := fmt.Sprintf("%s %s %s", req.Method, urlString, req.Proto)

	// 构建 Header
	var headerLines []string
	for key, values := range req.Header {
		if strings.ToLower(key) == "content-length" {
			continue // 跳过 Content-Length
		}
		if len(values) > 0 {
			// 将多个值用逗号分隔
			headerLines = append(headerLines, fmt.Sprintf("%s: %s", key, strings.Join(values, ",")))
		}
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(req.Body)
	if err != nil {
		fmt.Println("解码错误:", err)
	}
	// 如果 Body 不为空，计算 Content-Length
	if string(decodedBytes) != "" {
		contentLength := len(string(decodedBytes))
		headerLines = append(headerLines, fmt.Sprintf("Content-Length: %d", contentLength))
	}

	// 构建最终的 HTTP 请求字符串
	var buffer bytes.Buffer
	buffer.WriteString(requestLine + "\n")
	for _, line := range headerLines {
		buffer.WriteString(line + "\n")
	}
	if req.Body != "" {
		buffer.WriteString("\n" + string(decodedBytes))
	} else {
		buffer.WriteString("\n\n")
	}
	return buffer.String(), nil
}

// Levenshtein 计算两个字符串的Levenshtein距离
func Levenshtein(a, b string) int {
	lenA := len(a)
	lenB := len(b)
	if lenA == 0 {
		return lenB
	}
	if lenB == 0 {
		return lenA
	}

	// 创建距离矩阵
	dist := make([][]int, lenA+1)
	for i := range dist {
		dist[i] = make([]int, lenB+1)
		dist[i][0] = i
	}
	for j := 0; j <= lenB; j++ {
		dist[0][j] = j
	}

	// 填充距离矩阵
	for i := 1; i <= lenA; i++ {
		for j := 1; j <= lenB; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			dist[i][j] = min(
				dist[i-1][j]+1,      // 删除
				dist[i][j-1]+1,      // 插入
				dist[i-1][j-1]+cost, // 替换
			)
		}
	}

	return dist[lenA][lenB]
}

// min 返回三个整数中的最小值
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// StringSimilarity 计算两个字符串的相似度 (0-1之间)
func StringSimilarity(a, b string) float64 {
	distance := Levenshtein(a, b)
	maxLength := len(a)
	if len(b) > maxLength {
		maxLength = len(b)
	}
	return 1.0 - float64(distance)/float64(maxLength)
}

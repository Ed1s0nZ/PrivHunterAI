package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Result struct {
	Method    string `json:"method"`
	Host      string `json:"host"` // JSON 标签用于自定义字段名
	Path      string `json:"path"`
	RespBodyA string `json:"respBodyA"`
	RespBodyB string `json:"respBodyB"`
	Result    string `json:"result"`
}

func scan() {
	for {
		time.Sleep(3 * time.Second)
		logs.Range(func(key any, value any) bool {
			// fmt.Println("The type of x is", reflect.TypeOf(value))
			var r *RequestResponseLog
			if rr, ok := value.(*RequestResponseLog); ok {
				r = rr
			} else {
				fmt.Printf("Value is not of type RequestResponseLog\n")
			}

			// fmt.Println(r)
			if r.Request.Header != nil && r.Response.Header != nil && r.Response.Body != nil {
				result, resp1, resp2, err := sendHTTPAndKimi(r) // 主要
				if err != nil {
					fmt.Println(err)
				} else {
					var resultOutput Result
					resultOutput.Method = r.Request.Method
					resultOutput.Host = r.Request.URL.Host
					resultOutput.Path = r.Request.URL.Path
					resultOutput.RespBodyA = resp1
					resultOutput.RespBodyB = resp2
					resultOutput.Result = result
					jsonData, err := json.Marshal(resultOutput)
					if err != nil {
						log.Fatalf("Error marshaling to JSON: %v", err)
					}
					log.Println(string(jsonData))
					logs.Delete(key)
					return true // 返回true继续遍历，返回false停止遍历
				}
			} else {
				// logs.Delete(key) // 不可以添加logs.Delete(key)
				return true
			}
			return true
		})
	}
}

func sendHTTPAndKimi(r *RequestResponseLog) (result string, respA string, respB string, err error) {
	resp1 := string(r.Response.Body)

	fullURL := &url.URL{
		Scheme:   r.Request.URL.Scheme,
		Host:     r.Request.URL.Host,
		Path:     r.Request.URL.Path,
		RawQuery: r.Request.URL.RawQuery,
	}

	if isNotSuffix(r.Request.URL.Path, suffixes) && !containsString(r.Response.Header.Get("Content-Type"), allowedRespHeaders) {
		req, err := http.NewRequest(r.Request.Method, fullURL.String(), strings.NewReader(string(r.Request.Body)))
		if err != nil {
			fmt.Println("创建请求失败:", err)
			return "", "", "", err
		}
		req.Header = r.Request.Header
		req.Header.Set("Cookie", cookie2)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("请求失败:", err)
			return "", "", "", err
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return "", "", "", err
		}
		// 将响应体转换为字符串
		resp2 := string(bodyBytes)
		// 输出响应体字符串
		// fmt.Println("Response1 Body:", resp1)
		// fmt.Println("Response2 Body:", resp2)
		if len(resp1+resp2) < 65535 {
			result, err := detectPrivilegeEscalation(AI, resp1, resp2)
			if err != nil {
				fmt.Println("Error:", err)
				return "", "", "", err
			}
			return result, resp1, resp2, nil
		} else {
			return "请求包太大", resp1, resp2, nil
		}

		// log.Println("Result:", result)

	}
	return "白名单后缀或白名单Content-Type接口", resp1, "", nil
}

func detectPrivilegeEscalation(AI string, resp1, resp2 string) (string, error) {
	var result string
	var err error

	switch AI {
	case "kimi":
		result, err = kimi(resp1, resp2) // 调用 kimi 检测是否越权
	case "deepseek":
		result, err = deepSeek(resp1, resp2) // 调用 deepSeek 检测是否越权
	default:
		result, err = kimi(resp1, resp2) // 默认调用 kimi 检测是否越权
	}

	if err != nil {
		return "", err
	}
	return result, nil
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

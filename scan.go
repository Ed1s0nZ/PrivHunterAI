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
	aiapis "yuequanScan/AIAPIS"
	"yuequanScan/config"
)

type Result struct {
	Method    string `json:"method"`
	Host      string `json:"host"` // JSON 标签用于自定义字段名
	Path      string `json:"path"`
	RespBodyA string `json:"respBodyA"`
	RespBodyB string `json:"respBodyB"`
	Result    string `json:"result"`
	Reason    string `json:"reason"`
}

// 扫描结果
type ScanResult struct {
	Res    string `json:"res"`
	Reason string `json:"reason"`
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

			//
			if r.Request.Header != nil && r.Response.Header != nil && r.Response.Body != nil && r.Response.StatusCode == 200 {
				// fmt.Println(r)
				result, resp1, resp2, err := sendHTTPAndKimi(r) // 主要
				if err != nil {
					logs.Delete(key)
					// fmt.Println(r)
					fmt.Println(err)
				} else {
					var resultOutput Result
					resultOutput.Method = TruncateString(r.Request.Method)
					resultOutput.Host = TruncateString(r.Request.URL.Host)
					resultOutput.Path = TruncateString(r.Request.URL.Path)
					resultOutput.RespBodyA = TruncateString(resp1)
					resultOutput.RespBodyB = TruncateString(resp2)
					//

					result1, err := parseResponse(result)
					if err != nil {
						log.Fatalf("解析失败: %v", err)
					}

					var scanR ScanResult

					err = json.Unmarshal([]byte(result1), &scanR)
					if err != nil {
						log.Println("解析 JSON 数据失败("+result+": )", err)
					} else {
						resultOutput.Result = scanR.Res
						resultOutput.Reason = scanR.Reason
						jsonData, err := json.Marshal(resultOutput)
						if err != nil {
							log.Fatalf("Error marshaling to JSON: %v", err)
						}
						log.Println(string(jsonData))
						//--- 前端
						var dataItem DataItem
						// 解析 JSON 数据到结构体
						err = json.Unmarshal([]byte(jsonData), &dataItem)
						if err != nil {
							log.Fatalf("Error parsing JSON: %v", err)
						}
						// 打印解析后的结构体内容
						// fmt.Printf("Parsed DataItem: %+v\n", dataItem)
						// if dataItem.RespBodyB{

						// }
						if dataItem.Result != "white" {
							Resp = append(Resp, dataItem)
						}

						//---
						fmt.Println(PrintYuequan(resultOutput.Result, resultOutput.Method, resultOutput.Host+resultOutput.Path, resultOutput.Reason))
						logs.Delete(key)
						return true // 返回true继续遍历，返回false停止遍历
					}
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
	jsonDataReq, err := json.Marshal(r.Request)
	if err != nil {
		fmt.Println("Error marshaling:", err)
		return "", "", "", err // 返回错误
	}
	req1 := string(jsonDataReq)

	resp1 := string(r.Response.Body)

	fullURL := &url.URL{
		Scheme:   r.Request.URL.Scheme,
		Host:     r.Request.URL.Host,
		Path:     r.Request.URL.Path,
		RawQuery: r.Request.URL.RawQuery,
	}

	if isNotSuffix(r.Request.URL.Path, config.GetConfig().Suffixes) && !containsString(r.Response.Header.Get("Content-Type"), config.GetConfig().AllowedRespHeaders) {

		req, err := http.NewRequest(r.Request.Method, fullURL.String(), strings.NewReader(string(r.Request.Body)))
		if err != nil {
			fmt.Println("创建请求失败:", err)
			return "", "", "", err
		}
		req.Header = r.Request.Header
		// 增加其他头 2025 02 27
		if config.GetConfig().Headers2 != nil {
			for key, value := range config.GetConfig().Headers2 {
				req.Header.Set(key, value)
			}
		}
		// 2025 02 27 end
		// req.Header.Set("Cookie", config.GetConfig().Cookie2)
		// log.Println(req.Header)
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

		if len(resp1+resp2) < 65535 {
			fmt.Println("Serialized JSON:", req1)

			// 初始值
			var resultDetect string
			var detectErr error
			maxRetries := 5
			for i := 0; i < maxRetries; i++ {
				resultDetect, detectErr = detectPrivilegeEscalation(config.GetConfig().AI, req1, resp1, resp2, resp.Status)
				if detectErr == nil {
					break // 成功退出循环
				}
				// 可选：增加延迟避免频繁请求
				fmt.Println("AI分析异常，重试中，异常原因：", detectErr)
				time.Sleep(5 * time.Second) // 1秒延迟
			}

			if detectErr != nil {
				fmt.Println("Error after retries:", detectErr)
				return "", "", "", detectErr
			}

			return resultDetect, resp1, resp2, nil
		} else {
			return `{"res": "white", "reason": "请求包太大"}`, resp1, resp2, nil
		}

	}
	return `{"res": "white", "reason": "白名单后缀或白名单Content-Type接口"}`, resp1, "", nil
}

func detectPrivilegeEscalation(AI string, reqA, resp1, resp2, statusB string) (string, error) {
	var result string
	var err error

	switch AI {
	case "kimi":
		result, err = aiapis.Kimi(reqA, resp1, resp2, statusB) // 调用 kimi 检测是否越权
	case "deepseek":
		result, err = aiapis.DeepSeek(reqA, resp1, resp2, statusB) // 调用 deepSeek 检测是否越权
	case "qianwen":
		result, err = aiapis.Qianwen(reqA, resp1, resp2, statusB) // 调用 qianwen 检测是否越权
	default:
		result, err = aiapis.Kimi(reqA, resp1, resp2, statusB) // 默认调用 kimi 检测是否越权
	}

	if err != nil {
		return "", err
	}
	return result, nil
}

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

	"github.com/lqqyt2423/go-mitmproxy/proxy"
)

type Result struct {
	Method     string `json:"method"`
	Url        string `json:"url"` // JSON 标签用于自定义字段名
	Reqbody    string `json:"reqbody"`
	RespBodyA  string `json:"respBodyA"`
	RespBodyB  string `json:"respBodyB"`
	Result     string `json:"result"`
	Reason     string `json:"reason"`
	Confidence string `json:"confidence"`
}

// 扫描结果
type ScanResult struct {
	Res        string `json:"res"`
	Reason     string `json:"reason"`
	Confidence string `json:"confidence"`
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
					if r.Request.URL.RawQuery != "" {
						resultOutput.Url = TruncateString(r.Request.URL.Scheme + "://" + r.Request.URL.Host + r.Request.URL.Path + "?" + r.Request.URL.RawQuery)
					} else {
						resultOutput.Url = TruncateString(r.Request.URL.Scheme + "://" + r.Request.URL.Host + r.Request.URL.Path)
					}

					resultOutput.Reqbody = TruncateString(string(r.Request.Body))
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
						resultOutput.Confidence = scanR.Confidence
						jsonData, err := json.Marshal(resultOutput)
						if err != nil {
							log.Fatalf("Error marshaling to JSON: %v", err)
						}
						log.Println(string(jsonData))
						//--- 前端
						var dataItem Result
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
						log.Println(PrintYuequan(resultOutput.Result, resultOutput.Method, resultOutput.Url, resultOutput.Reason))
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
		log.Println("ReqA:", req1)
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

		requestInfo2 := proxy.Request{
			Method: req.Method,
			URL:    req.URL,

			Proto:  req.Proto,
			Header: req.Header,
			Body:   r.Request.Body,
		}
		jsonDataReq2, err := json.Marshal(requestInfo2)
		if err != nil {
			fmt.Println("Error marshaling:", err)
			return "", "", "", err // 返回错误
		}
		req2 := string(jsonDataReq2)
		log.Println("ReqB:", req2)
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
			if !MatchString(config.GetConfig().RespBodyBWhiteList, resp2) {
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
				return `{"res": "false", "reason": "匹配到关键字，判断为无越权（未消耗AI tokens）"}`, resp1, resp2, nil
			}

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
		model := "moonshot-v1-8k"
		aiurl := "https://api.moonshot.cn/v1/chat/completions"
		apikey := config.GetConfig().APIKeys.Kimi
		result, err = aiapis.AIScan(model, aiurl, apikey, reqA, resp1, resp2, statusB) // 调用 kimi 检测是否越权
	case "deepseek":
		model := "deepseek-chat"
		aiurl := "https://api.deepseek.com/v1/chat/completions"
		apikey := config.GetConfig().APIKeys.DeepSeek
		result, err = aiapis.AIScan(model, aiurl, apikey, reqA, resp1, resp2, statusB) // 调用 kimi 检测是否越权
	case "qianwen":
		model := "qwen-turbo"
		aiurl := "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
		apikey := config.GetConfig().APIKeys.Qianwen
		result, err = aiapis.AIScan(model, aiurl, apikey, reqA, resp1, resp2, statusB) // 调用 kimi 检测是否越权
	case "hunyuan":
		model := "hunyuan-turbo"
		aiurl := "https://api.hunyuan.cloud.tencent.com/v1/chat/completions"
		apikey := config.GetConfig().APIKeys.HunYuan
		result, err = aiapis.AIScan(model, aiurl, apikey, reqA, resp1, resp2, statusB) // 调用 hunyuan 检测是否越权
	case "glm":
		model := "glm-4-air"
		aiurl := "https://open.bigmodel.cn/api/paas/v4/chat/completions"
		apikey := config.GetConfig().APIKeys.Glm
		result, err = aiapis.AIScan(model, aiurl, apikey, reqA, resp1, resp2, statusB) // 调用 hunyuan 检测是否越权
	case "gpt":
		model := "gpt-4o"
		aiurl := "https://open.bigmodel.cn/api/paas/v4/chat/completions"
		apikey := config.GetConfig().APIKeys.Gpt
		result, err = aiapis.AIScan(model, aiurl, apikey, reqA, resp1, resp2, statusB) // 调用 hunyuan 检测是否越权
	default:
		model := "moonshot-v1-8k"
		aiurl := "https://api.moonshot.cn/v1/chat/completions"
		apikey := config.GetConfig().APIKeys.Kimi
		result, err = aiapis.AIScan(model, aiurl, apikey, reqA, resp1, resp2, statusB) // 调用 kimi 检测是否越权
	}

	if err != nil {
		return "", err
	}
	return result, nil
}

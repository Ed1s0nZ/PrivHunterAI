package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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
			if r.Request.Header != nil && r.Response.Header != nil {
				sendHTTPAndKimi(r) // 主要
				logs.Delete(key)
				return true // 返回true继续遍历，返回false停止遍历
			}
			return true
		})
	}
}

func sendHTTPAndKimi(r *RequestResponseLog) {
	resp1 := string(r.Response.Body)

	fullURL := &url.URL{
		Scheme:   r.Request.URL.Scheme,
		Host:     r.Request.URL.Host,
		Path:     r.Request.URL.Path,
		RawQuery: r.Request.URL.RawQuery,
	}
	// fmt.Println(fullURL.String())
	if isNotSuffix(fullURL.String(), suffixes) {
		req, err := http.NewRequest(r.Request.Method, fullURL.String(), strings.NewReader(string(r.Request.Body)))
		if err != nil {
			fmt.Println("创建请求失败:", err)
			return
		}
		req.Header = r.Request.Header
		req.Header.Set("Cookie", cookie2)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("请求失败:", err)
			return
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
		// 将响应体转换为字符串
		resp2 := string(bodyBytes)
		// 输出响应体字符串
		fmt.Println("Response1 Body:", resp1)
		fmt.Println("Response2 Body:", resp2)
		result, err := kimi(resp1, resp2) //调用kimi检测是否越权
		if err != nil {
			fmt.Println(err)
		}
		log.Println(result)
	}

}

func isNotSuffix(s string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return false
		}
	}
	return true
}

package main

import (
	"log"

	"github.com/lqqyt2423/go-mitmproxy/proxy"
)

type RequestResponseLog struct {
	Request  proxy.Request
	Response proxy.Response
}

// MyAddon 继承了 proxy.BaseAddon
type MyAddon struct {
	proxy.BaseAddon
}

// Request 方法处理 HTTP 请求
func (a *MyAddon) Request(f *proxy.Flow) {
	// 创建并记录请求日志
	logEntry := &RequestResponseLog{
		Request: *f.Request,
	}
	// 使用 Flow ID 作为键，将请求日志存入 sync.Map
	logs.Store(f.Id, logEntry)
	// log.Printf("Request URL: %s", logEntry.Request.URL)
}

// Response 方法处理 HTTP 响应
func (a *MyAddon) Response(f *proxy.Flow) {
	// 从 sync.Map 中获取对应的日志条目
	value, ok := logs.Load(f.Id)
	if ok {
		logEntry := value.(*RequestResponseLog) // 类型断言
		logEntry.Response = *f.Response
		// 如果不需要再存储，可以删除该日志条目
		// logs.Delete(f.Id)
	} else {
		// 如果找不到对应的请求日志，可以记录错误或处理异常情况
		log.Printf("No matching request log found for response with ID: %s", f.Id)
	}
}

func mitmproxy() {
	opts := &proxy.Options{
		Addr:              ":9080",
		StreamLargeBodies: 1024 * 1024 * 5,
	}

	p, err := proxy.NewProxy(opts)
	if err != nil {
		log.Fatal(err)
	}

	p.AddAddon(&MyAddon{}) // 添加 MyAddon

	log.Fatal(p.Start())
}

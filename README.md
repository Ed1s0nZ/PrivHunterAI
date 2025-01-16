# PrivHunterAI   
通过调用Kimi API来实现的越权漏洞检测（简陋版）。

## 使用方法
1. 下载源代码；
2. 编辑`config.go`文件，配置`apiKey`（Kimi的API秘钥） 和`cookie2`（响应2对应的cookie），可按需配置`suffixes`（接口后缀白名单，如.js）；
3. `go build`编译项目，并运行二进制文件；
4. BurpSuite 挂下级代理 `127.0.0.1:9080`（端口可在`mitmproxy.go` 的`Addr:":9080",` 中配置）即可开始扫描。   

## 效果
<img src="https://github.com/Ed1s0nZ/PrivHunterAI/blob/main/%E6%95%88%E6%9E%9C.png" width="500px">  


# 注意
声明：仅用于技术交流，请勿用于非法用途。

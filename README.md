# PrivHunterAI   
一个利用工作之余（摸鱼）时间花 2 小时完成的小工具，简易版支持通过被动代理调用 KIMI AI 进行越权漏洞检测，检测能力依赖 KIMI API 实现。目前功能较为基础，尚未优化输出，也未加入扫描失败后的重试机制等功能。

## 工作流程
<img src="https://github.com/Ed1s0nZ/PrivHunterAI/blob/main/%E6%B5%81%E7%A8%8B.png" width="500px">  

## 使用方法
1. 下载源代码；
2. 编辑`config.go`文件，配置`apiKey`（Kimi的API秘钥） 和`cookie2`（响应2对应的cookie），可按需配置`suffixes`（接口后缀白名单，如.js）；
3. `go build`编译项目，并运行二进制文件；
4. BurpSuite 挂下级代理 `127.0.0.1:9080`（端口可在`mitmproxy.go` 的`Addr:":9080",` 中配置）即可开始扫描。   

## 效果
<img src="https://github.com/Ed1s0nZ/PrivHunterAI/blob/main/%E6%95%88%E6%9E%9C.png" width="500px">  


# 注意
声明：仅用于技术交流，请勿用于非法用途。

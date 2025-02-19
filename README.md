# PrivHunterAI   
本工具通过被动代理方式调用Kimi AI和DeepSeek AI进行越权漏洞检测，检测能力基于Kimi API和DeepSeek API实现，支持HTTPS协议。
## 时间线
- 2025.02.06
  1. 新增DeepSeek AI引擎来检测越权；
  2. 优化流程图。
- 2025.02.18
  1. 优化输出格式；
  2. 增加扫描失败重试机制，避免出现漏扫。

## 工作流程
<img src="https://github.com/Ed1s0nZ/PrivHunterAI/blob/main/%E6%B5%81%E7%A8%8B%E5%9B%BE.png" width="800px">  

## Prompt
```
{
    "role": "你是一个AI，负责通过比较两个HTTP响应数据包来检测潜在的越权行为，并自行做出判断。",
    "inputs": {
      "responseA": "账号A请求某接口的响应。",
      "responseB": "使用账号B的Cookie重放请求的响应。"
    },
    "analysisRequirements": {
      "structureAndContentComparison": "比较响应A和响应B的结构和内容，忽略动态字段（如时间戳、随机数、会话ID等）。",
      "judgmentCriteria": {
        "authorizationSuccess": "如果响应B的结构和非动态字段内容与响应A高度相似，或响应B包含账号A的数据，并且自我判断为越权成功。",
        "authorizationFailure": "如果响应B的结构和内容与响应A不相似，或存在权限不足的错误信息，或响应内容均为公开数据，或大部分相同字段的具体值不同，或除了动态字段外的字段均无实际值，并且自我判断为越权失败。",
        "unknown": "其他情况，或无法确定是否存在越权，并且自我判断为无法确定。"
      }
    },
    "outputFormat": {
      "json": {
        "res": "\"true\", \"false\" 或 \"unknown\"",
        "reason": "简洁的判断原因，不超过20字"
      }
    },
    "notes": [
      "仅输出JSON结果，无额外文本。",
      "确保JSON格式正确，便于后续处理。",
      "保持客观，仅根据响应内容进行分析。"
    ],
    "process": [
      "接收并理解响应A和响应B。",
      "分析响应A和响应B，忽略动态字段。",
      "基于响应的结构、内容和相关性进行自我判断，包括但不限于：",
      "- 识别响应中可能的敏感数据或权限信息。",
      "- 评估响应与预期结果之间的一致性。",
      "- 确定是否存在明显的越权迹象。",
      "输出指定格式的JSON结果，包括判断和判断原因。"
    ]
  }
  
```

## 使用方法
1. 下载源代码；
2. 编辑`config.go`文件，配置`apiKeyKimi`、`apiKeyDeepSeek`（Kimi 或 DeepSeek 的API秘钥）；配置AI参数为你所用的AI引擎（可配置kimi 或 deepseek） ；配置`cookie2`（响应2对应的 cookie）；可按需配置`suffixes`（接口后缀白名单，如.js）；
3. `go build`编译项目，并运行二进制文件；
4. 首次启动后需安装证书以解析 HTTPS 流量，证书会在首次启动命令后自动生成，路径为 ~/.mitmproxy/mitmproxy-ca-cert.pem。安装步骤可参考 Python mitmproxy 文档：[About Certificates](https://docs.mitmproxy.org/stable/concepts-certificates/)。
5. BurpSuite 挂下级代理 `127.0.0.1:9080`（端口可在`mitmproxy.go` 的`Addr:":9080",` 中配置）即可开始扫描。   

## 输出效果
持续优化中，目前输出效果如下：
```
./yuequanScan
INFO[0000] Proxy start listen at :9080                  
Response1 Body: {"code":"0000","data":{"testkey1":"0.00%","testkey2":"0.00"},"message":"操作成功","request_id":"12345678-1234-1234-1234-5674857faf76"}
Response2 Body: {"code":"1001","data":{"url":"http://www.ceshi.com/api?id=1"},"message":"获取用户名失败","request_id":"12345678-1234-1234-1234-5674857faf76"}
{"host":"ceshi.com","path":"/api/test/v1/xxx/yyy/zzz","result":"{\"res\": \"false\", \"reason\": \"响应B错误信息明显\"}"}

Response1 Body: {"code":"0000","data":{"testkey1":"0.00%","testkey2":"0.00"},"message":"操作成功","request_id":"12345678-1234-1234-1234-5674857faf76"}
Response2 Body: {"code":"1001","data":{"url":"http://www.ceshi.com/api?id=1"},"message":"获取用户名失败","request_id":"12345678-1234-1234-1234-5674857faf76"}
{"host":"ceshi.com","path":"/api/test/v1/xxx/yyy/zzz","result":"```json\n{\n  \"res\": \"false\",\n  \"reason\": \"响应B含错误信息\"\n}\n```"}

Response1 Body: {"code":"0000","data":{"testkey1":"0.00%","testkey2":"0.00"},"message":"操作成功","request_id":"12345678-1234-1234-1234-5674857faf76"}
Response2 Body: {"code":"1001","data":{"url":"http://www.ceshi.com/api?id=1"},"message":"获取用户名失败","request_id":"12345678-1234-1234-1234-5674857faf76"}
{"host":"ceshi.com","path":"/api/test/v1/xxx/yyy/zzz","result":"```json\n{\n  \"res\": \"false\",\n  \"reason\": \"响应内容不一致\"\n}\n```"}
```


# 注意
声明：仅用于技术交流，请勿用于非法用途。

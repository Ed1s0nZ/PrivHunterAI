# PrivHunterAI   
本工具通过被动代理方式调用Kimi、DeepSeek和通义千问AI，实现越权漏洞检测。检测能力基于对应AI引擎的API实现，且支持HTTPS协议。
## 时间线
- 2025.02.06
  1. ⭐️新增DeepSeek AI引擎来检测越权；
  2. 优化流程图。
- 2025.02.18
  1. 优化输出格式；
  2. ⭐️新增扫描失败重试机制，避免出现漏扫；
  3. ⭐️新增响应Content-Type白名单，静态文件不扫描；
  4. ⭐️新增限制每次扫描向AI请求的最大字节，避免因请求包过大导致扫描失败。
- 2025.02.25 -02.26
  1. 优化文件目录结构，优化终端输出；
  2. ⭐️新增对URL的分析（初步判断是否可能是无需数据鉴权的公共接口）；
  3. ⭐️新增通义千问 AI引擎来检测越权;
  4. ⭐️新增前端结果展示功能。

## 工作流程
<img src="https://github.com/Ed1s0nZ/PrivHunterAI/blob/main/img/%E6%B5%81%E7%A8%8B%E5%9B%BE.png" width="800px">  

## Prompt
```
{"role": "你是一个AI，负责通过比较两个HTTP响应数据包来检测潜在的越权行为，并自行做出判断。",
    "inputs": {
        "url":"请求的url",
        "responseA": "账号A请求url的响应。",
        "responseB": "使用账号B的Cookie重放请求的响应。"
    },
    "analysisRequirements": {
      "structureAndContentComparison": "首先分析url的特征（但是url不作为主要判断因素），判断是否可能是无需数据鉴权的公共接口；然后比较响应A和响应B的结构和内容，忽略动态字段（如时间戳、随机数、会话ID等）。",
      "judgmentCriteria": {
        "authorizationSuccess（true）": "如果url不太可能是无需数据鉴权的公共接口，且响应B的结构和非动态字段内容与响应A高度相似；或响应B包含账号A的数据，并且自我判断为越权成功。",
        "authorizationFailure（false）": "如果url大概率是无需数据鉴权的公共接口，或响应B的结构和内容与响应A不相似，或存在权限不足的错误信息，或响应内容均为公开数据，或大部分相同字段的具体值不同，或除了动态字段外的字段均无实际值，并且自我判断为越权失败。",
        "unknown": "其他情况，或无法确定是否存在越权，并且自我判断为无法确定。"
      }
    },
    "outputFormat": {
      "json": {
        "res": "\"true\", \"false\" 或 \"unknown\"",
        "reason": "清晰的判断原因，总体不超过50字。"
      }
    },
    "notes": [
      "仅输出 JSON 格式的结果，不添加任何额外文本或解释。",
      "确保JSON格式正确，便于后续处理。",
      "保持客观，仅根据响应内容进行分析。"
    ],
    "process": [
      "接收并理解url、响应A和响应B。",
      "分析url、响应A和响应B，忽略动态字段。",
      "基于url、响应的结构、内容和相关性进行自我判断，包括但不限于：",
      "- 识别url的特征，判断是否可能是无需数据鉴权的公共接口。",
      "- 识别响应中可能的敏感数据或权限信息。",
      "- 评估响应与预期结果之间的一致性。",
      "- 根据url分析及响应的分析确定是否存在明显的越权迹象。",
      "输出指定格式的JSON结果，包括判断和判断原因。"
    ]
  }  
```

## 使用方法
1. 下载源代码；
2. 编辑根目录下的`config.json`文件，配置`AI`和对应的`apiKeys`（只需要配置一个即可）；（AI的值可配置qianwen、kimi 或 deepseek） ；
3. 配置`cookie2`（响应2对应的 cookie）；可按需配置`suffixes`、`allowedRespHeaders`（接口后缀白名单，如.js）；
4. 执行`go build`编译项目，并运行二进制文件；
5. 首次启动后需安装证书以解析 HTTPS 流量，证书会在首次启动命令后自动生成，路径为 ~/.mitmproxy/mitmproxy-ca-cert.pem。安装步骤可参考 Python mitmproxy 文档：[About Certificates](https://docs.mitmproxy.org/stable/concepts-certificates/)。
6. BurpSuite 挂下级代理 `127.0.0.1:9080`（端口可在`mitmproxy.go` 的`Addr:":9080",` 中配置）即可开始扫描；
7. 终端和web界面均可查看扫描结果，前端查看结果请访问`127.0.0.1:8222` 。

## 输出效果
持续优化中，目前输出效果如下：

1. 终端输出：
<img src="https://github.com/Ed1s0nZ/PrivHunterAI/blob/main/img/%E6%95%88%E6%9E%9C.png" width="800px">  

2. 前端输出（访问127.0.0.1:8222）：
<img src="https://github.com/Ed1s0nZ/PrivHunterAI/blob/main/img/%E5%89%8D%E7%AB%AF%E7%BB%93%E6%9E%9C%E5%B1%95%E7%A4%BA.png" width="800px">  


# 注意
声明：仅用于技术交流，请勿用于非法用途。

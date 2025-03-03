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
- 2025.02.25 -02.27
  1. 优化文件目录结构，优化终端输出；
  2. ⭐️新增对URL的分析（初步判断是否可能是无需数据鉴权的公共接口）；
  3. ⭐️新增通义千问 AI引擎来检测越权;
  4. ⭐️新增前端结果展示功能。
  5. ⭐️新增针对请求B添加其他headers的功能（适配有些鉴权不在cookie中做的场景）。
- 2025.03.01
  1. 优化Prompt，降低误报率；
  2. 优化重试机制，重试会提示类似:`AI分析异常，重试中，异常原因： API returned 401: {"code":"InvalidApiKey","message":"Invalid API-key provided.","request_id":"xxxxx"}`，每10秒重试一次，重试5次失败后放弃重试（避免无限重试）。
- 2025.03.03
  1. 🔍成本优化：在调用 AI 判断越权前，新增鉴权关键字（如 “暂无查询权限”“权限不足” 等）过滤环节，若匹配到关键字则直接输出未越权结果，节省 AI tokens 花销，提升资源利用效率。

## 工作流程
<img src="https://github.com/Ed1s0nZ/PrivHunterAI/blob/main/img/%E6%B5%81%E7%A8%8B.png" width="800px">  

## Prompt
```
{
  "role": "你是一个AI，负责通过比较两个HTTP响应数据包来检测潜在的越权行为，并自行做出判断。",
  "inputs": {
    "reqA": "原始请求A",
    "responseA": "账号A请求URL的响应。",
    "responseB": "使用账号B的Cookie（也可能是token等其他参数）重放请求的响应。",
    "statusB": "账号B重放请求的请求状态码。",
    "dynamicFields": ["timestamp", "nonce", "session_id", "uuid", "request_id"]
  },
  "analysisRequirements": {
    "structureAndContentComparison": {
      "urlAnalysis": "结合原始请求A和响应A分析，判断是否可能是无需数据鉴权的公共接口（不作为主要判断依据）。",
      "responseComparison": "比较响应A和响应B的结构和内容，忽略动态字段（如时间戳、随机数、会话ID、X-Request-ID等），并进行语义匹配。",
      "httpStatusCode": "对比HTTP状态码：403/401直接判定越权失败（false），500标记为未知（unknown），200需进一步分析。",
      "similarityAnalysis": "使用字段对比和文本相似度计算（Levenshtein/Jaccard）评估内容相似度。",
      "errorKeywords": "检查responseB是否包含 'Access Denied'、'Permission Denied'、'403 Forbidden' 等错误信息，若有，则判定越权失败。",
      "emptyResponseHandling": "如果responseB返回null、[]、{}或HTTP 204，且responseA有数据，判定为权限受限（false）。",
      "sensitiveDataDetection": "如果responseB包含responseA的敏感数据（如user_id、email、balance），判定为越权成功（true）。",
      "consistencyCheck": "如果responseB和responseA结构一致但关键数据不同，判定可能是权限控制正确（false）。"
    },
    "judgmentCriteria": {
      "authorizationSuccess (true)": "如果不是公共接口，且responseB的结构和非动态字段内容与responseA高度相似，或者responseB包含responseA的敏感数据，则判定为越权成功。",
      "authorizationFailure (false)": "如果是公共接口，或者responseB的结构和responseA不相似，或者responseB明确定义权限错误（403/401/Access Denied），或者responseB为空，则判定为越权失败。",
      "unknown": "如果responseB返回500，或者responseA和responseB结构不同但没有权限相关信息，或者responseB只是部分字段匹配但无法确定影响，则判定为unknown。"
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
    "确保 JSON 格式正确，便于后续处理。",
    "保持客观，仅根据响应内容进行分析。",
    "优先使用 HTTP 状态码、错误信息和数据结构匹配进行判断。",
    "支持用户提供额外的动态字段，提高匹配准确性。"
  ],
  "process": [
    "接收并理解原始请求A、responseA和responseB。",
    "分析原始请求A，判断是否是无需鉴权的公共接口。",
    "提取并忽略动态字段（时间戳、随机数、会话ID）。",
    "对比HTTP状态码，403/401直接判定为false，500标记为unknown。",
    "检查responseB是否包含responseA的敏感数据（如user_id、email），如果有，则判定为true。",
    "检查responseB是否返回错误信息（Access Denied / Forbidden），如果有，则判定为false。",
    "计算responseA和responseB的结构相似度，并使用Levenshtein编辑距离计算文本相似度。",
    "如果responseB内容为空（null、{}、[]），判断可能是权限受限，判定为false。",
    "根据分析结果，返回JSON结果。"
  ]
}
```

## 使用方法
1. 下载源代码；
2. 编辑根目录下的`config.json`文件，配置`AI`和对应的`apiKeys`（只需要配置一个即可）；（AI的值可配置qianwen、kimi 或 deepseek） ；
3. 配置`headers2`（请求B对应的headers）；可按需配置`suffixes`、`allowedRespHeaders`（接口后缀白名单，如.js）；
4. 执行`go build`编译项目，并运行二进制文件；
5. 首次启动后需安装证书以解析 HTTPS 流量，证书会在首次启动命令后自动生成，路径为 ~/.mitmproxy/mitmproxy-ca-cert.pem。安装步骤可参考 Python mitmproxy 文档：[About Certificates](https://docs.mitmproxy.org/stable/concepts-certificates/)。
6. BurpSuite 挂下级代理 `127.0.0.1:9080`（端口可在`mitmproxy.go` 的`Addr:":9080",` 中配置）即可开始扫描；
7. 终端和web界面均可查看扫描结果，前端查看结果请访问`127.0.0.1:8222` 。

### 配置文件介绍（config.json）
| 字段             | 用途                                                                                   | 内容举例                                                                                                  |
|------------------|----------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------|
| `AI`             | 指定所使用的 AI 模型                                                                   | `qianwen`、`kimi` 或 `deepseek`                                                                                              |
| `apiKeys`        | 存储不同 AI 服务对应的 API 密钥 （填一个即可，与AI对应）                                                        | - `"kimi": "sk-xxxxxxx"`<br>- `"deepseek": "sk-yyyyyyy"`<br>- `"qianwen": "sk-zzzzzzz"`                |
| `headers2`       | 自定义请求B的 HTTP 请求头信息                                                           | - `"Cookie": "Cookie2"`<br>- `"User-Agent": "PrivHunterAI"`<br>- `"Custom-Header": "CustomValue"`    |
| `suffixes`       | 需要过滤的文件后缀名列表                                                     | `.js`、`.ico`、`.png`、`.jpg`、 `.jpeg`                                                |
| `allowedRespHeaders` | 需要过滤的 HTTP 响应头中的内容类型（`Content-Type`）                                       | `image/png`、`text/html`、`application/pdf`、`text/css`、`audio/mpeg`、`audio/wav`、`video/mp4`、`application/grpc`|
| `respBodyBWhiteList` | 鉴权关键字（如暂无查询权限、权限不足），用于初筛未越权的接口 | - `参数错误`<br>- `数据页数不正确`<br>- `文件不存在`<br>- `系统繁忙，请稍后再试`<br>- `请求参数格式不正确`<br>- `权限不足`<br>- `Token不可为空`<br>- `内部错误`|

## 输出效果
持续优化中，目前输出效果如下：

1. 终端输出：
<img src="https://github.com/Ed1s0nZ/PrivHunterAI/blob/main/img/%E6%95%88%E6%9E%9C.png" width="800px">  

2. 前端输出（访问127.0.0.1:8222）：
<img src="https://github.com/Ed1s0nZ/PrivHunterAI/blob/main/img/%E7%BB%93%E6%9E%9C%E5%B1%95%E7%A4%BA.png" width="800px">  


# 注意
声明：仅用于技术交流，请勿用于非法用途。

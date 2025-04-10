# PrivHunterAI   
一款通过被动代理方式，利用主流 AI（如 Kimi、DeepSeek、GPT 等）检测越权漏洞的工具。其核心检测功能依托相关 AI 引擎的开放 API 构建，支持 HTTPS 协议的数据传输与交互。
## 时间线
- 2025.02.18
  1. ⭐️新增扫描失败重试机制，避免出现漏扫；
  2. ⭐️新增响应Content-Type白名单，静态文件不扫描；
  3. ⭐️新增限制每次扫描向AI请求的最大字节，避免因请求包过大导致扫描失败。
- 2025.02.25 -02.27
  1. ⭐️新增对URL的分析（初步判断是否可能是无需数据鉴权的公共接口）；
  2. ⭐️新增前端结果展示功能。
  3. ⭐️新增针对请求B添加其他headers的功能（适配有些鉴权不在cookie中做的场景）。
- 2025.03.01
  1. 优化Prompt，降低误报率；
  2. 优化重试机制，重试会提示类似:`AI分析异常，重试中，异常原因： API returned 401: {"code":"InvalidApiKey","message":"Invalid API-key provided.","request_id":"xxxxx"}`，每10秒重试一次，重试5次失败后放弃重试（避免无限重试）。
- 2025.03.03
  1. 💰成本优化：在调用 AI 判断越权前，新增鉴权关键字（如 “暂无查询权限”“权限不足” 等）过滤环节，若匹配到关键字则直接输出未越权结果，节省 AI tokens 花销，提升资源利用效率;
- 2025.03.21
  1. ⭐️新增终端输出请求包记录。
  

## 工作流程
<img src="https://github.com/Ed1s0nZ/PrivHunterAI/blob/main/img/%E6%B5%81%E7%A8%8B.png" width="800px">  

## Prompt
```json
{
    "role": "你是一个专注于HTTP语义分析的越权漏洞检测专家，负责通过对比HTTP数据包，精确检测潜在的越权漏洞，并提供合理谨慎的分析结果。",
    "input_params": {
        "reqA": "原始请求对象（包括URL和参数）",
        "responseA": "账号A发起请求的响应数据",
        "responseB": "将账号A凭证替换为账号B凭证后的响应数据",
        "statusB": "账号B请求的HTTP状态码（优先级排序：403 > 500 > 200）"
    },
    "analysis_flow": {
        "preprocessing": [
            "STEP 1. **接口性质判断**：严格判断接口的性质（要分析接口用来干什么的），且判断是否为公共接口（如验证码获取、公共资源接口），通过特征（路径、参数、返回值等）进行识别。",
            "STEP 2. **动态字段处理**：根据字段内容，进行自主分析，自动过滤动态字段（如 request_id 、 timestamp 、 nonce 等）。"
        ],
        "core_logic": {
            "快速判定通道（优先级从高到低）": [
                "1. **非越权行为**：若 responseB.status_code 为403或401 → 判断为无越权行为（ false ）。",
                "2. **非越权行为**：若 responseB 为空（ null 、 [] 、 {} ），且 responseA 有数据 → 判断为无越权行为（ false ）。",
                "3. **非越权行为**：若 responseB 与 responseA 关键字段（如 data.id 、 user_id 、 account_number 等）不一致，且是账号B的相关信息 → 判断为无越权行为（ false ）。",
                "4. **越权行为**：若 responseB 与 responseA 关键字段（如 data.id 、 user_id 、 account_number 等）完全一致，且未发现账号B相关信息 → 判断为越权行为（ true ）。",
                "5. **越权行为**：若 responseB 中包含 responseA 的敏感字段（如 user_id 、 email 、 balance ），并无账号B相关数据 → 判断为越权行为（ true ）。",
                "6. **越权行为**：若 responseB 数据完全为账号A的数据 → 判断为越权行为（ true ）。",
                "7. **无法判断**：若既不符合非越权行为标准，又不符合越权行为标准 → 无法判断（ unknown ）。",
                "8. **无法判断**：若 responseB.status_code 为500，或返回异常数据（如加密或乱码） → 无法判断（ unknown ）。"
            ],
            "深度分析模式（快速通道未触发时执行）": {
                "结构对比": [
                    "a. **字段层级对比**：基于JSON Path分析嵌套结构的差异，计算字段相似度。",
                    "b. **关键字段匹配**：对比关键字段的命名、位置和值（如 data.id 、 user_id 、 account_number 等）。"
                ],
                "语义分析": [
                    "i. **数值型字段**：检查是否符合数据特征（如金额字段是否在合理范围）。",
                    "ii. **文本型字段**：检查格式和命名规范（如用户ID是否采用相同格式）。",
                    "iii. **敏感字段监测**：检查是否泄露敏感信息（如 password 、 token 等字段）。"
                ]
            }
        }
    },
    "decision_tree": {
        "true": [
            "1. 接口为非公共接口，且结构相似度 > 80% → 判断为越权（ res: true ）。",
            "2. 关键业务字段（如订单号、用户ID、手机号等）的值和层级完全一致 → 判断为越权（ res: true ）。",
            "3.  responseB 与 responseA 字段完全一致，且均为账号A的数据，未出现账号B相关信息 → 判断为越权（ res: true ）。",
            "4. 操作类接口返回 success: true 且结构相同（如修改密码成功） → 判断为越权（ res: true ）。",
            "5.  responseB 中包含账号A的敏感字段（如 password 、 token ），且未出现账号B的信息 → 判断为越权（ res: true ）。"
        ],
        "false": [
            "1. 接口为公共接口（如验证码获取、公共资源接口） → 判断为非越权（ res: false ）。",
            "2. 结构差异显著（字段缺失率 > 30%） → 判断为非越权（ res: false ）。",
            "3. 关键业务字段（如订单号、用户ID、手机号等）的值或层级不一致 → 判断为非越权（ res: false ）。"
        ],
        "unknown": [
            "1. 不满足 true 和 false 条件的情况 → 无法判断（ res: unknown ）。",
            "2. 结构部分匹配（相似度 50%-80%） → 无法判断（ res: unknown ）。",
            "3. 返回数据为系统默认值（如 false 、 null ）或为加密格式 → 无法判断（ res: unknown ）。"
        ]
    },
    "output_spec": {
        "json": {
            "res": "结果为 true 、 false 或 unknown 。",
            "reason": "提供详细的分析过程和判断依据。",
            "confidence": "结果的可信度（百分比,string类型）。"
        }
    },
    "notes": [
        "1. 判断为越权时， res 返回 true ；非越权时，返回 false ；无法判断时，返回 unknown 。",
        "2. 保持输出为JSON格式，不添加任何额外文本。",
        "3. 确保JSON格式正确，便于后续处理。",
        "4. 保持客观，仅基于响应内容进行分析。",
        "5. 支持用户提供动态字段列表或解密方式，以提高分析准确性。"
    ],
    "advanced_config": {
        "similarity_threshold": {
            "structure": 0.8,
            "content": 0.7
        },
        "sensitive_fields": [
            "password",
            "token",
            "phone",
            "id_card"
        ],
        "auto_retry": {
            "when": "检测到加密数据或非常规格式时",
            "action": "建议提供解密方式后重新检测"
        }
    }
}
```

## 使用方法
1. 下载源代码 或 Releases；
2. 编辑根目录下的`config.json`文件，配置`AI`和对应的`apiKeys`（只需要配置一个即可）；（AI的值可配置qianwen、kimi、hunyuan、gpt、glm 或 deepseek） ；
3. 配置`headers2`（请求B对应的headers）；可按需配置`suffixes`、`allowedRespHeaders`（接口后缀白名单，如.js）；
4. 执行`go build`编译项目，并运行二进制文件（如果下载的是Releases可直接运行二进制文件）；
5. 首次启动程序后需安装证书以解析 HTTPS 流量，证书会在首次启动程序后自动生成，路径为 ~/.mitmproxy/mitmproxy-ca-cert.pem(Windows 路径为%USERPROFILE%\\.mitmproxy\mitmproxy-ca-cert.pem)。安装步骤可参考 Python mitmproxy 文档：[About Certificates](https://docs.mitmproxy.org/stable/concepts-certificates/)。
6. BurpSuite 挂下级代理 `127.0.0.1:9080`（端口可在`mitmproxy.go` 的`Addr:":9080",` 中配置）即可开始扫描；
7. 终端和web界面均可查看扫描结果，前端查看结果请访问`127.0.0.1:8222` 。

### 配置文件介绍（config.json）
| 字段             | 用途                                                                                   | 内容举例                                                                                                  |
|------------------|----------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------|
| `AI`             | 指定所使用的 AI 模型                                                                   | `qianwen`、`kimi`、`hunyuan` 、`gpt`、`glm` 或 `deepseek`                                                                                              |
| `apiKeys`        | 存储不同 AI 服务对应的 API 密钥 （填一个即可，与AI对应）                                                        | - `"kimi": "sk-xxxxxxx"`<br>- `"deepseek": "sk-yyyyyyy"`<br>- `"qianwen": "sk-zzzzzzz"`<br>- `"hunyuan": "sk-aaaaaaa"`                 |
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

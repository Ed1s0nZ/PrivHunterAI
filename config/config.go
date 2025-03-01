package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// 配置结构
type Config struct {
	AI                 string            `json:"AI"`
	Headers2           map[string]string `json:"headers2"`
	Suffixes           []string          `json:"suffixes"`
	AllowedRespHeaders []string          `json:"allowedRespHeaders"`
	APIKeys            struct {
		Kimi     string `json:"kimi"`
		DeepSeek string `json:"deepseek"`
		Qianwen  string `json:"qianwen"`
	} `json:"apiKeys"`
}

// 全局配置变量
var conf Config

var Prompt = `{
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
  `

// 加载配置文件
func loadConfig(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &conf); err != nil {
		return err
	}

	return nil
}

// 获取配置
func GetConfig() Config {
	return conf
}

// 初始化配置
func init() {
	configPath := "./config.json" // 配置文件路径

	if err := loadConfig(configPath); err != nil {
		fmt.Printf("Error loading config file: %v\n", err)
		os.Exit(1)
	}
}

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// 配置结构
type Config struct {
	AI                 string   `json:"AI"`
	Cookie2            string   `json:"cookie2"`
	Suffixes           []string `json:"suffixes"`
	AllowedRespHeaders []string `json:"allowedRespHeaders"`
	APIKeys            struct {
		Kimi     string `json:"kimi"`
		DeepSeek string `json:"deepseek"`
		Qianwen  string `json:"qianwen"`
	} `json:"apiKeys"`
}

// 全局配置变量
var conf Config

var Prompt = `{"role": "你是一个AI，负责通过比较两个HTTP响应数据包来检测潜在的越权行为，并自行做出判断。",
    "inputs": {
        "url":"请求的url",
        "responseA": "账号A请求url的响应。",
        "responseB": "使用账号B的Cookie重放请求的响应。"
    },
    "analysisRequirements": {
      "structureAndContentComparison": "首先分析url的特征（但是url不作为主要判断因素），判断是否可能是无需数据鉴权的公共接口；然后比较响应A和响应B的结构和内容，忽略动态字段（如时间戳、随机数、会话ID等）。",
      "judgmentCriteria": {
        "authorizationSuccess": "如果url不太可能是无需数据鉴权的公共接口，且响应B的结构和非动态字段内容与响应A高度相似；或响应B包含账号A的数据，并且自我判断为越权成功。",
        "authorizationFailure": "如果url大概率是无需数据鉴权的公共接口，或响应B的结构和内容与响应A不相似，或存在权限不足的错误信息，或响应内容均为公开数据，或大部分相同字段的具体值不同，或除了动态字段外的字段均无实际值，并且自我判断为越权失败。",
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
  `

// 加载配置文件
func loadConfig(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
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

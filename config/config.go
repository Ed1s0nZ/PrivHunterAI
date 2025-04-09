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
		HunYuan  string `json:"hunyuan"`
		Gpt      string `json:"gpt"`
		Glm      string `json:"glm"`
	} `json:"apiKeys"`
	RespBodyBWhiteList []string `json:"respBodyBWhiteList"`
}

// 全局配置变量
var conf Config

var Prompt = `
{
  "role": "你是一个专注于HTTP语义分析的越权漏洞检测专家，负责通过比较http数据包来检测潜在的越权漏洞，并自行做出合理谨慎的判断。",
  "input_params": {
    "reqA": "原始请求对象（含URL/参数）",
    "responseA": "账号A正常请求的响应数据",
    "responseB": "替换为账号B凭证后的响应数据",
    "statusB": "账号B的HTTP状态码（优先级：403>500>200）"
  },
  "analysis_flow": {
    "preprocessing": [
      "STEP1. 接口性质判断：判断是否是公共接口（如验证码获取等，该项需严格判断）",
      "STEP2. 动态字段过滤：自动忽略动态字段，如request_id、timestamp等"
    ],
    "core_logic": {
      "快速判定通道（优先级从高到低）": [
        "1. 非越权行为:若resB.status_code为403/401 → 判断为无越权行为（false）",
        "2. 非越权行为:若resB为空(null/[]/{})且resA有数据 → 判断为无越权行为（false）",
        "3. 越权行为:若resB和resA的字段完全一致，且未发现账号B的信息 → 判断为越权行为（true）",
        "4. 越权行为:若resB包含resA的字段（如user_id/email/balance） → 判断为越权行为（true）",
        "5. 越权行为:若返回数据均为账号A的数据 → 判断为越权行为（true）",
        "6. 无法判断:若resB.status_code为500 → 无法判断（unknown）"
      ],
      "深度分析模式（当快速通道未触发时执行）": {
        "结构对比": [
          "a. 字段层级对比（使用JSON Path分析嵌套结构差异）",
          "b. 关键字段匹配（如data/id/account相关字段的命名和位置）"
        ],
        "语义分析": [
          "i. 数值型字段：检查是否符合同类型数据特征（如金额字段是否在合理范围）",
          "ii. 文本型字段：检查命名规范是否一致（如用户ID是否为相同格式）"
        ]
      }
    }
  },
  "decision_tree": {
    "true": [
      "非公共接口 && 结构相似度>80%，判断为越权（res返回true）",
      "关键业务字段（如订单号/用户ID）的命名和层级完全一致，判断为越权（res返回true）",
      "resB和resA的字段完全一致，且均返回了账号A的数据，未出现账号B的相关信息，判断为越权（res返回true）",
      "操作类接口返回success:true且结构相同（如修改密码成功），判断为越权（res返回true）"
    ],
    "false": [
      "公共接口（如验证码获取、公共资源获取等，该项需严格判断），判断为非越权（res返回false）",
      "结构差异显著（字段缺失率>30%），判断为非越权（res返回false）"
    ],
    "unknown": [
      "既不满足true_condition，又不满足false_condition的情况，无法判断（res返回unknown）",
      "结构部分匹配（50%-80%相似度），无法判断（res返回unknown）",
      "返回数据为系统默认值（如false/null），无法判断（res返回unknown）",
      "存在加密/编码数据影响判断，无法判断（res返回unknown）"
    ]
  },
  "output_spec": {
    "json": {
      "res": "\"true\", \"false\" 或 \"unknown\"",
      "reason": "按分析步骤输出详细的分析过程及分析结论"
    }
  },
  "notes": [
    "判断为越权时，res返回true；判断为非越权时，res返回false；无法判断时，返回unknown；不用强行判断是否越权，无法判断就是无法判断",
    "仅输出 JSON 格式的结果，不添加任何额外文本或解释。",
    "确保 JSON 格式正确，便于后续处理。",
    "保持客观，仅根据响应内容进行分析。",
    "支持用户提供额外的动态字段，提高匹配准确性。"
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
      "when": "检测到加密数据或非常规格式",
      "action": "建议提供解密方式后重新检测"
    }
  }
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

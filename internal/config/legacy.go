// Package config supports explicit, local migration of the old configuration.
package config

import (
	"encoding/json"
	"errors"
	"os"

	"yuequanScan/internal/domain"
)

func ImportLegacy(path string) (domain.Settings, error) {
	raw, e := os.ReadFile(path)
	if e != nil {
		return domain.Settings{}, errors.New("无法读取旧配置文件")
	}
	var old struct {
		AI       string            `json:"AI"`
		Headers  map[string]string `json:"headers2"`
		Suffixes []string          `json:"suffixes"`
		Keys     map[string]string `json:"apiKeys"`
		Keywords []string          `json:"respBodyBWhiteList"`
	}
	if json.Unmarshal(raw, &old) != nil {
		return domain.Settings{}, errors.New("旧配置格式无效")
	}
	v := domain.Settings{Paused: true, Domains: []string{}, Headers: old.Headers, SkipSuffixes: old.Suffixes, DenyKeywords: old.Keywords, APIKey: old.Keys[old.AI]}
	switch old.AI {
	case "kimi":
		v.Endpoint = "https://api.moonshot.cn/v1/chat/completions"
		v.Model = "moonshot-v1-8k"
	case "deepseek":
		v.Endpoint = "https://api.deepseek.com/v1/chat/completions"
		v.Model = "deepseek-chat"
	case "qianwen":
		v.Endpoint = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
		v.Model = "qwen-plus"
	case "hunyuan":
		v.Endpoint = "https://api.hunyuan.cloud.tencent.com/v1/chat/completions"
		v.Model = "hunyuan-turbo"
	case "glm":
		v.Endpoint = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
		v.Model = "glm-4-air"
	case "gpt":
		v.Endpoint = "https://api.openai.com/v1/chat/completions"
		v.Model = "gpt-4o"
	default:
		return v, errors.New("旧配置的模型供应商不支持")
	}
	return v, nil
}

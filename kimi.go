package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// 定义请求参数结构体
type ChatCompletionRequestKimi struct {
	Model       string        `json:"model"`
	Messages    []MessageKimi `json:"messages"`
	Temperature float32       `json:"temperature"`
}

// 定义消息结构体
type MessageKimi struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// 定义响应结果结构体
type ChatCompletionResponseKimi struct {
	Choices []ChoiceKimi `json:"choices"`
}

type ChoiceKimi struct {
	Message MessageKimi `json:"message"`
}

func kimi(respA, respB string) (string, error) {
	// 设置 API Key 和请求 URL
	apiURL := "https://api.moonshot.cn/v1/chat/completions"

	// 创建请求参数
	request := ChatCompletionRequestKimi{
		Model: "moonshot-v1-8k",
		Messages: []MessageKimi{
			{
				Role:    "system",
				Content: "{\"role\": \"你是一个AI，负责通过比较两个HTTP响应数据包来检测潜在的越权行为，并自行做出判断。\",\"inputs\": {\"responseA\": \"账号A请求某接口的响应。\",\"responseB\": \"使用账号B的Cookie重放请求的响应。\"},\"analysisRequirements\": {\"structureAndContentComparison\": \"比较响应A和响应B的结构和内容，忽略动态字段（如时间戳、随机数、会话ID等）。\",\"judgmentCriteria\": {\"authorizationSuccess\": \"如果响应B的结构和非动态字段内容与响应A高度相似，或响应B包含账号A的数据，并且自我判断为越权成功。\",\"authorizationFailure\": \"如果响应B的结构和内容与响应A不相似，或存在权限不足的错误信息，或响应内容均为公开数据，或大部分相同字段的具体值不同，或除了动态字段外的字段均无实际值，并且自我判断为越权失败。\",\"unknown\": \"其他情况，或无法确定是否存在越权，并且自我判断为无法确定。\"}},\"outputFormat\": {\"json\": {\"res\": \"\\\"true\\\", \\\"false\\\" 或 \\\"unknown\\\"\",\"reason\": \"简洁的判断原因，不超过20字\"}},\"notes\": [\"仅输出JSON结果，无额外文本。\",\"确保JSON格式正确，便于后续处理。\",\"保持客观，仅根据响应内容进行分析。\"],\"process\": [\"接收并理解响应A和响应B。\",\"分析响应A和响应B，忽略动态字段。\",\"基于响应的结构、内容和相关性进行自我判断，包括但不限于：\",\"- 识别响应中可能的敏感数据或权限信息。\",\"- 评估响应与预期结果之间的一致性。\",\"- 确定是否存在明显的越权迹象。\",\"输出指定格式的JSON结果，包括判断和判断原因。\"]}",
			},
			{
				Role:    "user",
				Content: "A:" + respA + "B:" + respB,
			},
		},
		Temperature: 0.3,
	}

	// 将请求参数编码为 JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Println("JSON 编码失败:", err)
		return "", err
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("创建请求失败:", err)
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKeyKimi))
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("发送请求失败:", err)
		return "", err
	}
	defer resp.Body.Close()

	// 解析响应结果
	var response ChatCompletionResponseKimi
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("解析响应失败:", err)
		return "", err
	}

	// 输出回复内容
	if len(response.Choices) > 0 {
		result := response.Choices[0].Message.Content
		return result, nil
	} else {
		// 处理 Choices 为空的情况，例如返回一个默认值或错误
		return "", errors.New("error: choices is empty")
	}
	// result := response.Choices[0].Message.Content
	// return result, nil

}

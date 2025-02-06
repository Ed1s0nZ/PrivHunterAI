package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	apiURLDeepSeek = "https://api.deepseek.com/v1/chat/completions" // 根据实际 API 地址修改
	apiTimeout     = 30 * time.Second
)

type ChatCompletionRequestDeepSeek struct {
	Model       string            `json:"model"`
	Messages    []MessageDeepSeek `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
}

type MessageDeepSeek struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponseDeepSeek struct {
	ID      string           `json:"id"`
	Choices []ChoiceDeepSeek `json:"choices"`
	Error   struct {
		Message string `json:"message"`
	} `json:"error"`
}

type ChoiceDeepSeek struct {
	Message      MessageDeepSeek `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

// CreateChatCompletion 发送请求到 DeepSeek API
func CreateChatCompletion(request ChatCompletionRequestDeepSeek) (*ChatCompletionResponseDeepSeek, error) {
	client := &http.Client{Timeout: apiTimeout}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %v", err)
	}

	req, err := http.NewRequest("POST", apiURLDeepSeek, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKeyDeepSeek)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, body)
	}

	var response ChatCompletionResponseDeepSeek
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response failed: %v", err)
	}

	if response.Error.Message != "" {
		return nil, fmt.Errorf("API error: %s", response.Error.Message)
	}

	return &response, nil
}

func deepSeek(respA, respB string) (string, error) {
	// 示例请求
	request := ChatCompletionRequestDeepSeek{
		Model: "deepseek-chat", // 根据实际模型名称修改
		Messages: []MessageDeepSeek{
			{
				Role:    "system",
				Content: "{\"role\": \"你是一个AI，负责通过比较两个HTTP响应数据包来检测潜在的越权行为，并自行做出判断。\",\"inputs\": {\"responseA\": \"账号A请求某接口的响应。\",\"responseB\": \"使用账号B的Cookie重放请求的响应。\"},\"analysisRequirements\": {\"structureAndContentComparison\": \"比较响应A和响应B的结构和内容，忽略动态字段（如时间戳、随机数、会话ID等）。\",\"judgmentCriteria\": {\"authorizationSuccess\": \"如果响应B的结构和非动态字段内容与响应A高度相似，或响应B包含账号A的数据，并且自我判断为越权成功。\",\"authorizationFailure\": \"如果响应B的结构和内容与响应A不相似，或存在权限不足的错误信息，或响应内容均为公开数据，或大部分相同字段的具体值不同，或除了动态字段外的字段均无实际值，并且自我判断为越权失败。\",\"unknown\": \"其他情况，或无法确定是否存在越权，并且自我判断为无法确定。\"}},\"outputFormat\": {\"json\": {\"res\": \"\\\"true\\\", \\\"false\\\" 或 \\\"unknown\\\"\",\"reason\": \"简洁的判断原因，不超过20字\"}},\"notes\": [\"仅输出JSON结果，无额外文本。\",\"确保JSON格式正确，便于后续处理。\",\"保持客观，仅根据响应内容进行分析。\"],\"process\": [\"接收并理解响应A和响应B。\",\"分析响应A和响应B，忽略动态字段。\",\"基于响应的结构、内容和相关性进行自我判断，包括但不限于：\",\"- 识别响应中可能的敏感数据或权限信息。\",\"- 评估响应与预期结果之间的一致性。\",\"- 确定是否存在明显的越权迹象。\",\"输出指定格式的JSON结果，包括判断和判断原因。\"]}",
			},
			{
				Role:    "user",
				Content: "A:" + respA + "B:" + respB,
			},
		},
		Temperature: 0.7,
		MaxTokens:   500,
	}

	response, err := CreateChatCompletion(request)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return "", err
	}

	if len(response.Choices) > 0 {
		// fmt.Println("Response:")
		// fmt.Println(response.Choices[0].Message.Content)
		return response.Choices[0].Message.Content, nil
	} else {
		fmt.Println("No response received")
		return "", errors.New("no response received")
	}
}



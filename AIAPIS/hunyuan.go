package aiapis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"yuequanScan/config"
)

const (
	apiURLHunYuan = "https://api.hunyuan.cloud.tencent.com/v1/chat/completions" // 根据实际 API 地址修改
)

type ChatCompletionRequestHunYuan struct {
	Model       string           `json:"model"`
	Messages    []MessageHunYuan `json:"messages"`
	Temperature float64          `json:"temperature,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
}

type MessageHunYuan struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponseHunYuan struct {
	ID      string          `json:"id"`
	Choices []ChoiceHunYuan `json:"choices"`
	Error   struct {
		Message string `json:"message"`
	} `json:"error"`
}

type ChoiceHunYuan struct {
	Message      MessageHunYuan `json:"message"`
	FinishReason string         `json:"finish_reason"`
}

// CreateChatCompletion 发送请求到 DeepSeek API
func CreateChatCompletionHunYuan(request ChatCompletionRequestHunYuan) (*ChatCompletionResponseHunYuan, error) {
	client := &http.Client{Timeout: apiTimeout}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %v", err)
	}

	req, err := http.NewRequest("POST", apiURLHunYuan, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.GetConfig().APIKeys.HunYuan)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, body)
	}

	var response ChatCompletionResponseHunYuan
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response failed: %v", err)
	}

	if response.Error.Message != "" {
		return nil, fmt.Errorf("API error: %s", response.Error.Message)
	}

	return &response, nil
}

func HunYuan(reqA, respA, respB, statusB string) (string, error) {
	// 示例请求
	request := ChatCompletionRequestHunYuan{
		Model: "hunyuan-turbo", // 根据实际模型名称修改
		Messages: []MessageHunYuan{
			{
				Role:    "system",
				Content: config.Prompt,
			},
			{
				Role:    "user",
				Content: "reqA:" + reqA + "\n" + "responseA:" + respA + "\n" + "responseB:" + respB + "\n" + "statusB:" + statusB,
			},
		},
		Temperature: 0.7,
		MaxTokens:   500,
	}

	response, err := CreateChatCompletionHunYuan(request)
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

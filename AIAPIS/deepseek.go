package aiapis

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

func DeepSeek(url, respA, respB string) (string, error) {
	// 示例请求
	request := ChatCompletionRequestDeepSeek{
		Model: "deepseek-chat", // 根据实际模型名称修改
		Messages: []MessageDeepSeek{
			{
				Role:    "system",
				Content: prompt,
			},
			{
				Role:    "user",
				Content: "url:" + url + "A:" + respA + "B:" + respB,
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

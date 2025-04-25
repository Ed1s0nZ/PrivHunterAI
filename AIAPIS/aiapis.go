package aiapis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
	"yuequanScan/config"
)

const (
	apiTimeout = 30 * time.Second
)

type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Error   struct {
		Message string `json:"message"`
	} `json:"error"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// CreateChatCompletion 发送请求到  API
func CreateChatCompletion(request ChatCompletionRequest, aiurl string, aiapikey string) (*ChatCompletionResponse, error) {
	client := &http.Client{Timeout: apiTimeout}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %v", err)
	}

	req, err := http.NewRequest("POST", aiurl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+aiapikey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, body)
	}

	var response ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response failed: %v", err)
	}

	if response.Error.Message != "" {
		return nil, fmt.Errorf("API error: %s", response.Error.Message)
	}

	return &response, nil
}

func AIScan(model, aiurl, apikey, reqA, respA, respB, statusB string) (string, error) {
	input := map[string]string{
		"reqA":      reqA,
		"responseA": respA,
		"responseB": respB,
		"statusB":   statusB,
	}

	inputBytes, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling input: %v\n", err)
		return "", err
	}
	fmt.Println(string(inputBytes))
	request := ChatCompletionRequest{
		Model: model, // 根据实际模型名称修改
		Messages: []Message{
			{
				Role:    "system",
				Content: config.Prompt,
			},
			{
				Role:    "user",
				Content: string(inputBytes),
			},
		},
		Temperature: 0.2,
		MaxTokens:   2500,
	}

	response, err := CreateChatCompletion(request, aiurl, apikey)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return "", err
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	} else {
		fmt.Println("No response received")
		return "", errors.New("no response received")
	}
}

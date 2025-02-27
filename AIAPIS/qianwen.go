package aiapis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"yuequanScan/config"
)

// 通义千问API配置
const (
	API_URL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
)

// 请求结构体
type QWenRequest struct {
	Model      string        `json:"model"`
	Input      RequestInput  `json:"input"`
	Parameters RequestParams `json:"parameters"`
}

type RequestInput struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestParams struct {
	ResultFormat string `json:"result_format"`
}

// 响应结构体
type QWenResponse struct {
	Output struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	} `json:"output"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
}

// 调用通义千问
func Qianwen(url, respA, respB string) (string, error) {
	requestBody := QWenRequest{
		Model: "qwen-turbo", // 可根据需要切换模型
		Input: RequestInput{
			Messages: []Message{
				{
					Role:    "system",
					Content: config.Prompt,
				},
				{
					Role:    "user",
					Content: "url:" + url + "A:" + respA + "B:" + respB,
				},
			},
		},
		Parameters: RequestParams{
			ResultFormat: "message",
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("marshal request failed: %v", err)
	}

	req, err := http.NewRequest("POST", API_URL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request failed: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.GetConfig().APIKeys.Qianwen)
	req.Header.Set("X-DashScope-Request-ID", fmt.Sprintf("%d", time.Now().UnixNano()))
	req.Header.Set("X-DashScope-SSEService", "simple")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var response QWenResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode response failed: %v", err)
	}

	if len(response.Output.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return response.Output.Choices[0].Message.Content, nil
}

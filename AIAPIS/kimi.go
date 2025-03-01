package aiapis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"yuequanScan/config"
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

func Kimi(reqA, respA, respB, statusB string) (string, error) {
	// 设置 API Key 和请求 URL
	apiURL := "https://api.moonshot.cn/v1/chat/completions"

	// 创建请求参数
	request := ChatCompletionRequestKimi{
		Model: "moonshot-v1-8k",
		Messages: []MessageKimi{
			{
				Role:    "system",
				Content: config.Prompt,
			},
			{
				Role:    "user",
				Content: "reqA:" + reqA + "\n" + "responseA:" + respA + "\n" + "responseB:" + respB + "\n" + "statusB:" + statusB,
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.GetConfig().APIKeys.Kimi))
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

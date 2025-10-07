package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type DeepSeekClient struct {
	APIKey  string
	BaseURL string
}

type DeepSeekRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
	Stream    bool      `json:"stream"`
}

type DeepSeekResponse struct {
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

func NewDeepSeekClient(apiKey string) *DeepSeekClient {
	return &DeepSeekClient{
		APIKey:  apiKey,
		BaseURL: "https://api.deepseek.com/chat/completions",
	}
}

func (c *DeepSeekClient) Ask(question string) (string, error) {
	prompt := `Ты - полезный ассистент. Отвечай на русском языке вежливо и информативно.
Будь точным в ответах и предлагай полезные советы.

Вопрос: ` + question

	requestBody := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: 2000,
		Stream:    false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Printf("DeepSeek API Response Status: %d\n", resp.StatusCode)
	if resp.StatusCode != 200 {
		fmt.Printf("DeepSeek API Response Body: %s\n", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 401 {
			return "", fmt.Errorf("ошибка авторизации: неверный API ключ")
		} else if resp.StatusCode == 429 {
			return "", fmt.Errorf("превышен лимит запросов, попробуйте позже")
		} else if resp.StatusCode == 402 {
			return "", fmt.Errorf("недостаточно средств на счету")
		}
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var response DeepSeekResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	answer := response.Choices[0].Message.Content
	answer = strings.TrimSpace(answer)

	return answer, nil
}

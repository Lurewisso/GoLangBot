package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OpenRouterClient struct {
	APIKey string
}

type ORRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type ORResponse struct {
	Choices []ORChoice `json:"choices"`
	Error   *APIError  `json:"error,omitempty"`
}

type ORChoice struct {
	Message Message `json:"message"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func NewOpenRouterClient(apiKey string) *OpenRouterClient {
	return &OpenRouterClient{
		APIKey: apiKey,
	}
}

func (c *OpenRouterClient) Ask(question string) (string, error) {
	url := "https://openrouter.ai/api/v1/chat/completions"

	requestBody := ORRequest{
		Model: "deepseek/deepseek-chat-v3.1:free",
		Messages: []Message{
			{
				Role:    "user",
				Content: question,
			},
		},
		MaxTokens: 1000,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("HTTP-Referer", "https://github.com")
	req.Header.Set("X-Title", "Telegram RAG Bot")

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

	fmt.Printf("OpenRouter API Response Status: %d\n", resp.StatusCode)
	if resp.StatusCode != 200 {
		fmt.Printf("OpenRouter API Response Body: %s\n", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 401 {
			return "", fmt.Errorf("ошибка авторизации: неверный API ключ OpenRouter")
		} else if resp.StatusCode == 429 {
			return "", fmt.Errorf("превышен лимит запросов на OpenRouter, попробуйте позже")
		}
		return "", fmt.Errorf("OpenRouter API error: %s - %s", resp.Status, string(body))
	}

	var response ORResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("OpenRouter error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	answer := response.Choices[0].Message.Content
	answer = strings.TrimSpace(answer)

	return answer, nil
}

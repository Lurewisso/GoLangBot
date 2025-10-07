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

	systemPrompt := "Ты полезный ассистент. Отвечай быстро, подробно и без повторений."

	requestBody := ORRequest{
		Model: "deepseek/deepseek-chat-v3.1:free",
		Messages: []Message{
			{
				Role:    "user",
				Content: question,
			},
			{
				Role:    "system",
				Content: systemPrompt,
			},
		},
		MaxTokens: 1500,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("не удалось составить запрос: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("ошибка при создании запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("HTTP-Referer", "https://github.com")
	req.Header.Set("X-Title", "Telegram RAG Bot")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	fmt.Printf("DeepSeek API статус ответа: %d\n", resp.StatusCode)
	if resp.StatusCode != 200 {
		fmt.Printf("DeepSeek API тело запроса: %s\n", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 401 {
			return "", fmt.Errorf("ошибка авторизации: неверный API ключ OpenRouter")
		} else if resp.StatusCode == 429 {
			return "", fmt.Errorf("превышен лимит запросов на OpenRouter, попробуйте позже")
		}
		return "", fmt.Errorf("OpenRouter API ошибка: %s - %s", resp.Status, string(body))
	}

	var response ORResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("ошибка разбора ответа: %v", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("OpenRouter ошибка: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("нет ответа от ИИ")
	}

	answer := response.Choices[0].Message.Content
	answer = strings.TrimSpace(answer)

	return answer, nil
}

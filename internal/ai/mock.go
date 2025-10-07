package ai

import "strings"

type MockClient struct {
	APIKey string
}

func NewMockClient(apiKey string) *MockClient {
	return &MockClient{
		APIKey: apiKey,
	}
}

func (c *MockClient) Ask(question string) (string, error) {
	responses := map[string]string{
		"привет":     "👋 Привет! Я тестовый бот. В реальном режиме я бы использовал AI для ответа.",
		"как дела":   "🤖 У меня всё отлично! Сейчас я работаю в тестовом режиме.",
		"rag":        "🔍 RAG (Retrieval-Augmented Generation) - это технология, которая улучшает ответы AI, используя поиск по базе знаний перед генерацией ответа.",
		"бот":        "Я телеграм бот, написанный на Go, с использованием RAG архитектуры.",
		"погода":     "🌤️ К сожалению, я не могу предоставить актуальные данные о погоде в тестовом режиме.",
		"deepseek":   "DeepSeek - это мощная AI модель для генерации текста.",
		"openrouter": "OpenRouter - это платформа, которая предоставляет доступ к различным AI моделям.",
		"команды":    "Доступные команды: /start, /help, /ask, /info",
	}

	question = strings.ToLower(question)
	for key, response := range responses {
		if strings.Contains(question, key) {
			return response, nil
		}
	}

	return "🤔 Я получил ваш вопрос: '" + question + "'. В реальном режиме я бы отправил его в AI для обработки.", nil
}

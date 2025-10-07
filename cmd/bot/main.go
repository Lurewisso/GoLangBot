package main

import (
	"GolangtgBot/internal/ai"
	"GolangtgBot/internal/bot"
	"GolangtgBot/internal/config"
	"log"
)

func main() {
	cfg := config.Load()

	if cfg.TelegramToken == "" {
		log.Fatal("необходим TELEGRAM_TOKEN")
	}

	var aiClient ai.AIClient

	switch cfg.AIProvider {
	case "openrouter":
		if cfg.OpenRouterToken == "" {
			log.Fatal("необходим OPENROUTER_TOKEN во время использования модели openrouter")
		}
		aiClient = ai.NewOpenRouterClient(cfg.OpenRouterToken)
		log.Println("Используется openrouter модель")

	case "deepseek":
		if cfg.DeepSeekToken == "" {
			log.Fatal("необходим DEEPSEEK_TOKEN во время использования модели Deepseek")
		}
		aiClient = ai.NewDeepSeekClient(cfg.DeepSeekToken)
		log.Println("Используется Deepseek модель")

	default:
		aiClient = ai.NewMockClient("")
		log.Println("Используется фейковая ИИ система!")
	}

	telegramBot, err := bot.NewBot(cfg.TelegramToken, aiClient, cfg.DebugMode)
	if err != nil {
		log.Fatalf("Ошибка при создании сессии: %v", err)
	}

	log.Printf("Бот работает с %s моделью (DeepSeek)!", cfg.AIProvider)
	telegramBot.Start()
}

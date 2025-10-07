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
		log.Fatal("TELEGRAM_TOKEN is required")
	}

	var aiClient ai.AIClient

	switch cfg.AIProvider {
	case "openrouter":
		if cfg.OpenRouterToken == "" {
			log.Fatal("OPENROUTER_TOKEN is required when using openrouter provider")
		}
		aiClient = ai.NewOpenRouterClient(cfg.OpenRouterToken)
		log.Println("Using OpenRouter AI provider")

	case "deepseek":
		if cfg.DeepSeekToken == "" {
			log.Fatal("DEEPSEEK_TOKEN is required when using deepseek provider")
		}
		aiClient = ai.NewDeepSeekClient(cfg.DeepSeekToken)
		log.Println("Using DeepSeek AI provider")

	default:
		aiClient = ai.NewMockClient("")
		log.Println("Using mock AI client")
	}

	telegramBot, err := bot.NewBot(cfg.TelegramToken, aiClient, cfg.DebugMode)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("Bot started with %s provider!", cfg.AIProvider)
	telegramBot.Start()
}

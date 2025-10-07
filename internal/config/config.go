package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken   string
	AIProvider      string
	OpenRouterToken string
	DeepSeekToken   string
	DebugMode       bool
}

func Load() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, using environment variables")
	}

	return &Config{
		TelegramToken:   getEnv("TELEGRAM_TOKEN", ""),
		AIProvider:      getEnv("AI_PROVIDER", "openrouter"),
		OpenRouterToken: getEnv("OPENROUTER_TOKEN", ""),
		DeepSeekToken:   getEnv("DEEPSEEK_TOKEN", ""),
		DebugMode:       getEnvAsBool("DEBUG_MODE", true),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

package bot

import (
	"GolangtgBot/internal/ai"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	bot       *tgbotapi.BotAPI
	aiClient  ai.AIClient
	debugMode bool
}

func NewBot(token string, aiClient ai.AIClient, debug bool) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %v", err)
	}

	bot.Debug = debug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &TelegramBot{
		bot:       bot,
		aiClient:  aiClient,
		debugMode: debug,
	}, nil
}

func (tb *TelegramBot) Start() {
	commands := []tgbotapi.BotCommand{
		{
			Command:     "/start",
			Description: "Запустить бота",
		},
		{
			Command:     "/help",
			Description: "Помощь по командам",
		},
		{
			Command:     "/ask",
			Description: "Задать вопрос AI",
		},
		{
			Command:     "/info",
			Description: "Информация о боте",
		},
	}

	config := tgbotapi.NewSetMyCommands(commands...)
	tb.bot.Request(config)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tb.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		go tb.handleUpdate(update)
	}
}

func (tb *TelegramBot) handleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		tb.handleMessage(update.Message)
	}
}

func (tb *TelegramBot) handleMessage(message *tgbotapi.Message) {
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	if message.IsCommand() {
		tb.handleCommand(message)
		return
	}

	tb.handleAIQuestion(message)
}

func (tb *TelegramBot) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		tb.handleStartCommand(message)
	case "help":
		tb.handleHelpCommand(message)
	case "ask":
		tb.handleAskCommand(message)
	case "info":
		tb.handleInfoCommand(message)
	default:
		tb.handleUnknownCommand(message)
	}
}

func (tb *TelegramBot) handleStartCommand(message *tgbotapi.Message) {
	text := greeting
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyToMessageID = message.MessageID

	tb.bot.Send(msg)
}

func (tb *TelegramBot) handleHelpCommand(message *tgbotapi.Message) {
	text := commandHelper

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	tb.bot.Send(msg)
}

func (tb *TelegramBot) handleAskCommand(message *tgbotapi.Message) {
	question := strings.TrimSpace(strings.TrimPrefix(message.Text, "/ask"))

	if question == "" {
		text := emptyQuestion
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		msg.ParseMode = "Markdown"
		tb.bot.Send(msg)
		return
	}

	tb.processAIQuestion(message, question)
}

func (tb *TelegramBot) handleInfoCommand(message *tgbotapi.Message) {
	text := aboutBotInfo
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	tb.bot.Send(msg)
}

func (tb *TelegramBot) handleUnknownCommand(message *tgbotapi.Message) {
	text := unknownCommand
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	tb.bot.Send(msg)
}

func (tb *TelegramBot) handleAIQuestion(message *tgbotapi.Message) {
	tb.processAIQuestion(message, message.Text)
}

func (tb *TelegramBot) processAIQuestion(message *tgbotapi.Message, question string) {
	chatAction := tgbotapi.NewChatAction(message.Chat.ID, tgbotapi.ChatTyping)
	tb.bot.Send(chatAction)

	answer, err := tb.aiClient.Ask(question)
	if err != nil {
		log.Printf("AI error: %v", err)

		errorMessage := "❌ Не удалось обработать запрос.\n"

		if strings.Contains(err.Error(), "401") {
			errorMessage += "Ошибка авторизации API. Проверьте API ключ."
		} else if strings.Contains(err.Error(), "429") {
			errorMessage += "Превышен лимит запросов. Попробуйте через минуту."
		} else if strings.Contains(err.Error(), "402") {
			errorMessage += "Недостаточно средств на счету API."
		} else if strings.Contains(err.Error(), "no response") {
			errorMessage += "AI не вернул ответ. Попробуйте переформулировать вопрос."
		} else {
			errorMessage += "Техническая ошибка: " + err.Error()
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, errorMessage)
		msg.ReplyToMessageID = message.MessageID
		tb.bot.Send(msg)
		return
	}

	if len(answer) > 4000 {
		answer = answer[:4000] + "...\n\n⚠️ Ответ был обрезан из-за ограничений Telegram"
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, answer)
	msg.ReplyToMessageID = message.MessageID
	msg.ParseMode = "Markdown"

	if _, err := tb.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

package bot

import (
	"GolangtgBot/internal/ai"
	"GolangtgBot/internal/rag"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	bot         *tgbotapi.BotAPI
	aiClient    ai.AIClient
	ragPipeline *rag.RAGPipeline
	debugMode   bool
}

func NewBot(token string, aiClient ai.AIClient, debug bool) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания сессии: %v", err)
	}

	bot.Debug = debug
	log.Printf("Авторизация аккаунта %s", bot.Self.UserName)

	ragPipeline := rag.NewRAGPipeline()

	return &TelegramBot{
		bot:         bot,
		aiClient:    aiClient,
		ragPipeline: ragPipeline,
		debugMode:   debug,
	}, nil
}

func (tb *TelegramBot) Start() {
	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "Запустить бота",
		},
		{
			Command:     "help",
			Description: "Помощь по командам",
		},
		{
			Command:     "ask",
			Description: "Задать вопрос AI",
		},
		{
			Command:     "info",
			Description: "Информация о боте",
		},
		{
			Command:     "rag_stats",
			Description: "Статистика RAG базы знаний",
		},
		{
			Command:     "rag_add",
			Description: "Добавить документ в базу знаний",
		},
	}

	config := tgbotapi.NewSetMyCommands(commands...)
	_, err := tb.bot.Request(config)
	if err != nil {
		log.Printf("Ошибка настройки команд: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tb.bot.GetUpdatesChan(u)

	log.Println(statusBotOK)

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
	case "rag_stats":
		tb.handleRAGStatsCommand(message)
	case "rag_add":
		tb.handleRAGAddCommand(message)
	default:
		tb.handleUnknownCommand(message)
	}
}

func (tb *TelegramBot) handleStartCommand(message *tgbotapi.Message) {
	text := greeting

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	//msg.ParseMode = "Markdown"
	msg.ReplyToMessageID = message.MessageID

	tb.bot.Send(msg)
}

func (tb *TelegramBot) handleHelpCommand(message *tgbotapi.Message) {
	text := commandHelper

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	//msg.ParseMode = "Markdown"
	tb.bot.Send(msg)
}

func (tb *TelegramBot) handleAskCommand(message *tgbotapi.Message) {
	question := strings.TrimSpace(strings.TrimPrefix(message.Text, "/ask"))

	if question == "" {
		text := waitYourASK
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		//msg.ParseMode = "Markdown"
		tb.bot.Send(msg)
		return
	}

	tb.processAIQuestion(message, question)
}

func (tb *TelegramBot) handleInfoCommand(message *tgbotapi.Message) {
	text := aboutBotInfo

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	//msg.ParseMode = "Markdown"
	tb.bot.Send(msg)
}

func (tb *TelegramBot) handleRAGStatsCommand(message *tgbotapi.Message) {
	stats := tb.ragPipeline.GetStats()

	text := fmt.Sprintf(`📊 *RAG Statistics*

• Документов в базе: %d
• Слов в словаре: %d
• Размер хранилища: %s

Используйте /rag_add чтобы добавить документы в базу знаний.`,

		stats["total_documents"],
		stats["vocabulary_size"],
		stats["store_size"])

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	//msg.ParseMode = "Markdown"
	tb.bot.Send(msg)
}

func (tb *TelegramBot) handleRAGAddCommand(message *tgbotapi.Message) {
	content := strings.TrimSpace(strings.TrimPrefix(message.Text, "/rag_add"))

	if content == "" {
		text := rag_db_add
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		//msg.ParseMode = "Markdown"
		tb.bot.Send(msg)
		return
	}

	docID := tb.ragPipeline.AddDocument(content)

	text := fmt.Sprintf("✅ Документ добавлен в базу знаний\n\nID: `%s`\nТекст: %s", docID, content)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	//msg.ParseMode = "Markdown"
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

func (tb *TelegramBot) splitMessage(text string, maxLength int) []string {
	if len(text) <= maxLength {
		return []string{text}
	}

	var parts []string
	currentPart := ""

	sentences := strings.Split(text, ". ")

	for _, sentence := range sentences {
		if len(sentence) > maxLength {
			if currentPart != "" {
				parts = append(parts, currentPart)
				currentPart = ""
			}
			for len(sentence) > 0 {
				if len(sentence) <= maxLength {
					parts = append(parts, sentence)
					break
				}

				breakIndex := strings.LastIndex(sentence[:maxLength], " ")
				if breakIndex == -1 {
					breakIndex = maxLength
				}

				parts = append(parts, sentence[:breakIndex])
				sentence = sentence[breakIndex:]
				sentence = strings.TrimSpace(sentence)
			}
		} else if len(currentPart)+len(sentence)+2 <= maxLength {
			if currentPart != "" {
				currentPart += ". " + sentence
			} else {
				currentPart = sentence
			}
		} else {
			if currentPart != "" {
				parts = append(parts, currentPart+".")
				currentPart = sentence
			}
		}
	}

	if currentPart != "" {
		parts = append(parts, currentPart)
	}

	return parts
}

func (tb *TelegramBot) sendSplitMessage(chatID int64, text string, replyToMessageID int) {
	maxLength := 3800

	parts := tb.splitMessage(text, maxLength)

	for i, part := range parts {
		msg := tgbotapi.NewMessage(chatID, part)

		if i == 0 {
			msg.ReplyToMessageID = replyToMessageID
		}

		if len(parts) > 1 {
			part = fmt.Sprintf("📄 [%d/%d]\n\n%s", i+1, len(parts), part)
			msg.Text = part
		}

		if _, err := tb.bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки части сообщения %d: %v", i+1, err)
		}
	}
}

func (tb *TelegramBot) processAIQuestion(message *tgbotapi.Message, question string) {
	chatAction := tgbotapi.NewChatAction(message.Chat.ID, tgbotapi.ChatTyping)
	tb.bot.Send(chatAction)

	ragContext, foundDocs := tb.ragPipeline.ProcessQuery(question)

	log.Printf("RAG нашел %d релевантные документы для: %s", len(foundDocs), question)

	var prompt string
	if len(foundDocs) > 0 {
		prompt = ragContext + "\n\nНа основе контекста выше, ответь на вопрос: " + question + "\n\nБудь кратким и информативным. " +
			"Если в контексте нет точного ответа, используй свои знания."
	} else {
		prompt = "Вопрос: " + question + "\n\nОтветь как полезный ассистент на русском языке. Будь кратким и информативным."
	}

	answer, err := tb.aiClient.Ask(prompt)
	if err != nil {
		log.Printf("ИИ ошибка: %v", err)

		errorMessage := errorRequest

		if strings.Contains(err.Error(), "401") {
			errorMessage += "Ошибка авторизации API. Проверьте API ключ."
		} else if strings.Contains(err.Error(), "429") {
			errorMessage += "Превышен лимит запросов. Попробуйте через минуту."
		} else if strings.Contains(err.Error(), "402") {
			errorMessage += "Недостаточно средств на счету API."
		} else if strings.Contains(err.Error(), "no response") {
			errorMessage += "ИИ не вернул ответ. Попробуйте переформулировать вопрос."
		} else {
			errorMessage += "Техническая ошибка: " + err.Error()
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, errorMessage)
		msg.ReplyToMessageID = message.MessageID
		tb.bot.Send(msg)
		return
	}

	var prefix string
	if len(foundDocs) > 0 {
		prefix = "🔍 *На основе базы знаний:*\n\n"
	} else {
		prefix = "🤖 *Ответ:*\n\n"
	}

	fullAnswer := prefix + answer

	tb.sendSplitMessage(message.Chat.ID, fullAnswer, message.MessageID)
}

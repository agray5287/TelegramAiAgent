package main

import (
	"context"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// Load environment variables from .env file.
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	telegramToken := os.Getenv("TELEGRAM_API_KEY")
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if telegramToken == "" || openaiAPIKey == "" {
		log.Fatal("Missing TELEGRAM_API_KEY or OPENAI_API_KEY environment variables")
	}

	// Initialize Telegram Bot.
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatal("Error creating Telegram bot:", err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Create OpenAI GPT-4 client.
	aiClient := openai.NewClient(openaiAPIKey)

	// Set up update configuration.
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Process incoming updates.
	for update := range updates {
		if update.Message == nil {
			continue
		}
		// Handle each update.
		go handleUpdate(bot, aiClient, update)
	}
}

// handleUpdate sends the user's message to GPT‑4 and sends the response back to Telegram.
func handleUpdate(bot *tgbotapi.BotAPI, aiClient *openai.Client, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userText := update.Message.Text

	// Build a ChatCompletion request for GPT‑4.
	req := openai.ChatCompletionRequest{
		Model: "gpt-4o",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are a helpful and friendly assistant.",
			},
			{
				Role:    "user",
				Content: userText,
			},
		},
	}

	// Call GPT‑4 API.
	ctx := context.Background()
	resp, err := aiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Println("OpenAI API error:", err)
		return
	}

	// Retrieve the reply from GPT‑4.
	reply := resp.Choices[0].Message.Content

	// Build and send the Telegram message.
	msg := tgbotapi.NewMessage(chatID, reply)
	msg.ReplyToMessageID = update.Message.MessageID
	if _, err := bot.Send(msg); err != nil {
		log.Println("Error sending Telegram message:", err)
	}
}

package scheduler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
	openai "github.com/sashabaranov/go-openai"
)

// ScheduledTask represents a single scheduled GPT reminder task.
type ScheduledTask struct {
	Task   string `json:"task"`   // Human-readable task name or description
	Cron   string `json:"cron"`   // Cron expression for scheduling
	Prompt string `json:"prompt"` // Custom prompt to send to GPT
}

// LoadTasks reads and unmarshals the JSON config at the given path into a slice of ScheduledTask.
func LoadTasks(path string) ([]ScheduledTask, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tasks []ScheduledTask
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// StartScheduler sets up and starts cron jobs for each ScheduledTask.
// - bot: your Telegram Bot API instance
// - aiClient: the OpenAI GPT client
// - chatID: Telegram chat ID where reminders are sent
// - systemPrompt: base system prompt for GPT context
// - configPath: path to your tasks.json file
func StartScheduler(bot *tgbotapi.BotAPI, aiClient *openai.Client, chatID int64, systemPrompt, configPath string) {
	tasks, err := LoadTasks(configPath)
	if err != nil {
		log.Fatalf("Error loading tasks from %s: %v", configPath, err)
	}

	c := cron.New()
	for _, t := range tasks {
		// Capture loop variable
		task := t
		_, err := c.AddFunc(task.Cron, func() {
			sendDynamicReminder(bot, aiClient, chatID, systemPrompt, task)
		})
		if err != nil {
			log.Printf("Invalid cron expression '%s' for task '%s': %v", task.Cron, task.Task, err)
		}
	}
	c.Start()
	log.Println("Scheduler started with tasks from", configPath)
}

// sendDynamicReminder builds a ChatCompletion request using the task's prompt and sends the GPT-generated message via Telegram.
func sendDynamicReminder(bot *tgbotapi.BotAPI, aiClient *openai.Client, chatID int64, systemPrompt string, task ScheduledTask) {
	// Build the GPT request
	req := openai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: task.Prompt},
		},
	}

	ctx := context.Background()
	resp, err := aiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("OpenAI error for task '%s': %v", task.Task, err)
		return
	}

	// Extract GPT's reply
	reply := resp.Choices[0].Message.Content

	// Send the message to Telegram
	msg := tgbotapi.NewMessage(chatID, reply)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Telegram send error for task '%s': %v", task.Task, err)
	}
}

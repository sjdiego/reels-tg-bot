package telegram

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"reels-tg-bot/pkg/env"
	"reels-tg-bot/pkg/instagram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Run func
func Run() {
	bot, err := tgbotapi.NewBotAPI(env.GetEnv("TG_BOT_API_KEY"))
	if err != nil {
		log.Panic(err)
	}

	newUpdate := tgbotapi.NewUpdate(0)
	newUpdate.Timeout = 10

	updates, err := bot.GetUpdatesChan(newUpdate)

	for update := range updates {
		handleUpdate(bot, update.Message)
	}
}

func handleUpdate(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	// Display incoming messages
	if len(msg.From.UserName) > 0 {
		log.Printf("[%d] <%s @%s> %s\n", msg.From.ID, msg.From.FirstName, msg.From.UserName, msg.Text)
	} else {
		log.Printf("[%d] <%s> %s\n", msg.From.ID, msg.From.FirstName, msg.Text)
	}

	// Check if message comes from authorized user
	if len(env.GetEnv("TG_ADMIN_ID")) > 0 && env.GetEnv("TG_ADMIN_ID") == fmt.Sprint(msg.From.ID) {
		// Check if message matches with a Reels IG URL like:
		// https://www.instagram.com/reel/CLjJYuhFs24/
		re := regexp.MustCompile(`^https?://www\.instagram\.com/reel/([A-Za-z0-9-]{11})`)
		code := re.FindStringSubmatch(strings.TrimSpace(msg.Text))

		if len(code) > 1 && !msg.From.IsBot {
			handleReelMessage(bot, msg, code)
		}
	}
}

func handleReelMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, code []string) {
	// Notify user that URL is being processed
	message := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Starting download for %s", code[0]))
	sentMessage, _ := bot.Send(message)

	// Download file and get path on filesystem
	videoPath := instagram.Get(code[1])

	// If download was successful, create Video message and send it to user
	if len(videoPath) > 0 {
		fmt.Printf("Sending video to %s (ID: %d)\n", msg.From.FirstName, msg.From.ID)
		chatActionConfig := tgbotapi.NewChatAction(msg.Chat.ID, tgbotapi.ChatUploadVideo)
		bot.Send(chatActionConfig)
		videoConfig := tgbotapi.NewVideoUpload(msg.Chat.ID, videoPath)
		bot.Send(videoConfig)
	}

	// Delete previous message
	deleteConfig := tgbotapi.NewDeleteMessage(msg.Chat.ID, sentMessage.MessageID)
	bot.DeleteMessage(deleteConfig)
}

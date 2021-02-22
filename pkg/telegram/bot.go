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

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		u := update.Message

		// Display incoming messages
		if len(u.From.UserName) > 0 {
			log.Printf("[%d] <%s @%s> %s\n", u.From.ID, u.From.FirstName, u.From.UserName, u.Text)
		} else {
			log.Printf("[%d] <%s> %s\n", u.From.ID, u.From.FirstName, u.Text)
		}

		// Check if message matches with a Reels IG URL like:
		// https://www.instagram.com/reel/CLjJYuhFs24/
		re := regexp.MustCompile(`^https?://www\.instagram\.com/reel/([A-Za-z0-9-]{11})`)
		code := re.FindStringSubmatch(strings.TrimSpace(u.Text))

		if len(code) > 1 && !u.From.IsBot {
			// Notify user that URL is being processed
			message := tgbotapi.NewMessage(u.Chat.ID, fmt.Sprintf("Starting download for %s", code[0]))
			sentMessage, _ := bot.Send(message)

			// Download file and get path on filesystem
			videoPath := instagram.Get(code[1])

			// If download was successful, create Video message and send it to user
			if len(videoPath) > 0 {
				videoConfig := tgbotapi.NewVideoUpload(u.Chat.ID, videoPath)
				fmt.Printf("Sending video to %s (ID: %d)\n", u.From.FirstName, u.From.ID)
				bot.Send(videoConfig)
			}

			// Delete previous message
			deleteConfig := tgbotapi.NewDeleteMessage(u.Chat.ID, sentMessage.MessageID)
			bot.DeleteMessage(deleteConfig)
		}
	}
}

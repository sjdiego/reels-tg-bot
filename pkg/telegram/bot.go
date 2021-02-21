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
		log.Printf("<%v> %s\n\n", update.Message.From.ID, update.Message.Text)

		// https://www.instagram.com/reel/CLjJYuhFs24/
		re := regexp.MustCompile(`^https?://www\.instagram\.com/reel/([A-Za-z0-9]{11})`)
		code := re.FindStringSubmatch(strings.TrimSpace(update.Message.Text))

		if len(code) > 1 && !update.Message.From.IsBot {
			message := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				fmt.Sprintf("Starting download from %s", update.Message.Text),
			)
			sentMessage, _ := bot.Send(message)

			videoPath := instagram.Get(code[1])
			videoConfig := tgbotapi.NewVideoUpload(update.Message.Chat.ID, videoPath)
			fmt.Printf("Sending video to %s (ID: %d)\n", update.Message.From.FirstName, update.Message.From.ID)
			bot.Send(videoConfig)

			deleteConfig := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, sentMessage.MessageID)
			bot.DeleteMessage(deleteConfig)
		}
	}
}

package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"reels-tg-bot/pkg/env"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Run starts bot
func Run() {
	bot, err := tgbotapi.NewBotAPI(env.GetEnv("TG_BOT_API_KEY"))
	if err != nil {
		log.Panic(err)
	}

	newUpdate := tgbotapi.NewUpdate(0)
	newUpdate.Timeout = 10

	updates, err := bot.GetUpdatesChan(newUpdate)

	for update := range updates {
		printUpdateResponse(update)

		if update.Message != nil {
			HandleUpdate(bot, update.Message)
		}
	}
}

// CheckUserAuth matches provided user ID against stored ID in .env file
func CheckUserAuth(userID int) bool {
	adminID := env.GetEnv("TG_ADMIN_ID")
	if len(adminID) > 0 && adminID == fmt.Sprint(userID) {
		return true
	} else if len(adminID) == 0 {
		return true
	}
	return false
}

// printUpdateResponse displays a prettied JSON of incoming Update
func printUpdateResponse(update tgbotapi.Update) {
	if env.GetEnv("DEBUG_MODE") == "true" {
		bodyBytes := new(bytes.Buffer)
		json.NewEncoder(bodyBytes).Encode(update)

		var prettyJSON bytes.Buffer
		_ = json.Indent(&prettyJSON, bodyBytes.Bytes(), "", "\t")

		fmt.Println(string(prettyJSON.Bytes()))
	}
}

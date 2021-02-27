package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"reels-tg-bot/pkg/env"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Run starts bot
func Run() {
	bot, err := tgbotapi.NewBotAPI(env.GetEnv("TG_BOT_API_KEY"))
	if err != nil {
		log.Panic(err)
	}

	var updates tgbotapi.UpdatesChannel

	if len(env.GetEnv("APP_URL_HEROKU")) > 0 {
		updates = getUpdatesFromWebhook(bot, env.GetEnv("APP_URL_HEROKU"))
	} else {
		updates = getUpdatesWithPolling(bot)
	}

	for update := range updates {
		printUpdateResponse(update)

		log.Println(update.UpdateID) // debug test

		if update.Message != nil {
			HandleUpdate(bot, update.Message)
		}
	}
}

func getUpdatesFromWebhook(bot *tgbotapi.BotAPI, webhookURL string) tgbotapi.UpdatesChannel {
	bot.RemoveWebhook() // Removes previous webhook
	webhookInfo, err := bot.SetWebhook(tgbotapi.NewWebhook(strings.TrimRight(webhookURL, "/") + "/" + bot.Token))
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("New webhook successfully set: %s\n", webhookInfo.Description)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Webhook URL set on: %s\n", info.URL)
	}
	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s\n", info.LastErrorMessage)
	}

	go http.ListenAndServe(":"+env.GetEnv("PORT"), nil)

	return bot.ListenForWebhook("/" + bot.Token)
}

func getUpdatesWithPolling(bot *tgbotapi.BotAPI) tgbotapi.UpdatesChannel {
	bot.RemoveWebhook()

	newUpdate := tgbotapi.NewUpdate(0)
	newUpdate.Timeout = 10

	updates, err := bot.GetUpdatesChan(newUpdate)

	if err != nil {
		log.Println(err)
	}

	return updates
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

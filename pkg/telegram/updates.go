package telegram

import (
	"fmt"
	"log"
	"reels-tg-bot/pkg/instagram"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// HandleUpdate manages updates received from Telegram API
func HandleUpdate(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	// Display incoming messages
	if len(msg.From.UserName) > 0 {
		log.Printf("[%d] <%s @%s> %s\n", msg.From.ID, msg.From.FirstName, msg.From.UserName, msg.Text)
	} else {
		log.Printf("[%d] <%s> %s\n", msg.From.ID, msg.From.FirstName, msg.Text)
	}

	// Check if message comes from authorized user
	if CheckUserAuth(msg.From.ID) && handleReelMessage(bot, msg) {
		fmt.Printf("Video sent successfully to %d\n", msg.From.ID)
	}
}

// handleReelMessage checks for incoming messages and returns Video file if it is an Instagram Reel URL
// https://www.instagram.com/reel/CLjJYuhFs24/
func handleReelMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) bool {
	re := regexp.MustCompile(`^https?://www\.instagram\.com/reel/([A-Za-z0-9-_]{11})`)
	code := re.FindStringSubmatch(strings.TrimSpace(msg.Text))
	if len(code) < 1 && !msg.From.IsBot {
		return false
	}
	fmt.Printf("Detected IG Reels URL from %d\n", msg.From.ID)

	// Notify user that URL is being processed
	message := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Starting download for %s ...", code[0]))
	sentMessage, _ := bot.Send(message)

	// Download file and get path on filesystem
	videoPath, status := instagram.Get(code[1])

	// Notify user if download was unsuccessful
	if !status {
		deleteConfig := tgbotapi.NewDeleteMessage(msg.Chat.ID, sentMessage.MessageID)
		bot.DeleteMessage(deleteConfig)

		warningMsgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Something went wrong. Please try again.")
		warningMsgConfig.ReplyToMessageID = msg.MessageID
		bot.Send(warningMsgConfig)

		return false
	}

	fmt.Println("Sending video to Telegram servers...")

	// Send TG action of uploading video to user
	chatActionConfig := tgbotapi.NewChatAction(msg.Chat.ID, tgbotapi.ChatUploadVideo)
	bot.Send(chatActionConfig)

	videoConfig := tgbotapi.NewVideoUpload(msg.Chat.ID, videoPath)
	videoConfig.ReplyToMessageID = msg.MessageID
	bot.Send(videoConfig)

	// Delete previous message
	deleteConfig := tgbotapi.NewDeleteMessage(msg.Chat.ID, sentMessage.MessageID)
	bot.DeleteMessage(deleteConfig)

	return true
}

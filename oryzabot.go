package main

import (
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"strings"
	"github.com/expectocode/backend"
	"os"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("ORYZA_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	back := backend.NewBackend(os.Getenv("ORYZA_DB"))
	log.Printf("Backend %s", back)

	// bot update loop
	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		// If it's a /upload, upload the reply-message.
		// Else if it's a private chat, upload the message.
		if strings.HasPrefix(update.Message.Text, "/upload") {
			HandleUploadCommand(bot, update)
		} else if update.Message.Chat.IsPrivate() {
			Upload(update.Message, update.Message.From.ID, update.Message.Date)
		}
	}
}

func HandleUploadCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	//Upload the replied to message. If there is no reply, complain.
	upload_msg := update.Message.ReplyToMessage
	if upload_msg == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Please reply to a message to upload it")
		bot.Send(msg)
		return
	}

	// Message.From may be nil if the message is in a Channel
	// but who would use this in a channel?
	Upload(upload_msg, update.Message.From.ID, update.Message.Date)
	//msg := tgbotapi.NewMessage(update.Message.Chat.ID, "received")
	//msg.ReplyToMessageID = update.Message.MessageID
	//bot.Send(msg)
}

func Upload(message *tgbotapi.Message, sender_id int, send_timestamp int) {
	//Try to upload the given message as text, photo, or file
	//TODO implement this with a bunch of calls to the backend
	if message.Document != nil {
		var filename string
		if message.Document.FileName != "" {
			filename = message.Document.FileName
		} else {
			//TODO Make filename from other data - time, sender, mime. Or maybe just random alphanum
			filename = fmt.Sprintf("%s", message.Date)
		}
		log.Printf("%s", filename)
	//	backend.upload(filename, mimetype, "file", sender_id, send_timestamp)
	}
}

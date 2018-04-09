package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"gopkg.in/telegram-bot-api.v4"
)

const apiUrl = "http://localhost:8000/api/"

func main() {
	botToken := os.Getenv("ORYZA_BOT_TOKEN")
	if botToken == "" {
		log.Panic("You must set $ORYZA_BOT_TOKEN")
	}
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic("Could not start bot: %s", err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic("Error getting bot updates: ", err)
	}

	tokenMode := make(map[int]bool)
	db_path := os.Getenv("ORYZA_BOT_DB")
	if db_path == "" {
		log.Panic("ORYZA_BOT_DB must be set!")
	}
	db, err := bolt.Open(db_path, 0600, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		log.Fatal("Could not open bolt db: %s", err)
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("IdTokenMap"))
		if err != nil {
			return fmt.Errorf("create bucket error: %s", err)
		}
		return nil
	})

	// bot update loop
	for update := range updates {
		if update.Message == nil || update.Message.From == nil {
			// Message.From may be nil if the message is in a Channel
			// but who would use this in a channel?
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		if tokenMode[update.Message.From.ID] {
			// If we are waiting for a token from this person, dont take commands
			receiveToken(bot, update, &tokenMode, db)
			continue
		}
		// If it's a /upload, upload the reply-message.
		// Else if it's a private chat, upload the message.
		if strings.HasPrefix(update.Message.Text, "/") {
			if matched, _ := regexp.MatchString("^/upload($|\\s)",
				update.Message.Text); matched {
				go HandleUploadCommand(bot, update, db)
			} else if matched, _ := regexp.MatchString("^/delete?($|\\s)",
				update.Message.Text); matched {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Deletion not implemented yet")
				msg.ReplyToMessageID = update.Message.MessageID
				go bot.Send(msg)
				//check if theres a url in the /delete <thing> or in the reply message
			} else if matched, _ := regexp.MatchString("^/start($|\\s)",
				update.Message.Text); matched {
				go requestToken(bot, update, &tokenMode)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Unrecognised command")
				msg.ReplyToMessageID = update.Message.MessageID
				go bot.Send(msg)
			}
		} else if update.Message.Chat.IsPrivate() {
			go HandlePrivateMessage(bot, update, db)
		}
	}
}

func getToken(id uint32, db *bolt.DB) string {
	var token string
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("IdTokenMap"))
		bin := make([]byte, 4)
		binary.LittleEndian.PutUint32(bin, id)
		v := b.Get(bin)

		token = string(v)
		return nil
	})
	return token
}

func receiveToken(bot *tgbotapi.BotAPI, update tgbotapi.Update,
	modemap *map[int]bool, db *bolt.DB) {
	var msg tgbotapi.MessageConfig
	if matched, _ := regexp.MatchString("/cancel($|\\s)", update.Message.Text); matched {
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Cancelled.")
		msg.ReplyToMessageID = update.Message.MessageID
		go bot.Send(msg)
		(*modemap)[update.Message.From.ID] = false
		return
	} else if matched, _ := regexp.MatchString("[a-zA-Z0-9]{16}",
		update.Message.Text); !matched {
		// This doesn't look like a token
		msg = tgbotapi.NewMessage(update.Message.Chat.ID,
			`Hm, that doesn't look like a token. It should be 16 chars of letters and numbers. (send /cancel to stop)`)
		msg.ReplyToMessageID = update.Message.MessageID
		go bot.Send(msg)
		return
	} // So from now on, it's a good token format.
	// Insert the token
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("IdTokenMap"))
		bin := make([]byte, 4)
		binary.LittleEndian.PutUint32(bin, uint32(update.Message.From.ID))
		err := b.Put(bin, []byte(update.Message.Text))
		if err != nil {
			msg = tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf(`Error saving your token, sorry!
Report this: %s`, err))
			msg.ReplyToMessageID = update.Message.MessageID
			go bot.Send(msg)
			return fmt.Errorf("insert error: %s", err)
		}
		return nil
	})
	(*modemap)[update.Message.From.ID] = false
	msg = tgbotapi.NewMessage(update.Message.Chat.ID,
		"Thanks! You can now upload through me with /upload or by PMing me")
	msg.ReplyToMessageID = update.Message.MessageID
	go bot.Send(msg)
}

func requestToken(bot *tgbotapi.BotAPI, update tgbotapi.Update,
	modemap *map[int]bool) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"Please send your upload token")
	msg.ReplyToMessageID = update.Message.MessageID
	bot.Send(msg)
	(*modemap)[update.Message.From.ID] = true
}

func HandlePrivateMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *bolt.DB) {
	// Message.Chat is the same as Message.From in PM
	token := getToken(uint32(update.Message.From.ID), db)
	if token == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"You must send me your upload token with /start before you can use me")
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		return
	}

	resp, err := Upload(bot, update.Message, token, db)
	if err != nil {
		fail(bot, update.Message.From.ID, update.Message.MessageID,
			fmt.Sprintf("Could not upload: %s", err))
	}
	if resp != "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, resp)
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
	}
}

func HandleUploadCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *bolt.DB) {

	if update.Message.Chat.IsPrivate() {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Just send me what you want to upload.")
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		return
	}

	//Upload the replied to message. If there is no reply, complain.
	upload_msg := update.Message.ReplyToMessage
	if upload_msg == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Please reply to a message to upload it")
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		return
	}

	token := getToken(uint32(update.Message.From.ID), db)
	if token == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"You must send me your upload token in PM before you can use me")
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		return
	}

	resp, err := Upload(bot, upload_msg, token, db)
	if err != nil {
		fail(bot, update.Message.From.ID, update.Message.MessageID,
			fmt.Sprintf("Could not upload: %s", err))
	}
	if resp != "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, resp)
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
	}
	//msg := tgbotapi.NewMessage(update.Message.Chat.ID, "received")
	//msg.ReplyToMessageID = update.Message.MessageID
	//bot.Send(msg)
}

func fail(bot *tgbotapi.BotAPI, to_id int, reply_id int, text string) {
	log.Printf("Error with user %d: %s", to_id, text)
	msg := tgbotapi.NewMessage(int64(to_id), "Report this error: "+text)
	msg.ReplyToMessageID = reply_id
	bot.Send(msg)
}

func newUploadRequest(mimetype, filename, token, extraInfo string,
	filebody io.ReadCloser) (*http.Request, error) {

	body := &bytes.Buffer{}
	mwriter := multipart.NewWriter(body)
	part, err := mwriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, filebody)
	mwriter.WriteField("mimetype", mimetype)
	mwriter.WriteField("token", token)
	mwriter.WriteField("extrainfo", extraInfo)
	err = mwriter.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiUrl+"upload", body)
	req.Header.Set("Content-Type", mwriter.FormDataContentType())
	return req, err
}

func Upload(bot *tgbotapi.BotAPI, message *tgbotapi.Message, token string,
	db *bolt.DB) (string, error) {
	// Try to upload the given message as text, photo, or file
	// Return a bot response and an error, mutually exclusive
	// TODO implement other media types
	var uploadReq *http.Request
	client := http.Client{}

	if message.Document != nil {
		fileurl, err := bot.GetFileDirectURL(message.Document.FileID)
		if err != nil {
			return "", err
		}
		log.Printf("File ID %s, url %s, mime %s", message.Document.FileID, fileurl,
			message.Document.MimeType)
		fileresp, err := http.Get(fileurl)
		if err != nil {
			return "", err
		}
		defer fileresp.Body.Close()
		// Form a POST request
		uploadReq, err = newUploadRequest(message.Document.MimeType,
			message.Document.FileName, token, message.Caption, fileresp.Body)
		if err != nil {
			return "", err
		}
	} else if message.Photo != nil {
		biggest_photo := (*message.Photo)[len((*message.Photo))-1]
		log.Println("photo message %s", biggest_photo)
		fileurl, err := bot.GetFileDirectURL(biggest_photo.FileID)
		if err != nil {
			return "", err
		}
		fileresp, err := http.Get(fileurl)
		if err != nil {
			return "", err
		}
		defer fileresp.Body.Close()
		// Form a POST request
		uploadReq, err = newUploadRequest("image/jpeg", // photos always jpeg
			"photo", token, message.Caption, fileresp.Body)
		if err != nil {
			return "", err
		}
	} else if message.Sticker != nil {
		fileurl, err := bot.GetFileDirectURL(message.Sticker.FileID)
		if err != nil {
			return "", err
		}
		fileresp, err := http.Get(fileurl)
		if err != nil {
			return "", err
		}
		defer fileresp.Body.Close()
		uploadReq, err = newUploadRequest("image/webp", // stickers always webp
			"sticker", token, "", fileresp.Body)
		if err != nil {
			return "", err
		}
	} else {
		return "Unrecognised message type", nil
	}

	log.Println("finished, uploading")
	resp, err := client.Do(uploadReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data := make(map[string]interface{})
	rBody, err := ioutil.ReadAll(resp.Body)
	log.Println("API response body", resp.Body)
	err = json.Unmarshal(rBody, &data)
	if err != nil {
		return "", err
	}
	log.Println("API response", data)
	succ, ok := data["success"].(string)
	if !ok {
		return "", errors.New("could not interpret api response")
	}
	if succ == "true" {
		url, ok := data["url"].(string)
		if !ok {
			return "", errors.New("could not interpret api response")
		}
		return fmt.Sprintf("Uploaded at %s", url), nil
	} else {
		return "", errors.New(fmt.Sprintf("%s", data["reason"]))
	}
}

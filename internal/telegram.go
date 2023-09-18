package iternal

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	BotChatID int64
	TgBot     *tgbotapi.BotAPI
}

var usersInChat Users

type User struct {
	id        int64
	name      string
	userState string
}

type Users []*User

func NewTgBot() *Bot {
	token := os.Getenv("TG_BOT_TOKEN")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	//bot.Debug = true

	return &Bot{
		TgBot: bot,
	}
}

func (b *Bot) handleFile(update *tgbotapi.Update) error {
	fileLink, _ := b.TgBot.GetFileDirectURL(update.Message.Document.FileID)

	out, err := os.Create("./input.png")
	defer out.Close()

	resp, err := http.Get(fileLink)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (b *Bot) removeTmpFiles() error {
	err := os.Remove("./output.png")
	if err != nil {
		return err
	}

	err = os.Remove("./input.png")
	if err != nil {
		return err
	}

	return nil
}

func (b *Bot) isStartMessage(update *tgbotapi.Update) bool {
	return update.Message != nil && update.Message.Text == "/start"
}

func (b *Bot) isCallBackQuery(update *tgbotapi.Update) bool {
	return update.CallbackQuery != nil && update.CallbackQuery.Data != ""
}

func (b *Bot) printSysMessage(delay uint8, text string) {
	msg := tgbotapi.NewMessage(b.BotChatID, text)
	b.TgBot.Send(msg)

	time.Sleep(time.Second * time.Duration(delay))
}

func (b *Bot) printIntro(update *tgbotapi.Update) {
	b.printSysMessage(1, "Send photo without compression as a file")
	b.printSysMessage(1, "In .png or .jpg formats")
}

func (b *Bot) getKeyboardRow(btnName, btnCode string) []tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnName, btnCode))
}

func (b *Bot) askToPrintIntro() {
	msg := tgbotapi.NewMessage(b.BotChatID, "Read tutorial?")

	btn := b.getKeyboardRow("Read tutorial", "read_tutorial")
	btn2 := b.getKeyboardRow("Skip tutorial", "skip_tutorial")

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(btn, btn2)

	b.TgBot.Send(msg)
}

func (b *Bot) showMenu() {
	msg := tgbotapi.NewMessage(b.BotChatID, "Choose one option:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		b.getKeyboardRow("Object recognition", "object_recognition"),
		b.getKeyboardRow("Text recognition", "text_recognition"),
		b.getKeyboardRow("Nudity recognition", "nudity_recognition"),
		b.getKeyboardRow("Personal Projective Equipment", "personal_projective_equipment"),
	)
	b.TgBot.Send(msg)
}

func callBackQueryIsMissing(update *tgbotapi.Update) bool {
	return update.CallbackQuery == nil || update.CallbackQuery.From == nil
}

func getUserFromUpdate(update *tgbotapi.Update) (user *User, found bool) {
	if callBackQueryIsMissing(update) {
		return nil, false
	}
	userId := update.CallbackQuery.From.ID
	for _, userInChat := range usersInChat {
		if userId == userInChat.id {
			return userInChat, true
		}
	}

	return nil, false
}

func storeUserFromUpdate(update *tgbotapi.Update) (user *User, found bool) {
	if callBackQueryIsMissing(update) {
		return nil, false
	}

	user = &User{
		id:        update.CallbackQuery.From.ID,
		name:      strings.TrimSpace(update.CallbackQuery.From.FirstName + " " + update.CallbackQuery.From.LastName),
		userState: "",
	}

	usersInChat = append(usersInChat, user)

	return user, true
}

func (b *Bot) updateProcessing(update *tgbotapi.Update) {
	user, found := getUserFromUpdate(update)
	if !found {
		user, found = storeUserFromUpdate(update)
	}

	choiceCode := update.CallbackQuery.Data

	switch choiceCode {
	case "read_tutorial":
		b.printIntro(update)
		b.showMenu()
	case "skip_tutorial":
		b.showMenu()
	case "object_recognition":
		user.userState = "object_recognition"
		msg := tgbotapi.NewMessage(b.BotChatID, "Send image to recognize")
		b.TgBot.Send(msg)
	case "text_recognition":
		user.userState = "text_recognition"
		msg := tgbotapi.NewMessage(b.BotChatID, "Send image to recognize")
		b.TgBot.Send(msg)
	case "nudity_recognition":
		user.userState = "nudity_recognition"
		msg := tgbotapi.NewMessage(b.BotChatID, "Send image to recognize")
		b.TgBot.Send(msg)
	case "personal_projective_equipment":
		user.userState = "personal_projective_equipment"
		msg := tgbotapi.NewMessage(b.BotChatID, "Send image to recognize")
		b.TgBot.Send(msg)
	}
}

func TgHandler(bot *Bot) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.TgBot.GetUpdatesChan(u)

	awsConfig, err := createAWSConfig()
	if err != nil {
		log.Println(err)
	}

	for update := range updates {

		if bot.isCallBackQuery(&update) {
			bot.updateProcessing(&update)
		} else if bot.isStartMessage(&update) {
			bot.BotChatID = update.Message.Chat.ID

			bot.askToPrintIntro()
		} else if update.Message.Document != nil {

			err = bot.handleFile(&update)
			if err != nil {
				log.Println(err)
			}

			var user *User

			userId := update.SentFrom().ID
			for _, userInChat := range usersInChat {
				if userId == userInChat.id {
					user = userInChat
				}
			}

			switch user.userState {
			case "object_recognition":
				recognizeObjectHandler(awsConfig)

				msg := tgbotapi.NewPhoto(bot.BotChatID, tgbotapi.FilePath("./output.png"))
				bot.TgBot.Send(msg)

				err = bot.removeTmpFiles()
				if err != nil {
					log.Println("can't remove files")
				}

				bot.showMenu()
			case "text_recognition":
				recognizeTextHandler(awsConfig)

				msg := tgbotapi.NewPhoto(bot.BotChatID, tgbotapi.FilePath("./output.png"))
				bot.TgBot.Send(msg)

				err = bot.removeTmpFiles()
				if err != nil {
					log.Println("can't remove files")
				}

				bot.showMenu()
			case "nudity_recognition":
				res, err := recognizeNudityHandler(awsConfig)
				if err != nil {
					fmt.Println(err)
				}

				for _, item := range res {
					msg := tgbotapi.NewMessage(bot.BotChatID, item)
					bot.TgBot.Send(msg)
					time.Sleep(time.Millisecond * time.Duration(50))
				}

			case "personal_projective_equipment":
				recognizePPEHandler(awsConfig)

				msg := tgbotapi.NewPhoto(bot.BotChatID, tgbotapi.FilePath("./output.png"))
				bot.TgBot.Send(msg)

				err = bot.removeTmpFiles()
				if err != nil {
					log.Println("can't remove files")
				}

				bot.showMenu()
			}

		}
	}
}

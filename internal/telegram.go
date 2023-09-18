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

var botChatID int64
var tgBot *tgbotapi.BotAPI

var usersInChat Users

type User struct {
	id        int64
	name      string
	userState string
}

type Users []*User

func NewTgBot() *tgbotapi.BotAPI {
	token := os.Getenv("TG_BOT_TOKEN")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	//bot.Debug = true

	tgBot = bot

	return tgBot
}

func handleFile(update *tgbotapi.Update) error {
	fileLink, _ := tgBot.GetFileDirectURL(update.Message.Document.FileID)

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

func removeTmpFiles() error {
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

func isStartMessage(update *tgbotapi.Update) bool {
	return update.Message != nil && update.Message.Text == "/start"
}

func isCallBackQuery(update *tgbotapi.Update) bool {
	return update.CallbackQuery != nil && update.CallbackQuery.Data != ""
}

func printSysMessage(delay uint8, text string) {
	msg := tgbotapi.NewMessage(botChatID, text)
	tgBot.Send(msg)

	time.Sleep(time.Second * time.Duration(delay))
}

func printIntro(update *tgbotapi.Update) {
	printSysMessage(1, "Send photo without compression as a file")
	printSysMessage(1, "In .png or .jpg formats")
}

func getKeyboardRow(btnName, btnCode string) []tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnName, btnCode))
}

func askToPrintIntro() {
	msg := tgbotapi.NewMessage(botChatID, "Read tutorial?")

	btn := getKeyboardRow("Read tutorial", "read_tutorial")
	btn2 := getKeyboardRow("Skip tutorial", "skip_tutorial")

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(btn, btn2)

	tgBot.Send(msg)
}

func showMenu() {
	msg := tgbotapi.NewMessage(botChatID, "Choose one option:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		getKeyboardRow("Object recognition", "object_recognition"),
		getKeyboardRow("Text recognition", "text_recognition"),
		getKeyboardRow("Nudity recognition", "nudity_recognition"),
		getKeyboardRow("Personal Projective Equipment", "personal_projective_equipment"),
	)
	tgBot.Send(msg)
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

func updateProcessing(update *tgbotapi.Update) {
	user, found := getUserFromUpdate(update)
	if !found {
		user, found = storeUserFromUpdate(update)
	}

	choiceCode := update.CallbackQuery.Data

	switch choiceCode {
	case "read_tutorial":
		printIntro(update)
		showMenu()
	case "skip_tutorial":
		showMenu()
	case "object_recognition":
		user.userState = "object_recognition"
		msg := tgbotapi.NewMessage(botChatID, "Send image to recognize")
		tgBot.Send(msg)
	case "text_recognition":
		user.userState = "text_recognition"
		msg := tgbotapi.NewMessage(botChatID, "Send image to recognize")
		tgBot.Send(msg)
	case "nudity_recognition":
		user.userState = "nudity_recognition"
		msg := tgbotapi.NewMessage(botChatID, "Send image to recognize")
		tgBot.Send(msg)
	case "personal_projective_equipment":
		user.userState = "personal_projective_equipment"
		msg := tgbotapi.NewMessage(botChatID, "Send image to recognize")
		tgBot.Send(msg)
	}
}

func TgHandler() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tgBot.GetUpdatesChan(u)

	awsConfig, err := createAWSConfig()
	if err != nil {
		log.Println(err)
	}

	for update := range updates {

		if isCallBackQuery(&update) {
			updateProcessing(&update)
		} else if isStartMessage(&update) {
			botChatID = update.Message.Chat.ID

			askToPrintIntro()
		} else if update.Message.Document != nil {

			err = handleFile(&update)
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

				msg := tgbotapi.NewPhoto(botChatID, tgbotapi.FilePath("./output.png"))
				tgBot.Send(msg)

				err = removeTmpFiles()
				if err != nil {
					log.Println("can't remove files")
				}

				showMenu()
			case "text_recognition":
				recognizeTextHandler(awsConfig)

				msg := tgbotapi.NewPhoto(botChatID, tgbotapi.FilePath("./output.png"))
				tgBot.Send(msg)

				err = removeTmpFiles()
				if err != nil {
					log.Println("can't remove files")
				}

				showMenu()
			case "nudity_recognition":
				res, err := recognizeNudityHandler(awsConfig)
				if err != nil {
					fmt.Println(err)
				}

				for _, item := range res {
					msg := tgbotapi.NewMessage(botChatID, item)
					tgBot.Send(msg)
					time.Sleep(time.Millisecond * time.Duration(50))
				}

			case "personal_projective_equipment":
				recognizePPEHandler(awsConfig)

				msg := tgbotapi.NewPhoto(botChatID, tgbotapi.FilePath("./output.png"))
				tgBot.Send(msg)

				err = removeTmpFiles()
				if err != nil {
					log.Println("can't remove files")
				}

				showMenu()
			}

		}
	}
}

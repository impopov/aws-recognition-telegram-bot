package main

import (
	"log"

	telegram "github.com/impopov/aws-recognition-telegram-bot/internal"
)

func main() {
	//init bot
	bot := telegram.NewTgBot()

	log.Printf("Authorized on account %s", bot.Self.UserName)

	telegram.TgHandler()
}

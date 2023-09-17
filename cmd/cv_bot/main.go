package main

import (
	telegram "cv_kimimo_bot/internal"
	"log"
)

func main() {
	//init bot
	bot := telegram.NewTgBot()

	log.Printf("Authorized on account %s", bot.TgBot.Self.UserName)

	telegram.TgHandler(bot)
}

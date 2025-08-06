package main

import (
	"log"
	"nimadir/bot"
	"nimadir/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env fayl topilmadi")
	}

	token := config.GetBotToken()
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	botAPI.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := botAPI.GetUpdatesChan(u)

	for update := range updates {
		bot.HandleUpdate(botAPI, update)
	}
}

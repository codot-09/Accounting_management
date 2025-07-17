package main

import (
	"log"
	"net/http"
	"os"
	"nimadir/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	go keepAlive() // Render uchun port ochamiz

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN not set")
	}

	b, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}
	b.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.GetUpdatesChan(u)

	for update := range updates {
		bot.HandleUpdate(b, update)
	}
}

// Bu funksiya HTTP portni ochib turadi, Render to‘xtatib yubormasligi uchun
func keepAlive() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Bot is running!"))
	})
	log.Printf("Listening on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

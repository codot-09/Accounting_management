package config

import (
	"log"
	"os"
)

func GetBotToken() string {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN not set")
	}
	return token
}

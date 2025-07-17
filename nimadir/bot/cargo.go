package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Cargo struct {
	ID     int    `json:"id"`
	Date   string `json:"date"`
	Amount int    `json:"amount"`
	Photo  string `json:"photo"`
}

var cargoTemp = make(map[int64]*Cargo)

func startAddCargo(bot *tgbotapi.BotAPI, chatID int64) {
	cargoTemp[chatID] = &Cargo{}
	msg := tgbotapi.NewMessage(chatID, "Sana kiriting (YYYY-MM-DD):")
	bot.Send(msg)
}

func saveCargo(chatID int64, message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	c := cargoTemp[chatID]
	if c == nil {
		return
	}
	if c.Date == "" {
		_, err := time.Parse("2006-01-02", message.Text)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Noto‘g‘ri sana formati"))
			return
		}
		c.Date = message.Text
		bot.Send(tgbotapi.NewMessage(chatID, "Summani yuboring:"))
		return
	}
	if c.Amount == 0 {
		amt, err := strconv.Atoi(message.Text)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Noto‘g‘ri summa"))
			return
		}
		c.Amount = amt
		bot.Send(tgbotapi.NewMessage(chatID, "Endi rasm yuboring:"))
		return
	}
	if c.Photo == "" && message.Photo != nil {
		photo := message.Photo[len(message.Photo)-1]
		c.Photo = photo.FileID
		appendCargo(*c)
		delete(cargoTemp, chatID)
		bot.Send(tgbotapi.NewMessage(chatID, "✅ Yuk saqlandi."))
		return
	}

	if c.Photo == "" && message.Photo == nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Iltimos, rasm yuboring."))
	}
}

func appendCargo(c Cargo) {
	file, _ := os.ReadFile("data/cargo.json")
	var list []Cargo
	json.Unmarshal(file, &list)
	c.ID = len(list) + 1
	list = append(list, c)
	data, _ := json.MarshalIndent(list, "", "  ")
	os.WriteFile("data/cargo.json", data, 0644)
}

func sendCargoPage(bot *tgbotapi.BotAPI, chatID int64, page int) {
	cargos := loadCargos()

	if len(cargos) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Yuklar ro'yxati hozircha bo'sh."))
		return
	}

	pageSize := 10
	start := page * pageSize
	end := start + pageSize

	if start >= len(cargos) {
		bot.Send(tgbotapi.NewMessage(chatID, "Bunday sahifa yo'q."))
		return
	}

	if end > len(cargos) {
		end = len(cargos)
	}

	text := "📋 Yuklar ro'yxati:\n\n"
	for i := start; i < end; i++ {
		c := cargos[i]
		text += fmt.Sprintf("🆔 %d\n📅 %s\n💰 %d\n\n", c.ID, c.Date, c.Amount)
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	if page > 0 {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("⬅️ Oldingi", fmt.Sprintf("prev_%d", page-1)))
	}
	if end < len(cargos) {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("➡️ Keyingi", fmt.Sprintf("next_%d", page+1)))
	}
	if len(row) > 0 {
		buttons = append(buttons, row)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	if len(buttons) > 0 {
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
	}
	bot.Send(msg)
}

func showCargoByID(bot *tgbotapi.BotAPI, chatID int64, idText string) {
	id, err := strconv.Atoi(idText)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "ID noto‘g‘ri kiritildi"))
		return
	}
	file, err := os.ReadFile("data/cargo.json")
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Fayl topilmadi"))
		return
	}
	var list []Cargo
	json.Unmarshal(file, &list)
	for _, c := range list {
		if c.ID == id {
			caption := "📦 Yuk ma'lumotlari:\n" +
				"ID: " + strconv.Itoa(c.ID) + "\n" +
				"📅 Sana: " + c.Date + "\n" +
				"💵 Summa: " + strconv.Itoa(c.Amount)
			if c.Photo != "" {
				photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(c.Photo))
				photo.Caption = caption
				bot.Send(photo)
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, caption))
			}
			return
		}
	}
	bot.Send(tgbotapi.NewMessage(chatID, "Yuk topilmadi."))
}

func loadCargos() []Cargo {
	file, err := os.ReadFile("data/cargo.json")
	if err != nil {
		return []Cargo{}
	}
	var list []Cargo
	json.Unmarshal(file, &list)
	return list
}
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
	msg := tgbotapi.NewMessage(chatID, "📦 Yangi yuk qo'shish.\n\nNecha so'mlik yuk qo'shmoqchisiz?")
	bot.Send(msg)
}

func saveCargo(chatID int64, message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	c := cargoTemp[chatID]
	if c == nil {
		return
	}
	if c.Amount == 0 {
		amt, err := strconv.Atoi(message.Text)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❗ Summani faqat son bilan kiriting."))
			return
		}
		c.Amount = amt
		c.Date = time.Now().Format("02-01-2006")
		bot.Send(tgbotapi.NewMessage(chatID, "📷 Endi yuk rasmini yuboring."))
		return
	}
	if c.Photo == "" && message.Photo != nil {
		photo := message.Photo[len(message.Photo)-1]
		c.Photo = photo.FileID
		appendCargo(*c)
		delete(cargoTemp, chatID)
		bot.Send(tgbotapi.NewMessage(chatID, "✅ Yuk saqlandi! Rahmat."))
		return
	}

	if c.Photo == "" && message.Photo == nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Iltimos, yukning rasmini yuboring."))
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
		bot.Send(tgbotapi.NewMessage(chatID, "📦 Hozircha yuklar mavjud emas."))
		return
	}

	pageSize := 10
	start := page * pageSize
	end := start + pageSize
	if start >= len(cargos) {
		bot.Send(tgbotapi.NewMessage(chatID, "Bunday sahifa mavjud emas."))
		return
	}
	if end > len(cargos) {
		end = len(cargos)
	}

	text := "📋 *Yuklar ro'yxati:*\n\n"
	for i := start; i < end; i++ {
		c := cargos[i]
		text += fmt.Sprintf("`#%d` | %s | %d so'm\n", c.ID, c.Date, c.Amount)
	}
	text += fmt.Sprintf("\nSahifa: %d", page+1)

	var buttons [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	if page > 0 {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("⬅️ Oldingi", "prev_cargo"))
	}
	if end < len(cargos) {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("➡️ Keyingi", "next_cargo"))
	}
	if len(row) > 0 {
		buttons = append(buttons, row)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if len(buttons) > 0 {
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
	}
	bot.Send(msg)
}

func showCargoByID(bot *tgbotapi.BotAPI, chatID int64, idText string) {
	id, err := strconv.Atoi(idText)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❗ ID noto‘g‘ri kiritildi."))
		return
	}
	file, err := os.ReadFile("data/cargo.json")
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ma'lumotlar fayli topilmadi."))
		return
	}
	var list []Cargo
	json.Unmarshal(file, &list)
	for _, c := range list {
		if c.ID == id {
			caption := fmt.Sprintf("*Yuk ma'lumotlari:*\n\nID: %d\nSana: %s\nSumma: %d so'm", c.ID, c.Date, c.Amount)
			if c.Photo != "" {
				photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(c.Photo))
				photo.Caption = caption
				photo.ParseMode = "Markdown"
				bot.Send(photo)
			} else {
				msg := tgbotapi.NewMessage(chatID, caption)
				msg.ParseMode = "Markdown"
				bot.Send(msg)
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
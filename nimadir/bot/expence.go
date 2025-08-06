package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Expense struct {
	ID      int    `json:"id"`
	Date    string `json:"date"`
	Amount  int    `json:"amount"`
	Photo   string `json:"photo"`
	Contact string `json:"contact"`
}

var expenseTemp = make(map[int64]*Expense)

func startAddExpense(bot *tgbotapi.BotAPI, chatID int64) {
	expenseTemp[chatID] = &Expense{}
	msg := tgbotapi.NewMessage(chatID, "💸 Yangi chiqim.\n\nNecha so'm chiqim bo'ldi?")
	bot.Send(msg)
}

func saveExpense(chatID int64, message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	e := expenseTemp[chatID]
	if e == nil {
		return
	}
	if e.Amount == 0 {
		amt, err := strconv.Atoi(message.Text)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❗ Summani faqat son bilan kiriting."))
			return
		}
		e.Amount = amt
		e.Date = time.Now().Format("02-01-2006")
		bot.Send(tgbotapi.NewMessage(chatID, "📱 Telefon raqamni kiriting:"))
		return
	}
	if e.Contact == "" && message.Photo == nil {
		e.Contact = message.Text
		bot.Send(tgbotapi.NewMessage(chatID, "📷 Chek rasmini yuboring:"))
		return
	}
	if e.Photo == "" && message.Photo != nil {
		photo := message.Photo[len(message.Photo)-1]
		e.Photo = photo.FileID
		appendExpense(*e)
		delete(expenseTemp, chatID)
		bot.Send(tgbotapi.NewMessage(chatID, "✅ Chiqim saqlandi! Rahmat."))
		return
	}

	if e.Photo == "" && message.Photo == nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Iltimos, chek rasmini yuboring."))
	}
}

func appendExpense(e Expense) {
	file, _ := os.ReadFile("data/expense.json")
	var list []Expense
	json.Unmarshal(file, &list)
	e.ID = len(list) + 1
	list = append(list, e)
	data, _ := json.MarshalIndent(list, "", "  ")
	os.WriteFile("data/expense.json", data, 0644)
}

func sendExpensePage(bot *tgbotapi.BotAPI, chatID int64, page int) {
	expenses := loadExpenses()

	if len(expenses) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "💸 Hozircha chiqimlar mavjud emas."))
		return
	}

	pageSize := 10
	start := page * pageSize
	end := start + pageSize
	if start >= len(expenses) {
		bot.Send(tgbotapi.NewMessage(chatID, "Bunday sahifa mavjud emas."))
		return
	}
	if end > len(expenses) {
		end = len(expenses)
	}

	text := "📋 *Chiqimlar ro'yxati:*\n\n"
	for i := start; i < end; i++ {
		e := expenses[i]
		text += fmt.Sprintf("`#%d` | %s | %d so'm | %s\n", e.ID, e.Date, e.Amount, e.Contact)
	}
	text += fmt.Sprintf("\nSahifa: %d", page+1)

	var buttons [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	if page > 0 {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("⬅️ Oldingi", "prev_expense"))
	}
	if end < len(expenses) {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("➡️ Keyingi", "next_expense"))
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

func showExpenseByID(bot *tgbotapi.BotAPI, chatID int64, idText string) {
	id, err := strconv.Atoi(idText)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❗ ID noto‘g‘ri kiritildi."))
		return
	}
	file, err := os.ReadFile("data/expense.json")
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ma'lumotlar fayli topilmadi."))
		return
	}
	var list []Expense
	json.Unmarshal(file, &list)
	for _, e := range list {
		if e.ID == id {
			caption := fmt.Sprintf("*Chiqim ma'lumotlari:*\n\nID: %d\nSana: %s\nSumma: %d so'm\nTel: %s",
				e.ID, e.Date, e.Amount, e.Contact)
			if e.Photo != "" {
				photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(e.Photo))
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
	bot.Send(tgbotapi.NewMessage(chatID, "Chiqim topilmadi."))
}

func loadExpenses() []Expense {
	file, err := os.ReadFile("data/expense.json")
	if err != nil {
		return []Expense{}
	}
	var list []Expense
	json.Unmarshal(file, &list)
	return list
}

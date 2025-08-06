package bot

import (
	"fmt"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var pageMap = map[int64]int{}
var searchMode = map[int64]string{}
var confirmClearMsgID = map[int64]int{}

func HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message != nil {
		chatID := update.Message.Chat.ID
		text := update.Message.Text

		if update.Message.IsCommand() {
			if update.Message.Command() == "start" {
				msg := tgbotapi.NewMessage(chatID,
					"👋 *Salom!* Botga xush kelibsiz!\n\nQuyidagi bo'limlardan birini tanlang:",
				)
				msg.ParseMode = "Markdown"
				msg.ReplyMarkup = mainMenu()
				bot.Send(msg)
			}
			return
		}

		if mode, ok := searchMode[chatID]; ok && mode != "" {
			if mode == "cargo" {
				showCargoByID(bot, chatID, text)
			} else if mode == "expense" {
				showExpenseByID(bot, chatID, text)
			}
			searchMode[chatID] = ""
			deleteMessage(bot, chatID, update.Message.MessageID)
			return
		}

		if cargoTemp[chatID] != nil {
			saveCargo(chatID, update.Message, bot)
			return
		}

		if expenseTemp[chatID] != nil {
			saveExpense(chatID, update.Message, bot)
			return
		}

		switch text {
		case "📊 Statistika":
			stats := getStatistics()
			profitValue := stats.TotalExpense - stats.TotalCargo
			var profitText string
			if profitValue >= 0 {
				profitText = fmt.Sprintf("🟢 Foyda: *%s so'm*", formatNumber(profitValue))
			} else {
				profitText = fmt.Sprintf("🔴 Zarar: *%s so'm*", formatNumber(-profitValue))
			}
			text := fmt.Sprintf(
				"📊 *Statistika:*\n\n"+
					"💰 Umumiy kirim: *%s so'm*\n"+
					"💸 Umumiy chiqim: *%s so'm*\n"+
					"%s\n\n"+
					"Oxirgi 7 kunlik hisobot uchun *📄 PDF* yuklab olishingiz mumkin.",
				formatNumber(stats.TotalCargo),
				formatNumber(stats.TotalExpense),
				profitText,
			)
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("📄 PDF Yuklab olish", "download_pdf"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("📊 Grafik", "show_chart"),
				),
			)
			bot.Send(msg)

		case "📦 Kirgan yuk":
			msg := tgbotapi.NewMessage(chatID, "📦 *Yuk bo'limi* — kerakli amalni tanlang:")
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = cargoMenu()
			bot.Send(msg)
		case "💸 Chiqarilgan pul":
			msg := tgbotapi.NewMessage(chatID, "💸 *Chiqimlar bo'limi* — kerakli amalni tanlang:")
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = expenseMenu()
			bot.Send(msg)
		case "⬅️ Orqaga":
			msg := tgbotapi.NewMessage(chatID, "🔙 Bosh menyu:")
			msg.ReplyMarkup = mainMenu()
			bot.Send(msg)
		case "➕ Yangi yuk":
			startAddCargo(bot, chatID)
		case "📋 Yuklar ro'yxati":
			pageMap[chatID] = 0
			sendCargoPage(bot, chatID, 0)
		case "🔍 Yuk Qidirish":
			searchMode[chatID] = "cargo"
			bot.Send(tgbotapi.NewMessage(chatID, "🔍 Yuk ID ni kiriting:"))
		case "➕ Yangi chiqim":
			startAddExpense(bot, chatID)
		case "📋 Chiqimlar ro'yxati":
			pageMap[chatID] = 0
			sendExpensePage(bot, chatID, 0)
		case "🔍 Chiqim Qidirish":
			searchMode[chatID] = "expense"
			bot.Send(tgbotapi.NewMessage(chatID, "🔍 Chiqim ID ni kiriting:"))
		case "🗑 Ma'lumotlarni tozalash":
			msg := tgbotapi.NewMessage(chatID, "⚠️ *Diqqat!* Barcha yuk, chiqim va fayllar o‘chiriladi. Davom etishni xohlaysizmi?")
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("✅ Ha, tozalansin", "confirm_clear"),
					tgbotapi.NewInlineKeyboardButtonData("❌ Bekor qilish", "cancel_clear"),
				),
			)
			sentMsg, _ := bot.Send(msg)
			confirmClearMsgID[chatID] = sentMsg.MessageID
			deleteMessage(bot, chatID, update.Message.MessageID)
		}
	}

	if update.CallbackQuery != nil {
		chatID := update.CallbackQuery.Message.Chat.ID
		data := update.CallbackQuery.Data

		switch data {
		case "confirm_clear":
			deleteMessage(bot, chatID, confirmClearMsgID[chatID])
			err := clearAllData()
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Ma'lumotlarni tozalashda xatolik yuz berdi!"))
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "✅ Barcha ma'lumotlar tozalandi."))
			}
			delete(confirmClearMsgID, chatID)

		case "cancel_clear":
			deleteMessage(bot, chatID, confirmClearMsgID[chatID])
			bot.Send(tgbotapi.NewMessage(chatID, "❎ Tozalash bekor qilindi."))
			delete(confirmClearMsgID, chatID)

		case "prev_cargo":
			if pageMap[chatID] > 0 {
				pageMap[chatID]--
			}
			sendCargoPage(bot, chatID, pageMap[chatID])

		case "next_cargo":
			pageMap[chatID]++
			sendCargoPage(bot, chatID, pageMap[chatID])

		case "prev_expense":
			if pageMap[chatID] > 0 {
				pageMap[chatID]--
			}
			sendExpensePage(bot, chatID, pageMap[chatID])

		case "next_expense":
			pageMap[chatID]++
			sendExpensePage(bot, chatID, pageMap[chatID])

		case "download_pdf":
			stats := getStatistics()
			fileName, err := generatePDFReport(stats)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ PDF yaratishda xatolik!"))
			} else {
				doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(fileName))
				doc.Caption = "📄 Hisobot PDF shaklida"
				bot.Send(doc)
			}

		case "show_chart":
			stats := getStatistics()
			fileName, err := generateChart(stats)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Grafik yaratishda xatolik!"))
			} else {
				photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(fileName))
				photo.Caption = "📊 Oxirgi 7 kunlik kirim/chiqim diagrammasi"
				bot.Send(photo)
			}
		}

		bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
	}
}

func mainMenu() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 Statistika"),
			tgbotapi.NewKeyboardButton("📦 Kirgan yuk"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💸 Chiqarilgan pul"),
			tgbotapi.NewKeyboardButton("🗑 Ma'lumotlarni tozalash"),
		),
	)
}

func deleteMessage(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	bot.Request(tgbotapi.NewDeleteMessage(chatID, messageID))
}

func clearAllData() error {
	err1 := os.Remove("cargo.json")
	err2 := os.Remove("expense.json")
	err3 := os.RemoveAll("charts")
	err4 := os.RemoveAll("pdfs")

	if err1 != nil && !os.IsNotExist(err1) {
		return err1
	}
	if err2 != nil && !os.IsNotExist(err2) {
		return err2
	}
	if err3 != nil && !os.IsNotExist(err3) {
		return err3
	}
	if err4 != nil && !os.IsNotExist(err4) {
		return err4
	}

	return nil
}

func cargoMenu() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Yangi yuk"),
			tgbotapi.NewKeyboardButton("📋 Yuklar ro'yxati"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔍 Yuk Qidirish"),
			tgbotapi.NewKeyboardButton("⬅️ Orqaga"),
		),
	)
}

func expenseMenu() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Yangi chiqim"),
			tgbotapi.NewKeyboardButton("📋 Chiqimlar ro'yxati"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔍 Chiqim Qidirish"),
			tgbotapi.NewKeyboardButton("⬅️ Orqaga"),
		),
	)
}

func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	if len(s) > 0 {
		parts = append([]string{s}, parts...)
	}
	return strings.Join(parts, ".")
}

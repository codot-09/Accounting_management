package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var pageMap = map[int64]int{}
var searchMode = map[int64]bool{}

func HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message != nil {
		chatID := update.Message.Chat.ID
		text := update.Message.Text

		if update.Message.IsCommand() {
			if update.Message.Command() == "start" {
				msg := tgbotapi.NewMessage(chatID, "Xush kelibsiz! Tanlang:")
				msg.ReplyMarkup = mainMenu()
				bot.Send(msg)
			}
			return
		}

		if searchMode[chatID] {
			searchMode[chatID] = false
			showCargoByID(bot, chatID, text)
			return
		}

		switch text {
		case "📊 Statistika":
			bot.Send(tgbotapi.NewMessage(chatID, "Statistika: Hozircha yo‘q"))
		case "📦 Kirgan yuk":
			msg := tgbotapi.NewMessage(chatID, "Tanlang:")
			msg.ReplyMarkup = cargoMenu()
			bot.Send(msg)
		case "💸 Chiqarilgan pul":
			bot.Send(tgbotapi.NewMessage(chatID, "Chiqarilgan pul: Hozircha yo‘q"))
		case "⬅️ Orqaga":
			msg := tgbotapi.NewMessage(chatID, "Bosh menyu:")
			msg.ReplyMarkup = mainMenu()
			bot.Send(msg)
		case "➕ Yangi yuk":
			startAddCargo(bot, chatID)
		case "📋 Yuklar ro'yxati":
			pageMap[chatID] = 0
			sendCargoPage(bot, chatID, 0)
		case "🔍 Qidirish":
			searchMode[chatID] = true
			bot.Send(tgbotapi.NewMessage(chatID, "ID ni kiriting:"))
		default:
			saveCargo(chatID, update.Message, bot)
		}
	}

	if update.CallbackQuery != nil {
		chatID := update.CallbackQuery.Message.Chat.ID
		data := update.CallbackQuery.Data

		if data == "prev" {
			if pageMap[chatID] > 0 {
				pageMap[chatID]--
			}
			sendCargoPage(bot, chatID, pageMap[chatID])
		} else if data == "next" {
			pageMap[chatID]++
			sendCargoPage(bot, chatID, pageMap[chatID])
		}
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
		),
	)
}

func cargoMenu() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Yangi yuk"),
			tgbotapi.NewKeyboardButton("📋 Yuklar ro'yxati"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔍 Qidirish"),
			tgbotapi.NewKeyboardButton("⬅️ Orqaga"),
		),
	)
}

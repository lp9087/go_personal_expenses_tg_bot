package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var mainKeyboard = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{
	tgbotapi.NewInlineKeyboardButtonData("Работа с категориями", "workWithCategories"),
	tgbotapi.NewInlineKeyboardButtonData("Работа с расходами", "workWithExpenses"),
})

func (b *Bot) handleCommand(update tgbotapi.Update, updates tgbotapi.UpdatesChannel) {

	switch update.CallbackQuery.Data {
	case "addCategory":
		b.service.createCategory(b, update, updates)
	case "createExpenses":
		b.service.createExpenses(b, update, updates)

	case "myCategories":
		b.service.getCategories(b, update, updates)

	case "mainPage":
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Список команд")
		msg.ReplyMarkup = mainKeyboard
		b.bot.Send(msg)

	case "workWithCategories":
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите действие:")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("Добавить категорию", "addCategory"),
			tgbotapi.NewInlineKeyboardButtonData("Мои категории", "myCategories"),
			tgbotapi.NewInlineKeyboardButtonData("Назад", "mainPage"),
		})
		b.bot.Send(msg)

	case "workWithExpenses":
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите действие:")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("Добавить расход", "createExpenses"),
		})
		b.bot.Send(msg)
	}
}

func (b *Bot) handleMessage(update *tgbotapi.Update) {
	switch update.Message.Text {
	case "/start":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я бот-ассистент. Введите /hello, чтобы получить мое приветствие.")
		b.bot.Send(msg)
	case "/hello":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я бот-ассистент. Как я могу вам помочь?")
		b.bot.Send(msg)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Список команд")
		msg.ReplyMarkup = mainKeyboard
		b.bot.Send(msg)
	}
}

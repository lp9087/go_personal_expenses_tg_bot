package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
)

type CategoryName struct {
	ID    uint
	Name  string
	Owner string
}

type Expenses struct {
	ID         uint
	CategoryID int
	Category   CategoryName
	Amount     int
}

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Ошибка сбора локальных переменных с .env файла. Текст ошибки: %s", err)
	}

	Host := os.Getenv("DB_HOST")
	Port := os.Getenv("DB_PORT")
	User := os.Getenv("DB_USER")
	Password := os.Getenv("DB_PASSWORD")
	Name := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", Host, User, Password, Name, Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Ошибка подключения к БД. Текст ошибки: %s", err)
	}

	db.AutoMigrate(&CategoryName{}, &Expenses{})

	bot, err := tgbotapi.NewBotAPI("5806902616:AAEbGoKlc8oukusym8YDNAXHvDoBvg1noxc")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	mainKeyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Работа с категориями", "workWithCategories"),
		tgbotapi.NewInlineKeyboardButtonData("Работа с расходами", "workWithExpenses"),
	})

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Запускаем функции для обработки сообщений и состояний пользователей

	for update := range updates {

		if update.CallbackQuery != nil {

			switch update.CallbackQuery.Data {
			case "addCategory":
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вы выбрали добавить категорию. Введите название: ")
				msg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
				bot.Send(msg)

				// Ожидаем ответа пользователя
				update = <-updates
				if update.Message != nil {
					// Обрабатываем ответ пользователя
					owner := update.Message.Chat.UserName + update.Message.Chat.FirstName + update.Message.Chat.LastName
					category := &CategoryName{Name: update.Message.Text, Owner: owner}
					db.Create(&category)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Вы ввели название категории: %s. ID в БД: %v", update.Message.Text, category.ID))
					bot.Send(msg)
				}
			case "createExpenses":

				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите категорию:")
				var categories []CategoryName
				ownerQuery := update.CallbackQuery.From.UserName + update.CallbackQuery.From.FirstName + update.CallbackQuery.From.LastName
				db.Where("owner = ?", ownerQuery).Find(&categories)

				var buttons []tgbotapi.InlineKeyboardButton
				for _, category := range categories {
					button := tgbotapi.NewInlineKeyboardButtonData(category.Name, fmt.Sprintf("%v", category.ID))
					buttons = append(buttons, button)
				}

				backButton := tgbotapi.NewInlineKeyboardButtonData("Назад", "mainPage")
				buttons = append(buttons, backButton)

				categoriesKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons)
				msg.ReplyMarkup = categoriesKeyboard
				msg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
				bot.Send(msg)
				// Ожидаем ответа пользователя
				category := <-updates
				var CategoryID int
				if category.CallbackQuery != nil {
					CategoryID, _ = strconv.Atoi(category.CallbackQuery.Data)
				}

				amountMsg := tgbotapi.NewMessage(category.CallbackQuery.Message.Chat.ID, fmt.Sprintf("Введите сумму расхода:"))
				amountMsg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
				bot.Send(amountMsg)

				amount := <-updates
				value, _ := strconv.Atoi(amount.Message.Text)
				expenses := &Expenses{CategoryID: CategoryID, Amount: value}
				db.Create(&expenses)
				resMsg := tgbotapi.NewMessage(amount.Message.Chat.ID, fmt.Sprintf("Вы создали расходю ID в БД: %v", expenses.ID))
				bot.Send(resMsg)

			case "myCategories":
				var categories []CategoryName
				if update.CallbackQuery != nil {
					ownerQuery := update.CallbackQuery.From.UserName + update.CallbackQuery.From.FirstName + update.CallbackQuery.From.LastName
					db.Where("owner = ?", ownerQuery).Find(&categories)

					var buttons []tgbotapi.InlineKeyboardButton
					for _, category := range categories {
						button := tgbotapi.NewInlineKeyboardButtonData(category.Name, fmt.Sprintf("category:%v", category.ID))
						buttons = append(buttons, button)
					}

					backButton := tgbotapi.NewInlineKeyboardButtonData("Назад", "mainPage")
					buttons = append(buttons, backButton)

					categoriesKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons)

					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ваши категории")
					msg.ReplyMarkup = categoriesKeyboard
					bot.Send(msg)
				}

			case "mainPage":
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Список команд")
				msg.ReplyMarkup = mainKeyboard
				bot.Send(msg)

			case "workWithCategories":

				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите действие:")
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonData("Добавить категорию", "addCategory"),
					tgbotapi.NewInlineKeyboardButtonData("Мои категории", "myCategories"),
					tgbotapi.NewInlineKeyboardButtonData("Назад", "mainPage"),
				})
				bot.Send(msg)
			case "workWithExpenses":

				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите действие:")
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonData("Добавить расход", "createExpenses"),
				})
				bot.Send(msg)
			}
		} else if update.Message != nil {
			// обработка обычных сообщений
			switch update.Message.Text {
			case "/start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я бот-ассистент. Введите /hello, чтобы получить мое приветствие.")
				bot.Send(msg)
			case "/hello":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я бот-ассистент. Как я могу вам помочь?")
				bot.Send(msg)
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Список команд")
				msg.ReplyMarkup = mainKeyboard
				bot.Send(msg)
			}
		}

	}

}

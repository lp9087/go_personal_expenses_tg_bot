package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"log"
	"strconv"
	"tg_bot_expenses/pkg/repository/models"
)

type Service struct {
	db *gorm.DB
}

func (c *Service) createCategory(client *Bot, update tgbotapi.Update, updates tgbotapi.UpdatesChannel) {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вы выбрали добавить категорию. Введите название: ")
	msg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
	if _, err := client.bot.Send(msg); err != nil {
		log.Fatalf("Ошибка отправки сообщения. Текст ошибки: %s", err)
	}

	// Ожидаем ответа пользователя
	update = <-updates
	if update.Message != nil {
		// Обрабатываем ответ пользователя
		owner := update.Message.Chat.UserName + update.Message.Chat.FirstName + update.Message.Chat.LastName
		category := &models.CategoryName{Name: update.Message.Text, Owner: owner}
		c.db.Create(&category)
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Вы ввели название категории: %s. ID в БД: %v", update.Message.Text, category.ID))
		if _, err := client.bot.Send(msg); err != nil {
			log.Fatalf("Ошибка отправки сообщения. Текст ошибки: %s", err)
		}
	}
}

func (c *Service) deleteCategory(client *Bot, update tgbotapi.Update, updates tgbotapi.UpdatesChannel) {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите категорию:")
	var categories []models.CategoryName
	ownerQuery := update.CallbackQuery.From.UserName + update.CallbackQuery.From.FirstName + update.CallbackQuery.From.LastName
	c.db.Where("owner = ?", ownerQuery).Find(&categories)

	buttons := c.getCategoriesButtons(update)

	categoriesKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg.ReplyMarkup = categoriesKeyboard
	msg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
	if _, err := client.bot.Send(msg); err != nil {
		log.Fatalf("Ошибка отправки сообщения. Текст ошибки: %s", err)
	}

	category := <-updates
	var CategoryID int
	var categoryError error
	if category.CallbackQuery != nil {
		CategoryID, categoryError = strconv.Atoi(category.CallbackQuery.Data)
		if categoryError != nil {
			c.cancelAction(client, update)
			return
		}
	}
	//var toDelete models.CategoryName
	c.db.Where("id = ?", CategoryID).Delete(&models.CategoryName{})
	resMsg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вы успешно удалили категорию")
	if _, err := client.bot.Send(resMsg); err != nil {
		log.Fatalf("Ошибка отправки сообщения. Текст ошибки: %s", err)
	}
}

func (c *Service) createExpenses(client *Bot, update tgbotapi.Update, updates tgbotapi.UpdatesChannel) {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите категорию:")
	var categories []models.CategoryName
	ownerQuery := update.CallbackQuery.From.UserName + update.CallbackQuery.From.FirstName + update.CallbackQuery.From.LastName
	c.db.Where("owner = ?", ownerQuery).Find(&categories)

	buttons := c.getCategoriesButtons(update)

	categoriesKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg.ReplyMarkup = categoriesKeyboard
	msg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
	if _, err := client.bot.Send(msg); err != nil {
		log.Fatalf("Ошибка отправки сообщения. Текст ошибки: %s", err)
	}

	category := <-updates
	var CategoryID int
	var categoryError error
	if category.CallbackQuery != nil {
		CategoryID, categoryError = strconv.Atoi(category.CallbackQuery.Data)
		if categoryError != nil {
			c.cancelAction(client, update)
			return
		}
	} else {
		c.cancelAction(client, update)
		return
	}

	amountMsg := tgbotapi.NewMessage(category.CallbackQuery.Message.Chat.ID, fmt.Sprintf("Введите сумму расхода:"))
	amountMsg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
	if _, err := client.bot.Send(amountMsg); err != nil {
		log.Fatalf("Ошибка отправки сообщения. Текст ошибки: %s", err)
	}

	amount := <-updates
	value, _ := strconv.Atoi(amount.Message.Text)
	expenses := &models.Expenses{CategoryID: CategoryID, Amount: value}
	c.db.Create(&expenses)
	resMsg := tgbotapi.NewMessage(amount.Message.Chat.ID, fmt.Sprintf("Вы создали расходю ID в БД: %v", expenses.ID))
	if _, err := client.bot.Send(resMsg); err != nil {
		log.Fatalf("Ошибка отправки сообщения. Текст ошибки: %s", err)
	}
}

func (c *Service) getExpenses(client *Bot, update tgbotapi.Update) {

	type ExpenseWithCategoryName struct {
		Amount float64
		Name   string
	}

	var expenses []ExpenseWithCategoryName

	if update.CallbackQuery != nil {
		ownerQuery := update.CallbackQuery.From.UserName + update.CallbackQuery.From.FirstName + update.CallbackQuery.From.LastName
		c.db.Model(&models.Expenses{}).
			Select("expenses.amount, category_names.name").
			Joins("left join category_names on category_names.id = expenses.category_id").
			Where("category_names.owner = ?", ownerQuery).
			Find(&expenses)

		var buttons [][]tgbotapi.InlineKeyboardButton
		if len(expenses) == 0 {
			button := tgbotapi.NewInlineKeyboardButtonData("Расходы отсутствуют", "Ваши расходы")
			row := []tgbotapi.InlineKeyboardButton{button}
			buttons = append(buttons, row)
		}
		for _, consumption := range expenses {
			button := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Категория: %v | Количество: %v", consumption.Name, consumption.Amount), "Ваши расходы")
			row := []tgbotapi.InlineKeyboardButton{button}
			buttons = append(buttons, row)
		}

		categoriesKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ваши расходы")
		msg.ReplyMarkup = categoriesKeyboard
		if _, err := client.bot.Send(msg); err != nil {
			log.Fatalf("Ошибка отправки сообщения. Текст ошибки: %s", err)
		}
	}
}

func (c *Service) deleteExpense(client *Bot, update tgbotapi.Update, updates tgbotapi.UpdatesChannel) {

	type ExpenseWithCategoryName struct {
		Id     uint
		Amount float64
		Name   string
	}

	var expenses []ExpenseWithCategoryName

	if update.CallbackQuery != nil {
		ownerQuery := update.CallbackQuery.From.UserName + update.CallbackQuery.From.FirstName + update.CallbackQuery.From.LastName
		c.db.Model(&models.Expenses{}).
			Select("expenses.id, expenses.amount, category_names.name").
			Joins("left join category_names on category_names.id = expenses.category_id").
			Where("category_names.owner = ?", ownerQuery).
			Find(&expenses)

		var buttons [][]tgbotapi.InlineKeyboardButton
		if len(expenses) == 0 {
			button := tgbotapi.NewInlineKeyboardButtonData("Расходы отсутствуют", "Ваши расходы")
			row := []tgbotapi.InlineKeyboardButton{button}
			buttons = append(buttons, row)
		}
		for _, consumption := range expenses {
			button := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Категория: %v | Количество: %v", consumption.Name, consumption.Amount), fmt.Sprintf("%v", consumption.Id))
			row := []tgbotapi.InlineKeyboardButton{button}
			buttons = append(buttons, row)
		}

		categoriesKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Какой расход удалить")
		msg.ReplyMarkup = categoriesKeyboard
		if _, err := client.bot.Send(msg); err != nil {
			log.Fatalf("Ошибка отправки сообщения. Текст ошибки: %s", err)
		}
		expenses := <-updates
		var expenseID int
		var expenseError error

		if expenses.CallbackQuery != nil {
			expenseID, expenseError = strconv.Atoi(expenses.CallbackQuery.Data)
			if expenseError != nil {
				c.cancelAction(client, update)
				return
			}
		} else {
			c.cancelAction(client, update)
			return
		}
		c.db.Where("id = ?", expenseID).Delete(&models.Expenses{})
		resMsg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вы успешно удалили расход")
		if _, err := client.bot.Send(resMsg); err != nil {
			log.Fatalf("Ошибка отправки сообщения. Текст ошибки: %s", err)
		}
	}
}

func (c *Service) getCategories(client *Bot, update tgbotapi.Update) {

	buttons := c.getCategoriesButtons(update)

	categoriesKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ваши категории")
	msg.ReplyMarkup = categoriesKeyboard
	if _, err := client.bot.Send(msg); err != nil {
		log.Fatalf("Ошибка отправки сообщения. Текст ошибки: %s", err)
	}
}

func (c *Service) getCategoriesButtons(update tgbotapi.Update) [][]tgbotapi.InlineKeyboardButton {
	var categories []models.CategoryName
	var buttons [][]tgbotapi.InlineKeyboardButton

	if update.CallbackQuery != nil {
		ownerQuery := update.CallbackQuery.From.UserName + update.CallbackQuery.From.FirstName + update.CallbackQuery.From.LastName
		c.db.Where("owner = ?", ownerQuery).Find(&categories)

		for _, category := range categories {
			button := tgbotapi.NewInlineKeyboardButtonData(category.Name, fmt.Sprintf("%v", category.ID))
			row := []tgbotapi.InlineKeyboardButton{button}
			buttons = append(buttons, row)
		}

		backButton := tgbotapi.NewInlineKeyboardButtonData("Назад", "mainPage")
		row := []tgbotapi.InlineKeyboardButton{backButton}
		buttons = append(buttons, row)
	}
	return buttons
}

func (c *Service) cancelAction(client *Bot, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Действие отменено")
	if _, err := client.bot.Send(msg); err != nil {
		log.Fatalf("Ошибка отправки сообщения. Текст ошибки: %s", err)
	}

	msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Главное меню")
	msg.ReplyMarkup = mainKeyboard
	if _, err := client.bot.Send(msg); err != nil {
		log.Fatalf("Ошибка отправки сообщения. Текст ошибки: %s", err)
	}
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

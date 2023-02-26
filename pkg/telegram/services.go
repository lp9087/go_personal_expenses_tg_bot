package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"strconv"
	"tg_bot_expenses/pkg/repository/models"
)

type Service struct {
	db *gorm.DB
}

func (c *Service) createCategory(client *Bot, update tgbotapi.Update, updates tgbotapi.UpdatesChannel) {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вы выбрали добавить категорию. Введите название: ")
	msg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
	client.bot.Send(msg)

	// Ожидаем ответа пользователя
	update = <-updates
	if update.Message != nil {
		// Обрабатываем ответ пользователя
		owner := update.Message.Chat.UserName + update.Message.Chat.FirstName + update.Message.Chat.LastName
		category := &models.CategoryName{Name: update.Message.Text, Owner: owner}
		c.db.Create(&category)
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Вы ввели название категории: %s. ID в БД: %v", update.Message.Text, category.ID))
		client.bot.Send(msg)
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
	client.bot.Send(msg)

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
	client.bot.Send(resMsg)
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
	client.bot.Send(msg)

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
	client.bot.Send(amountMsg)

	amount := <-updates
	value, _ := strconv.Atoi(amount.Message.Text)
	expenses := &models.Expenses{CategoryID: CategoryID, Amount: value}
	c.db.Create(&expenses)
	resMsg := tgbotapi.NewMessage(amount.Message.Chat.ID, fmt.Sprintf("Вы создали расходю ID в БД: %v", expenses.ID))
	client.bot.Send(resMsg)
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
		for _, consumption := range expenses {
			button := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Категория: %v | Количество: %v", consumption.Name, consumption.Amount), "Ваши расходы")
			row := []tgbotapi.InlineKeyboardButton{button}
			buttons = append(buttons, row)
		}

		categoriesKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ваши расходы")
		msg.ReplyMarkup = categoriesKeyboard
		client.bot.Send(msg)
	}
}

func (c *Service) getCategories(client *Bot, update tgbotapi.Update) {

	buttons := c.getCategoriesButtons(update)

	categoriesKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ваши категории")
	msg.ReplyMarkup = categoriesKeyboard
	client.bot.Send(msg)
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
	client.bot.Send(msg)

	msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Главное меню")
	msg.ReplyMarkup = mainKeyboard
	client.bot.Send(msg)
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

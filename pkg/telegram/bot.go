package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot     *tgbotapi.BotAPI
	service *Service
}

func NewBot(bot *tgbotapi.BotAPI, service *Service) *Bot {
	return &Bot{
		bot:     bot,
		service: service,
	}
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)

	for update := range updates {
		// Handle commands
		if update.CallbackQuery != nil {
			b.handleCommand(update, updates)
			continue
		}
		if update.Message != nil { // ignore any non-Message Updates
			b.handleMessage(&update)
		}

	}

	return nil
}

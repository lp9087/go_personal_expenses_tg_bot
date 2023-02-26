package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"tg_bot_expenses/pkg/repository/models"
	"tg_bot_expenses/pkg/telegram"
)

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

	db.AutoMigrate(&models.CategoryName{}, &models.Expenses{})

	bot, err := tgbotapi.NewBotAPI("5806902616:AAEbGoKlc8oukusym8YDNAXHvDoBvg1noxc")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	service := telegram.NewService(db)

	client := telegram.NewBot(bot, service)

	if err := client.Start(); err != nil {
		log.Fatal(err)
	}

}

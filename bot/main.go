package main

import (
	"aaa/config"
	"aaa/db"
	"aaa/handler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func main() {
	botToken := config.BotToken
	weatherApiKey := config.WeatherApiKey

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Авторизован как %s", bot.Self.UserName)

	database, err := db.InitDB()
	if err != nil {
		log.Panic(err)
	}
	defer database.Close()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			cmd := update.Message.Command()

			switch cmd {
			case "start":
				handler.StartCommand(bot, chatID)
			case "help":
				handler.HelpCommand(bot, chatID)
			case "weather":
				handler.WeatherCommand(bot, chatID)
			case "savecity":
				handler.SaveCityCommand(bot, chatID)
			case "mycity":
				handler.MyCityCommand(bot, chatID, database, weatherApiKey)
			default:
				handler.DefaultCommand(bot, chatID)
			}
		}

		if update.CallbackQuery != nil {
			handler.CallbackQueryHandler(bot, update.CallbackQuery, database, weatherApiKey)
		}
	}
}

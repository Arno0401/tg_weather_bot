package handler

import (
	"aaa/citiesmap"
	"aaa/weather"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"sort"
	"strings"
)

func StartCommand(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Привет! Введи /help для списка команд.")
	menu := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/weather"),
			tgbotapi.NewKeyboardButton("/savecity"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/mycity"),
			tgbotapi.NewKeyboardButton("/help"),
		),
	)
	msg.ReplyMarkup = menu
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Error ", err)
		return
	}
}

func HelpCommand(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Команды:\n/weather — узнать погоду\n/savecity — сохранить город\n/mycity — погода в сохранённом городе")
	_, err := bot.Send(msg)
	if err != nil {
		return
	}
}
func WeatherCommand(bot *tgbotapi.BotAPI, chatID int64) {
	showKeys(bot, chatID, "weather_")
}

func SaveCityCommand(bot *tgbotapi.BotAPI, chatID int64) {
	showKeys(bot, chatID, "savecity_")
}

func showKeys(bot *tgbotapi.BotAPI, chatID int64, prefix string) {
	keys := make([]string, 0, len(citiesmap.CitiesMap))
	for key := range citiesmap.CitiesMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	for i, key := range keys {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(key, fmt.Sprintf("%s%s", prefix, key)))
		if (i+1)%5 == 0 {
			inlineKeyboard = append(inlineKeyboard, row)
			row = nil
		}
	}

	if len(row) > 0 {
		inlineKeyboard = append(inlineKeyboard, row)
	}

	msg := tgbotapi.NewMessage(chatID, "Выберите первую букву Города:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)
	_, err := bot.Send(msg)
	if err != nil {
		return
	}
}

func CallbackQueryHandler(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB, weatherApiKey string) {
	data := callbackQuery.Data
	if len(data) > 13 && data[:14] == "savecity_city_" {
		log.Println("------------------------------", data)
	}
	chatID := callbackQuery.Message.Chat.ID

	switch {
	case len(data) > 13 && data[:14] == "savecity_city_":
		city := data[14:]
		if cities, ok := citiesmap.CitiesMap[city]; !ok {
			showCities(bot, chatID, cities, "")
		} else {
			_, err := bot.Send(tgbotapi.NewMessage(chatID, "Ключ не найден!"))
			if err != nil {
				return
			}
		}
		if err := saveCityToDB(db, chatID, city); err != nil {
			_, err := bot.Send(tgbotapi.NewMessage(chatID, "Не удалось сохранить город."))
			if err != nil {
				return
			}
		} else {
			_, err := bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Город %s сохранён.", city)))
			if err != nil {
				return
			}
		}
	case len(data) > 10 && data[:9] == "savecity_":
		key := data[9:]
		key = strings.Trim(key, " ")
		if cities, ok := citiesmap.CitiesMap[key]; ok {
			showCities(bot, chatID, cities, "savecity_city_")
		} else {
			_, err := bot.Send(tgbotapi.NewMessage(chatID, "Ключ не найден!"))
			if err != nil {
				return
			}
		}

	case len(data) > 8 && data[:8] == "weather_":
		key := data[8:]
		if cities, ok := citiesmap.CitiesMap[key]; ok {
			showCities(bot, chatID, cities, "city_")
		} else {
			_, err := bot.Send(tgbotapi.NewMessage(chatID, "Ключ не найден!"))
			if err != nil {
				return
			}
		}

	case len(data) > 5 && data[:5] == "city_":
		city := data[5:]
		weatherInfo, err := weather.GetWeather(city, weatherApiKey)
		if err != nil {
			weatherInfo = "Не удалось получить данные о погоде."
		}
		_, err = bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Погода в %s:\n%s", city, weatherInfo)))
		if err != nil {
			return
		}

	default:
		_, err := bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная ошибка"))
		if err != nil {
			return
		}
	}
}

func showCities(bot *tgbotapi.BotAPI, chatID int64, cities []string, prefix string) {
	var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	for i, city := range cities {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(city, fmt.Sprintf("%s%s", prefix, city)))
		if (i+1)%5 == 0 {
			inlineKeyboard = append(inlineKeyboard, row)
			row = nil
		}
	}

	if len(row) > 0 {
		inlineKeyboard = append(inlineKeyboard, row)
	}

	msg := tgbotapi.NewMessage(chatID, "Выберите город:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)
	_, err := bot.Send(msg)
	if err != nil {
		return
	}
}

func saveCityToDB(db *sql.DB, chatID int64, city string) error {
	query := `
		INSERT INTO users (chat_id, city) VALUES ($1, $2)
		ON CONFLICT (chat_id) DO UPDATE SET city = EXCLUDED.city;
	`
	_, err := db.Exec(query, chatID, city)
	return err
}

func MyCityCommand(bot *tgbotapi.BotAPI, chatID int64, db *sql.DB, weatherApiKey string) {
	city, err := getCityFromDB(db, chatID)
	if err != nil || city == "" {
		_, err := bot.Send(tgbotapi.NewMessage(chatID, "Для начала сохраните ваш город с помощью команды /savecity."))
		if err != nil {
			return
		}
		return
	}

	weatherInfo, err := weather.GetWeather(city, weatherApiKey)
	if err != nil {
		weatherInfo = "Не удалось получить данные о погоде."
	}
	_, err = bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Погода в %s:\n%s", city, weatherInfo)))
	if err != nil {
		return
	}
}

func getCityFromDB(db *sql.DB, chatID int64) (string, error) {
	var city string
	query := "SELECT city FROM users WHERE chat_id = $1"
	err := db.QueryRow(query, chatID).Scan(&city)
	if err != nil {
		return "", err
	}
	return city, nil
}

func DefaultCommand(bot *tgbotapi.BotAPI, chatID int64) {
	_, err := bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда. Введи /help для списка команд."))
	if err != nil {
		return
	}
}

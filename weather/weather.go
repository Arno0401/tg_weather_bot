package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type WeatherResponse struct {
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
}

func GetWeather(city string, apiKey string) (string, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&lang=ru&appid=%s", city, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var weatherData WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		return "", err
	}
	result := fmt.Sprintf("Погода в %s: %.2f°C, %s", city, weatherData.Main.Temp, weatherData.Weather[0].Description)
	return result, nil
}

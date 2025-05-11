package main

import (
	"log"

	"github.com/Powdersumm/Yandexlmsfinalproject/internal/application"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Предупреждение: .env файл не найден, переменные окружения не загружены")
	}
}

func main() {
	app := application.New()
	app.RunServer()

}

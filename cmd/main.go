package main

import (
	"log"

	"github.com/Powdersumm/Yandexlmsfinalproject/internal/application"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		log.Fatal("Ошибка загрузки .env файла:", err)
	}

	// Создание приложения
	app := application.New()

	// Запуск сервера
	if err := app.RunServer(); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}

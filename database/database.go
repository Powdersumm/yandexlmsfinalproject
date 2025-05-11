package database

import (
	"log"
	"os"

	"github.com/Powdersumm/Yandexlmsfinalproject/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DB_DSN")

	// Логируем строку подключения
	log.Println("Подключение к базе данных:", dsn)

	// Подключение к базе данных через GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err) // Вместо panic используем log.Fatal
	}

	// Автоматическое создание таблиц, если их нет
	if err := db.AutoMigrate(&models.User{}, &models.Expression{}); err != nil {
		log.Fatal("Ошибка миграции базы данных:", err)
	}

	DB = db
	log.Println("Успешное подключение к базе данных!") // Сообщение об успешном подключении
}

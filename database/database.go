package database

import (
	"fmt"
	"log"
	"os"

	"github.com/Powdersumm/Yandexlmsfinalproject/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() error {
	// Формируем DSN из переменных окружения
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	log.Println("Подключение к базе данных:", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("ошибка подключения: %w", err)
	}

	// Включаем расширение для UUID и выполняем миграции
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	if err := db.AutoMigrate(&models.User{}, &models.Expression{}); err != nil {
		return fmt.Errorf("ошибка миграции: %w", err)
	}

	DB = db
	log.Println("Успешное подключение!")
	return nil
}

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

func Connect() error { // Теперь функция возвращает error
	dsn := os.Getenv("DB_DSN")
	log.Println("Подключение к базе данных:", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %w", err) // Возвращаем ошибку вместо log.Fatal
	}

	if err := db.AutoMigrate(&models.User{}, &models.Expression{}); err != nil {
		return fmt.Errorf("ошибка миграции базы данных: %w", err)
	}

	DB = db
	log.Println("Успешное подключение к базе данных!")
	return nil // Возвращаем nil при успешном подключении
}

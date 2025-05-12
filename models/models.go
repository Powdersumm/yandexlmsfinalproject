package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Login        string `gorm:"unique"`
	PasswordHash string
}

type Expression struct {
	gorm.Model
	ID         string `gorm:"type:uuid;primaryKey"` // Используйте UUID
	UserID     uint   `gorm:"index"`                // Связь с пользователем
	Expression string `gorm:"not null"`
	Status     string `gorm:"default:'pending'"` // Статусы: pending/completed/error
	Result     float64
}

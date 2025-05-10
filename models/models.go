package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Login        string `gorm:"unique"`
	PasswordHash string
}

type Expression struct {
	gorm.Model
	UserID     uint
	Expression string
	Status     string
	Result     float64
}

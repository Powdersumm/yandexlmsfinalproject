package models

import (
	"time"
)

type Expression struct {
	ID         string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID     uint   `gorm:"index"`
	Expression string
	Status     string
	Result     float64
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

type ExpressionTask struct {
	ID         string
	Expression string
	UserID     uint
}

package models

import (
	"time"
)

type Expression struct {
	ID         string `gorm:"primaryKey"`
	UserID     uint   `gorm:"index"`
	Expression string
	Status     string // pending, processing, completed, error
	Result     float64
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

type ExpressionTask struct {
	ID         string
	Expression string
	UserID     uint
}

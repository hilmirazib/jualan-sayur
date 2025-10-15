package model

import "time"

type VerificationToken struct {
	ID        int64 `gorm:"primaryKey"`
	UserID    int64 `gorm:"index"`
	Token     string
	TokenType string
	ExpiresAt time.Time
	User      User `gorm:"foreignKey:UserID"`
}

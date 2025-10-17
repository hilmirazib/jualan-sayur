package model

import "time"

type BlacklistToken struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	TokenHash  string    `gorm:"column:token_hash;type:varchar(256);not null"`
	ExpiresAt  time.Time `gorm:"column:expires_at;type:timestamp;not null"`
	CreatedAt  time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP"`
}

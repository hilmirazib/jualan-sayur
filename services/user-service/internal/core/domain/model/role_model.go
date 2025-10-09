package model

import "time"

type Role struct {
	ID        int64  `gorm:"PrimaryKey"`
	Name      string `gorm:"unique"`
	Users     []User `gorm:"many2many:user_role;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

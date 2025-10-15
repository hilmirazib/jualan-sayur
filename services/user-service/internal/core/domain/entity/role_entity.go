package entity

import "time"

type RoleEntity struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

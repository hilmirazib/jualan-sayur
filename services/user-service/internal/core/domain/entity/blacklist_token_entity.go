package entity

import "time"

type BlacklistTokenEntity struct {
	ID         int64
	TokenHash  string
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

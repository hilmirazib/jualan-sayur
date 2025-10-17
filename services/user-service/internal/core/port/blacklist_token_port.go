package port

import (
	"context"
)

type BlacklistTokenInterface interface {
	AddToBlacklist(ctx context.Context, tokenHash string, expiresAt int64) error
	IsTokenBlacklisted(ctx context.Context, tokenHash string) bool
}

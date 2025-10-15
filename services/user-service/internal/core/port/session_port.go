package port

import (
	"context"
	"user-service/internal/core/domain/entity"
)

type SessionInterface interface {
	StoreToken(ctx context.Context, userID int64, sessionID string, token string) error
	GetToken(ctx context.Context, userID int64, sessionID string) (string, error)
	DeleteToken(ctx context.Context, userID int64, sessionID string) error
	DeleteAllUserTokens(ctx context.Context, userID int64) error
	ValidateToken(ctx context.Context, userID int64, sessionID string, token string) bool
	GetUserSessions(ctx context.Context, userID int64) ([]entity.SessionInfo, error)
}

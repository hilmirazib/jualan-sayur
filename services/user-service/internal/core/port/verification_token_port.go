package port

import (
	"context"
	"user-service/internal/core/domain/entity"
)

type VerificationTokenInterface interface {
	CreateVerificationToken(ctx context.Context, token *entity.VerificationTokenEntity) error
	GetVerificationToken(ctx context.Context, token string) (*entity.VerificationTokenEntity, error)
	DeleteVerificationToken(ctx context.Context, token string) error
}

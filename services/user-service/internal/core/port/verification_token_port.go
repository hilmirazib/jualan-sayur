package port

import (
	"context"
	"user-service/internal/core/domain/entity"
)

type VerificationTokenInterface interface {
	CreateVerificationToken(ctx context.Context, token *entity.VerificationTokenEntity) error
}

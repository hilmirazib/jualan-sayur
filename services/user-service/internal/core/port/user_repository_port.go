package port

import (
	"context"
	"user-service/internal/core/domain/entity"
)

type UserRepositoryInterface interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error)
}

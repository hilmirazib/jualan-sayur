package port

import (
	"context"
	"user-service/internal/core/domain/entity"
)

type UserServiceInterface interface {
	SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error)
}

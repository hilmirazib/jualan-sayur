package port

import (
	"context"
	"user-service/internal/core/domain/entity"
)

type UserRepositoryInterface interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error)
	GetUserByEmailIncludingUnverified(ctx context.Context, email string) (*entity.UserEntity, error)
	CreateUser(ctx context.Context, user *entity.UserEntity) (*entity.UserEntity, error)
	GetRoleByName(ctx context.Context, name string) (*entity.RoleEntity, error)
	UpdateUserVerificationStatus(ctx context.Context, userID int64, isVerified bool) error
}

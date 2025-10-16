package port

import (
	"context"
	"user-service/internal/core/domain/entity"
)

type UserServiceInterface interface {
	SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error)
	CreateUserAccount(ctx context.Context, email, name, password, passwordConfirmation string) error
	VerifyUserAccount(ctx context.Context, token string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword, passwordConfirmation string) error
}

package port

import (
	"context"
	"io"
	"user-service/internal/core/domain/entity"
)

type UserServiceInterface interface {
	SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error)
	CreateUserAccount(ctx context.Context, email, name, password, passwordConfirmation string) error
	VerifyUserAccount(ctx context.Context, token string) error
	VerifyEmailChange(ctx context.Context, token string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword, passwordConfirmation string) error
	Logout(ctx context.Context, userID int64, sessionID, tokenString string, tokenExpiresAt int64) error
	GetProfile(ctx context.Context, userID int64) (*entity.UserEntity, error)
	UploadProfileImage(ctx context.Context, userID int64, file io.Reader, contentType, filename string) (string, error)
	UpdateProfile(ctx context.Context, userID int64, name, email, phone, address string, lat, lng float64, photo string) error
	GetCustomers(ctx context.Context, search string, page, limit int, orderBy string) ([]entity.UserEntity, *entity.PaginationEntity, error)
	GetCustomerByID(ctx context.Context, customerID int64) (*entity.UserEntity, error)
}

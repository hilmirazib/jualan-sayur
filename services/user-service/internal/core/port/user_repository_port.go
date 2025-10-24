package port

import (
	"context"
	"user-service/internal/core/domain/entity"
)

type UserRepositoryInterface interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error)
	CreateUser(ctx context.Context, user *entity.UserEntity) (*entity.UserEntity, error)
	GetRoleByName(ctx context.Context, name string) (*entity.RoleEntity, error)
	UpdateUserVerificationStatus(ctx context.Context, userID int64, isVerified bool) error
	GetUserByEmailIncludingUnverified(ctx context.Context, email string) (*entity.UserEntity, error)
	UpdateUserPassword(ctx context.Context, userID int64, hashedPassword string) error
	GetUserByID(ctx context.Context, userID int64) (*entity.UserEntity, error)
	UpdateUserPhoto(ctx context.Context, userID int64, photoURL string) error
	UpdateUserEmail(ctx context.Context, userID int64, email string) error
	UpdateUserProfile(ctx context.Context, userID int64, name, email, phone, address string, lat, lng float64, photo string) error
	GetCustomers(ctx context.Context, search string, page, limit int, orderBy string) ([]entity.UserEntity, int64, error)
	GetCustomerByID(ctx context.Context, customerID int64) (*entity.UserEntity, error)
}

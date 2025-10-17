package service

import (
	"context"
	"errors"
	"user-service/config"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/port"
)

var (
	ErrInvalidEmail = errors.New("invalid email format")
	ErrUserNotFound = errors.New("user not found")
)

type UserService struct {
	AuthServiceInterface
	config *config.Config
}

func (u *UserService) GetProfile(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	return u.AuthServiceInterface.GetProfile(ctx, userID)
}

func NewUserService(userRepo port.UserRepositoryInterface, sessionRepo port.SessionInterface, jwtUtil port.JWTInterface, verificationTokenRepo port.VerificationTokenInterface, emailPublisher port.EmailInterface, blacklistTokenRepo port.BlacklistTokenInterface, storage port.StorageInterface, cfg *config.Config) port.UserServiceInterface {
	return &UserService{
		AuthServiceInterface: NewAuthService(userRepo, sessionRepo, jwtUtil, verificationTokenRepo, emailPublisher, blacklistTokenRepo, storage),
		config:               cfg,
	}
}

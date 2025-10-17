package service

import (
	"errors"
	"user-service/config"
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

func NewUserService(userRepo port.UserRepositoryInterface, sessionRepo port.SessionInterface, jwtUtil port.JWTInterface, verificationTokenRepo port.VerificationTokenInterface, emailPublisher port.EmailInterface, blacklistTokenRepo port.BlacklistTokenInterface, cfg *config.Config) port.UserServiceInterface {
	return &UserService{
		AuthServiceInterface: NewAuthService(userRepo, sessionRepo, jwtUtil, verificationTokenRepo, emailPublisher, blacklistTokenRepo),
		config:               cfg,
	}
}

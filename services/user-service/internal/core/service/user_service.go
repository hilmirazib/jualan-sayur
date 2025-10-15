package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"user-service/config"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/port"
	"user-service/utils"

	"github.com/rs/zerolog/log"
)

var (
	ErrInvalidEmail = errors.New("invalid email format")
	ErrUserNotFound = errors.New("user not found")
)

type UserService struct {
	userRepo    port.UserRepositoryInterface
	sessionRepo port.SessionInterface
	jwtUtil     port.JWTInterface
	config      *config.Config
}

func NewUserService(userRepo port.UserRepositoryInterface, sessionRepo port.SessionInterface, jwtUtil port.JWTInterface, cfg *config.Config) port.UserServiceInterface {
	return &UserService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtUtil:     jwtUtil,
		config:      cfg,
	}
}

// Signin implements port.UserServiceInterface.
func (s *UserService) SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error) {
	// Input validation
	if err := s.validateEmail(req.Email); err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("[UserService-SignIn] Invalid email format")
		return nil, "", err
	}

	// Business logic: normalize email to lowercase
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("[UserService-SignIn] Failed to get user from repository")
		if err.Error() == "record not found" {
			return nil, "", errors.New("user not found")
		}
		return nil, "", err
	}

	if checkPass := utils.CheckPasswordHash(req.Password, user.Password); !checkPass {
		log.Warn().Str("email", req.Email).Msg("[UserService-SignIn] Incorrect password")
		return nil, "", errors.New("incorrect password")
	}

	// Generate session ID
	sessionID := "sess_" + fmt.Sprintf("%d", time.Now().UnixNano())

	// Generate JWT token with session ID
	token, err := s.jwtUtil.GenerateJWTWithSession(user.ID, user.Email, user.RoleName, sessionID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", user.ID).Msg("[UserService-SignIn] Failed to generate JWT token")
		return nil, "", errors.New("failed to generate token")
	}

	// Store token in Redis session
	err = s.sessionRepo.StoreToken(ctx, user.ID, sessionID, token)
	if err != nil {
		log.Error().Err(err).Int64("user_id", user.ID).Msg("[UserService-SignIn] Failed to store token in session")
		return nil, "", errors.New("failed to create session")
	}

	log.Info().Int64("user_id", user.ID).Str("email", req.Email).Str("session_id", sessionID).Msg("[UserService-SignIn] User signed in successfully")
	return user, token, nil
}

// validateEmail performs basic email validation
func (s *UserService) validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}

	email = strings.TrimSpace(email)
	if len(email) == 0 {
		return ErrInvalidEmail
	}

	// Basic email format check
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return ErrInvalidEmail
	}

	return nil
}

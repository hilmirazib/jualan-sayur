package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	userRepo             port.UserRepositoryInterface
	sessionRepo          port.SessionInterface
	jwtUtil              port.JWTInterface
	verificationTokenRepo port.VerificationTokenInterface
	emailPublisher       port.EmailInterface
	config               *config.Config
}

func NewUserService(userRepo port.UserRepositoryInterface, sessionRepo port.SessionInterface, jwtUtil port.JWTInterface, verificationTokenRepo port.VerificationTokenInterface, emailPublisher port.EmailInterface, cfg *config.Config) port.UserServiceInterface {
	return &UserService{
		userRepo:             userRepo,
		sessionRepo:          sessionRepo,
		jwtUtil:              jwtUtil,
		verificationTokenRepo: verificationTokenRepo,
		emailPublisher:       emailPublisher,
		config:               cfg,
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

// CreateUserAccount implements port.UserServiceInterface.
func (s *UserService) CreateUserAccount(ctx context.Context, email, name, password, passwordConfirmation string) error {
	// Input validation
	if err := s.validateEmail(email); err != nil {
		log.Error().Err(err).Str("email", email).Msg("[UserService-CreateUserAccount] Invalid email format")
		return err
	}

	if err := s.validatePassword(password, passwordConfirmation); err != nil {
		log.Error().Err(err).Str("email", email).Msg("[UserService-CreateUserAccount] Password validation failed")
		return err
	}

	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)

	// Check if email already exists (including unverified users)
	existingUser, err := s.userRepo.GetUserByEmailIncludingUnverified(ctx, email)
	if err == nil && existingUser != nil {
		log.Warn().Str("email", email).Msg("[UserService-CreateUserAccount] Email already exists")
		return errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("[UserService-CreateUserAccount] Failed to hash password")
		return errors.New("failed to process password")
	}

	// Create user entity
	userEntity := &entity.UserEntity{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		IsVerified: false, // User is not verified initially
	}

	// Create user in database
	createdUser, err := s.userRepo.CreateUser(ctx, userEntity)
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("[UserService-CreateUserAccount] Failed to create user")
		return errors.New("failed to create account")
	}

	// Generate verification token
	token, err := s.generateVerificationToken()
	if err != nil {
		log.Error().Err(err).Int64("user_id", createdUser.ID).Msg("[UserService-CreateUserAccount] Failed to generate verification token")
		return errors.New("failed to generate verification token")
	}

	// Create verification token entity
	verificationToken := &entity.VerificationTokenEntity{
		UserID:    createdUser.ID,
		Token:     token,
		TokenType: "email_verification",
		ExpiresAt: time.Now().Add(24 * time.Hour), // Token expires in 24 hours
	}

	// Save verification token
	err = s.verificationTokenRepo.CreateVerificationToken(ctx, verificationToken)
	if err != nil {
		log.Error().Err(err).Int64("user_id", createdUser.ID).Msg("[UserService-CreateUserAccount] Failed to save verification token")
		return errors.New("failed to create verification token")
	}

	// Send verification email
	err = s.emailPublisher.SendVerificationEmail(ctx, email, token)
	if err != nil {
		log.Error().Err(err).Int64("user_id", createdUser.ID).Str("email", email).Msg("[UserService-CreateUserAccount] Failed to send verification email")
		// Don't return error here, account is created but email failed
		log.Warn().Int64("user_id", createdUser.ID).Msg("[UserService-CreateUserAccount] Account created but email sending failed")
	}

	log.Info().Int64("user_id", createdUser.ID).Str("email", email).Msg("[UserService-CreateUserAccount] User account created successfully")
	return nil
}

// VerifyUserAccount implements port.UserServiceInterface.
func (s *UserService) VerifyUserAccount(ctx context.Context, token string) error {
	// Get verification token
	verificationToken, err := s.verificationTokenRepo.GetVerificationToken(ctx, token)
	if err != nil {
		if err.Error() == "record not found" {
			log.Warn().Str("token", token).Msg("[UserService-VerifyUserAccount] Verification token not found or expired")
			return errors.New("invalid or expired verification token")
		}
		log.Error().Err(err).Str("token", token).Msg("[UserService-VerifyUserAccount] Failed to get verification token")
		return errors.New("failed to verify token")
	}

	// Update user verification status
	err = s.userRepo.UpdateUserVerificationStatus(ctx, verificationToken.UserID, true)
	if err != nil {
		log.Error().Err(err).Int64("user_id", verificationToken.UserID).Msg("[UserService-VerifyUserAccount] Failed to update user verification status")
		return errors.New("failed to verify account")
	}

	// Delete used verification token (one-time use)
	err = s.verificationTokenRepo.DeleteVerificationToken(ctx, token)
	if err != nil {
		log.Error().Err(err).Str("token", token).Msg("[UserService-VerifyUserAccount] Failed to delete verification token")
		// Don't return error here, account is already verified
	}

	log.Info().Int64("user_id", verificationToken.UserID).Str("token", token).Msg("[UserService-VerifyUserAccount] User account verified successfully")
	return nil
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

// validatePassword validates password and confirmation
func (s *UserService) validatePassword(password, confirmation string) error {
	if password == "" {
		return errors.New("password is required")
	}

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if password != confirmation {
		return errors.New("password confirmation does not match")
	}

	return nil
}

// generateVerificationToken generates a secure random token
func (s *UserService) generateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

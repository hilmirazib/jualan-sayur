package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/port"
	"user-service/utils"

	"github.com/rs/zerolog/log"
)

type AuthServiceInterface interface {
	SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error)
	CreateUserAccount(ctx context.Context, email, name, password, passwordConfirmation string) error
	VerifyUserAccount(ctx context.Context, token string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword, passwordConfirmation string) error
}

type AuthService struct {
	userRepo             port.UserRepositoryInterface
	sessionRepo          port.SessionInterface
	jwtUtil              port.JWTInterface
	verificationTokenRepo port.VerificationTokenInterface
	emailPublisher       port.EmailInterface
}

func NewAuthService(userRepo port.UserRepositoryInterface, sessionRepo port.SessionInterface, jwtUtil port.JWTInterface, verificationTokenRepo port.VerificationTokenInterface, emailPublisher port.EmailInterface) AuthServiceInterface {
	return &AuthService{
		userRepo:             userRepo,
		sessionRepo:          sessionRepo,
		jwtUtil:              jwtUtil,
		verificationTokenRepo: verificationTokenRepo,
		emailPublisher:       emailPublisher,
	}
}

func (s *AuthService) SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error) {
	if err := s.validateEmail(req.Email); err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("[AuthService-SignIn] Invalid email format")
		return nil, "", err
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("[AuthService-SignIn] Failed to get user from repository")
		if err.Error() == "record not found" {
			return nil, "", errors.New("user not found")
		}
		return nil, "", err
	}

	if checkPass := utils.CheckPasswordHash(req.Password, user.Password); !checkPass {
		log.Warn().Str("email", req.Email).Msg("[AuthService-SignIn] Incorrect password")
		return nil, "", errors.New("incorrect password")
	}

	sessionID := "sess_" + fmt.Sprintf("%d", time.Now().UnixNano())

	token, err := s.jwtUtil.GenerateJWTWithSession(user.ID, user.Email, user.RoleName, sessionID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", user.ID).Msg("[AuthService-SignIn] Failed to generate JWT token")
		return nil, "", errors.New("failed to generate token")
	}

	err = s.sessionRepo.StoreToken(ctx, user.ID, sessionID, token)
	if err != nil {
		log.Error().Err(err).Int64("user_id", user.ID).Msg("[AuthService-SignIn] Failed to store token in session")
		return nil, "", errors.New("failed to create session")
	}

	log.Info().Int64("user_id", user.ID).Str("email", req.Email).Str("session_id", sessionID).Msg("[AuthService-SignIn] User signed in successfully")
	return user, token, nil
}

func (s *AuthService) CreateUserAccount(ctx context.Context, email, name, password, passwordConfirmation string) error {
	if err := s.validateEmail(email); err != nil {
		log.Error().Err(err).Str("email", email).Msg("[AuthService-CreateUserAccount] Invalid email format")
		return err
	}

	if err := s.validatePassword(password, passwordConfirmation); err != nil {
		log.Error().Err(err).Str("email", email).Msg("[AuthService-CreateUserAccount] Password validation failed")
		return err
	}

	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)

	existingUser, err := s.userRepo.GetUserByEmailIncludingUnverified(ctx, email)
	if err == nil && existingUser != nil {
		log.Warn().Str("email", email).Msg("[AuthService-CreateUserAccount] Email already exists")
		return errors.New("email already exists")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("[AuthService-CreateUserAccount] Failed to hash password")
		return errors.New("failed to process password")
	}

	userEntity := &entity.UserEntity{
		Name:       name,
		Email:      email,
		Password:   hashedPassword,
		IsVerified: false,
	}

	createdUser, err := s.userRepo.CreateUser(ctx, userEntity)
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("[AuthService-CreateUserAccount] Failed to create user")
		return errors.New("failed to create account")
	}

	token, err := s.generateVerificationToken()
	if err != nil {
		log.Error().Err(err).Int64("user_id", createdUser.ID).Msg("[AuthService-CreateUserAccount] Failed to generate verification token")
		return errors.New("failed to generate verification token")
	}

	verificationToken := &entity.VerificationTokenEntity{
		UserID:    createdUser.ID,
		Token:     token,
		TokenType: "email_verification",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err = s.verificationTokenRepo.CreateVerificationToken(ctx, verificationToken)
	if err != nil {
		log.Error().Err(err).Int64("user_id", createdUser.ID).Msg("[AuthService-CreateUserAccount] Failed to save verification token")
		return errors.New("failed to create verification token")
	}

	err = s.emailPublisher.SendVerificationEmail(ctx, email, token)
	if err != nil {
		log.Error().Err(err).Int64("user_id", createdUser.ID).Str("email", email).Msg("[AuthService-CreateUserAccount] Failed to send verification email")
		log.Warn().Int64("user_id", createdUser.ID).Msg("[AuthService-CreateUserAccount] Account created but email sending failed")
	}

	log.Info().Int64("user_id", createdUser.ID).Str("email", email).Msg("[AuthService-CreateUserAccount] User account created successfully")
	return nil
}

func (s *AuthService) VerifyUserAccount(ctx context.Context, token string) error {
	verificationToken, err := s.verificationTokenRepo.GetVerificationToken(ctx, token)
	if err != nil {
		if err.Error() == "record not found" {
			log.Warn().Str("token", token).Msg("[AuthService-VerifyUserAccount] Verification token not found or expired")
			return errors.New("invalid or expired verification token")
		}
		log.Error().Err(err).Str("token", token).Msg("[AuthService-VerifyUserAccount] Failed to get verification token")
		return errors.New("failed to verify token")
	}

	err = s.userRepo.UpdateUserVerificationStatus(ctx, verificationToken.UserID, true)
	if err != nil {
		log.Error().Err(err).Int64("user_id", verificationToken.UserID).Msg("[AuthService-VerifyUserAccount] Failed to update user verification status")
		return errors.New("failed to verify account")
	}

	err = s.verificationTokenRepo.DeleteVerificationToken(ctx, token)
	if err != nil {
		log.Error().Err(err).Str("token", token).Msg("[AuthService-VerifyUserAccount] Failed to delete verification token")
	}

	log.Info().Int64("user_id", verificationToken.UserID).Str("token", token).Msg("[AuthService-VerifyUserAccount] User account verified successfully")
	return nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	if err := s.validateEmail(email); err != nil {
		log.Error().Err(err).Str("email", email).Msg("[AuthService-ForgotPassword] Invalid email format")
		return err
	}

	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if err.Error() == "record not found" {
			log.Warn().Str("email", email).Msg("[AuthService-ForgotPassword] User not found")
			return nil
		}
		log.Error().Err(err).Str("email", email).Msg("[AuthService-ForgotPassword] Failed to get user from repository")
		return errors.New("failed to process request")
	}

	if !user.IsVerified {
		log.Warn().Str("email", email).Msg("[AuthService-ForgotPassword] User account not verified")
		return nil
	}

	token, err := s.generateVerificationToken()
	if err != nil {
		log.Error().Err(err).Int64("user_id", user.ID).Msg("[AuthService-ForgotPassword] Failed to generate reset token")
		return errors.New("failed to generate reset token")
	}

	resetToken := &entity.VerificationTokenEntity{
		UserID:    user.ID,
		Token:     token,
		TokenType: "password_reset",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	err = s.verificationTokenRepo.CreateVerificationToken(ctx, resetToken)
	if err != nil {
		log.Error().Err(err).Int64("user_id", user.ID).Msg("[AuthService-ForgotPassword] Failed to save reset token")
		return errors.New("failed to create reset token")
	}

	err = s.emailPublisher.SendPasswordResetEmail(ctx, email, token)
	if err != nil {
		log.Error().Err(err).Int64("user_id", user.ID).Str("email", email).Msg("[AuthService-ForgotPassword] Failed to send password reset email")
		log.Warn().Int64("user_id", user.ID).Msg("[AuthService-ForgotPassword] Reset token created but email sending failed")
	}

	log.Info().Int64("user_id", user.ID).Str("email", email).Msg("[AuthService-ForgotPassword] Password reset request processed successfully")
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword, passwordConfirmation string) error {
	if err := s.validatePassword(newPassword, passwordConfirmation); err != nil {
		log.Error().Err(err).Msg("[AuthService-ResetPassword] Password validation failed")
		return err
	}

	resetToken, err := s.verificationTokenRepo.GetVerificationToken(ctx, token)
	if err != nil {
		if err.Error() == "record not found" {
			log.Warn().Str("token", token).Msg("[AuthService-ResetPassword] Reset token not found or expired")
			return errors.New("invalid or expired reset token")
		}
		log.Error().Err(err).Str("token", token).Msg("[AuthService-ResetPassword] Failed to get reset token")
		return errors.New("failed to validate token")
	}

	if resetToken.TokenType != "password_reset" {
		log.Warn().Str("token", token).Str("token_type", resetToken.TokenType).Msg("[AuthService-ResetPassword] Token is not for password reset")
		return errors.New("invalid token type")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		log.Error().Err(err).Int64("user_id", resetToken.UserID).Msg("[AuthService-ResetPassword] Failed to hash new password")
		return errors.New("failed to process password")
	}

	err = s.userRepo.UpdateUserPassword(ctx, resetToken.UserID, hashedPassword)
	if err != nil {
		log.Error().Err(err).Int64("user_id", resetToken.UserID).Msg("[AuthService-ResetPassword] Failed to update user password")
		return errors.New("failed to update password")
	}

	err = s.verificationTokenRepo.DeleteVerificationToken(ctx, token)
	if err != nil {
		log.Error().Err(err).Str("token", token).Msg("[AuthService-ResetPassword] Failed to delete reset token")
	}

	log.Info().Int64("user_id", resetToken.UserID).Str("token", token).Msg("[AuthService-ResetPassword] Password reset successfully")
	return nil
}

func (s *AuthService) validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}

	email = strings.TrimSpace(email)
	if len(email) == 0 {
		return ErrInvalidEmail
	}

	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return ErrInvalidEmail
	}

	return nil
}

func (s *AuthService) validatePassword(password, confirmation string) error {
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

func (s *AuthService) generateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

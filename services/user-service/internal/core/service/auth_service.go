package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
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
	Logout(ctx context.Context, userID int64, sessionID, tokenString string, tokenExpiresAt int64) error
	GetProfile(ctx context.Context, userID int64) (*entity.UserEntity, error)
	UploadProfileImage(ctx context.Context, userID int64, file io.Reader, contentType, filename string) (string, error)
	UpdateProfile(ctx context.Context, userID int64, name, email, phone, address string, lat, lng float64, photo string) error
}

type AuthService struct {
	userRepo              port.UserRepositoryInterface
	sessionRepo           port.SessionInterface
	jwtUtil               port.JWTInterface
	verificationTokenRepo port.VerificationTokenInterface
	emailPublisher        port.EmailInterface
	blacklistTokenRepo    port.BlacklistTokenInterface
	storage               port.StorageInterface
}

func NewAuthService(userRepo port.UserRepositoryInterface, sessionRepo port.SessionInterface, jwtUtil port.JWTInterface, verificationTokenRepo port.VerificationTokenInterface, emailPublisher port.EmailInterface, blacklistTokenRepo port.BlacklistTokenInterface, storage port.StorageInterface) AuthServiceInterface {
	return &AuthService{
		userRepo:              userRepo,
		sessionRepo:           sessionRepo,
		jwtUtil:               jwtUtil,
		verificationTokenRepo: verificationTokenRepo,
		emailPublisher:        emailPublisher,
		blacklistTokenRepo:    blacklistTokenRepo,
		storage:               storage,
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

func (s *AuthService) Logout(ctx context.Context, userID int64, sessionID, tokenString string, tokenExpiresAt int64) error {
	// Delete session from Redis (primary logout mechanism)
	err := s.sessionRepo.DeleteToken(ctx, userID, sessionID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Str("session_id", sessionID).Msg("[AuthService-Logout] Failed to delete session token")
		return errors.New("failed to logout")
	}

	// Add token to blacklist for maximum security (prevent reuse if token stolen)
	if tokenString != "" && tokenExpiresAt > 0 {
		hash := sha256.Sum256([]byte(tokenString))
		tokenHash := hex.EncodeToString(hash[:])

		err = s.blacklistTokenRepo.AddToBlacklist(ctx, tokenHash, tokenExpiresAt)
		if err != nil {
			log.Error().Err(err).Int64("user_id", userID).Str("session_id", sessionID).Msg("[AuthService-Logout] Failed to add token to blacklist")
			// Don't fail logout if blacklist fails, just log the error
		} else {
			log.Info().Int64("user_id", userID).Str("session_id", sessionID).Msg("[AuthService-Logout] Token added to blacklist successfully")
		}
	}

	log.Info().Int64("user_id", userID).Str("session_id", sessionID).Msg("[AuthService-Logout] User logged out successfully")
	return nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("[AuthService-GetProfile] Failed to get user profile")
		if err.Error() == "record not found" {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	log.Info().Int64("user_id", userID).Msg("[AuthService-GetProfile] User profile retrieved successfully")
	return user, nil
}

func (s *AuthService) UploadProfileImage(ctx context.Context, userID int64, file io.Reader, contentType, filename string) (string, error) {
	log.Info().Int64("user_id", userID).Str("content_type", contentType).Str("filename", filename).Msg("[AuthService-UploadProfileImage] Starting image upload")

	// Get current user to check for existing photo
	currentUser, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("[AuthService-UploadProfileImage] Failed to get current user data")
		return "", errors.New("failed to get user data")
	}

	// Upload file to storage
	imageURL, err := s.storage.UploadFile(ctx, "", "", file, contentType)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("[AuthService-UploadProfileImage] Failed to upload image to storage")
		return "", errors.New("failed to upload image")
	}

	// Update user photo URL in database
	err = s.userRepo.UpdateUserPhoto(ctx, userID, imageURL)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Str("image_url", imageURL).Msg("[AuthService-UploadProfileImage] Failed to update user photo in database")
		// Try to delete uploaded file if database update fails
		newObjectName := s.extractObjectNameFromURL(imageURL)
		if newObjectName != "" {
			if deleteErr := s.storage.DeleteFile(ctx, "", newObjectName); deleteErr != nil {
				log.Error().Err(deleteErr).Str("image_url", imageURL).Msg("[AuthService-UploadProfileImage] Failed to delete uploaded file after database error")
			}
		}
		return "", errors.New("failed to update profile")
	}

	// Delete old photo from storage if it exists
	if currentUser.Photo != "" && currentUser.Photo != imageURL {
		oldObjectName := s.extractObjectNameFromURL(currentUser.Photo)
		if oldObjectName != "" {
			if deleteErr := s.storage.DeleteFile(ctx, "", oldObjectName); deleteErr != nil {
				log.Warn().Err(deleteErr).Str("old_photo_url", currentUser.Photo).Msg("[AuthService-UploadProfileImage] Failed to delete old photo from storage")
				// Don't fail the upload if old photo deletion fails
			} else {
				log.Info().Int64("user_id", userID).Str("old_photo_url", currentUser.Photo).Msg("[AuthService-UploadProfileImage] Old photo deleted successfully")
			}
		}
	}

	log.Info().Int64("user_id", userID).Str("image_url", imageURL).Msg("[AuthService-UploadProfileImage] Profile image uploaded successfully")
	return imageURL, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID int64, name, email, phone, address string, lat, lng float64, photo string) error {
	// Validate email format
	if err := s.validateEmail(email); err != nil {
		log.Error().Err(err).Str("email", email).Msg("[AuthService-UpdateProfile] Invalid email format")
		return err
	}

	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)
	address = strings.TrimSpace(address)

	// Check if email is already used by another user
	existingUser, err := s.userRepo.GetUserByEmailIncludingUnverified(ctx, email)
	if err != nil && err.Error() != "record not found" {
		log.Error().Err(err).Str("email", email).Msg("[AuthService-UpdateProfile] Failed to check email uniqueness")
		return errors.New("failed to validate email")
	}

	// If email exists and it's not the current user, return error
	if existingUser != nil && existingUser.ID != userID {
		log.Warn().Str("email", email).Int64("existing_user_id", existingUser.ID).Int64("current_user_id", userID).Msg("[AuthService-UpdateProfile] Email already exists")
		return errors.New("email already exists")
	}

	// Update user profile
	err = s.userRepo.UpdateUserProfile(ctx, userID, name, email, phone, address, lat, lng, photo)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Str("email", email).Msg("[AuthService-UpdateProfile] Failed to update user profile")
		return errors.New("failed to update profile")
	}

	log.Info().Int64("user_id", userID).Str("email", email).Msg("[AuthService-UpdateProfile] User profile updated successfully")
	return nil
}

func (s *AuthService) generateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// URL format: https://project.supabase.co/storage/v1/object/public/bucket-name/object-name
func (s *AuthService) extractObjectNameFromURL(url string) string {
	// Find the position after "/storage/v1/object/public/"
	parts := strings.Split(url, "/storage/v1/object/public/")
	if len(parts) != 2 {
		return ""
	}

	// The second part contains "bucket-name/object-name"
	// extract everything after the first "/"
	bucketAndObject := parts[1]
	slashIndex := strings.Index(bucketAndObject, "/")
	if slashIndex == -1 || slashIndex == len(bucketAndObject)-1 {
		return ""
	}

	// Return the object name (everything after the first "/")
	return bucketAndObject[slashIndex+1:]
}

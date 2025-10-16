package service

import (
	"context"
	"errors"
	"testing"
	"user-service/config"
	"user-service/internal/core/domain/entity"
	"user-service/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_SignIn_UserNotFound(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, nil, nil, nil, nil, &config.Config{})

	ctx := context.Background()
	email := "notfound@example.com"

	mockRepo.On("GetUserByEmail", ctx, email).Return(nil, errors.New("record not found"))

	// Execute
	user, token, err := service.SignIn(ctx, entity.UserEntity{
		Email:    email,
		Password: "password123",
	})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Equal(t, "user not found", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestUserService_SignIn_InvalidEmail(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, nil, nil, nil, nil, &config.Config{})

	ctx := context.Background()

	// Execute with empty email
	user, token, err := service.SignIn(ctx, entity.UserEntity{
		Email:    "",
		Password: "password123",
	})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Equal(t, ErrInvalidEmail, err)

	// Mock should not be called
	mockRepo.AssertNotCalled(t, "GetUserByEmail", mock.Anything, mock.Anything)
}

func TestUserService_CreateUserAccount_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	mockEmailPublisher := new(MockEmailPublisher)
	service := NewUserService(mockUserRepo, nil, nil, mockVerificationTokenRepo, mockEmailPublisher, &config.Config{})

	ctx := context.Background()
	email := "test@example.com"
	name := "Test User"
	password := "password123"
	passwordConfirmation := "password123"

	// Mock expectations
	mockUserRepo.On("GetUserByEmailIncludingUnverified", ctx, email).Return(nil, errors.New("record not found"))
	mockUserRepo.On("CreateUser", ctx, mock.AnythingOfType("*entity.UserEntity")).Return(&entity.UserEntity{ID: 1, Email: email, Name: name}, nil)
	mockVerificationTokenRepo.On("CreateVerificationToken", ctx, mock.AnythingOfType("*entity.VerificationTokenEntity")).Return(nil)
	mockEmailPublisher.On("SendVerificationEmail", ctx, email, mock.AnythingOfType("string")).Return(nil)

	// Execute
	err := service.CreateUserAccount(ctx, email, name, password, passwordConfirmation)

	// Assert
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockVerificationTokenRepo.AssertExpectations(t)
	mockEmailPublisher.AssertExpectations(t)
}

func TestUserService_CreateUserAccount_EmailAlreadyExists(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	service := NewUserService(mockUserRepo, nil, nil, nil, nil, &config.Config{})

	ctx := context.Background()
	email := "existing@example.com"

	existingUser := &entity.UserEntity{ID: 1, Email: email}
	mockUserRepo.On("GetUserByEmailIncludingUnverified", ctx, email).Return(existingUser, nil)

	// Execute
	err := service.CreateUserAccount(ctx, email, "Test User", "password123", "password123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "email already exists", err.Error())
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_VerifyUserAccount_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	service := NewUserService(mockUserRepo, nil, nil, mockVerificationTokenRepo, nil, &config.Config{})

	ctx := context.Background()
	token := "valid-token"

	// Mock expectations
	verificationToken := &entity.VerificationTokenEntity{UserID: 1, Token: token}
	mockVerificationTokenRepo.On("GetVerificationToken", ctx, token).Return(verificationToken, nil)
	mockUserRepo.On("UpdateUserVerificationStatus", ctx, int64(1), true).Return(nil)
	mockVerificationTokenRepo.On("DeleteVerificationToken", ctx, token).Return(nil)

	// Execute
	err := service.VerifyUserAccount(ctx, token)

	// Assert
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockVerificationTokenRepo.AssertExpectations(t)
}

func TestUserService_AdminCheck_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockJWTUtil := new(MockJWTUtil)
	mockConfig := &config.Config{
		App: config.App{
			JwtSecretKey: "test-secret-key",
			JwtIssuer:    "test-issuer",
		},
	}
	service := NewUserService(mockUserRepo, mockSessionRepo, mockJWTUtil, nil, nil, mockConfig)

	ctx := context.Background()
	email := "admin@example.com"
	password := "adminpass123"

	hashedPassword, _ := utils.HashPassword(password)
	adminUser := &entity.UserEntity{
		ID:       1,
		Email:    email,
		Password: hashedPassword,
		RoleName: "admin",
		Name:     "Admin User",
	}

	// Mock expectations
	mockUserRepo.On("GetUserByEmail", ctx, email).Return(adminUser, nil)
	mockJWTUtil.On("GenerateJWTWithSession", int64(1), email, "admin", mock.AnythingOfType("string")).Return("admin-jwt-token", nil)
	mockSessionRepo.On("StoreToken", ctx, int64(1), mock.AnythingOfType("string"), "admin-jwt-token").Return(nil)

	// Execute
	user, token, err := service.SignIn(ctx, entity.UserEntity{
		Email:    email,
		Password: password,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
	assert.Equal(t, "admin", user.RoleName)
	assert.Equal(t, "admin-jwt-token", token)
	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
	mockJWTUtil.AssertExpectations(t)
}

func TestUserService_ForgotPassword_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	mockEmailPublisher := new(MockEmailPublisher)
	service := NewUserService(mockUserRepo, nil, nil, mockVerificationTokenRepo, mockEmailPublisher, &config.Config{})

	ctx := context.Background()
	email := "user@example.com"

	existingUser := &entity.UserEntity{
		ID:         1,
		Email:      email,
		IsVerified: true,
	}

	// Mock expectations
	mockUserRepo.On("GetUserByEmail", ctx, email).Return(existingUser, nil)
	mockVerificationTokenRepo.On("CreateVerificationToken", ctx, mock.AnythingOfType("*entity.VerificationTokenEntity")).Return(nil)
	mockEmailPublisher.On("SendPasswordResetEmail", ctx, email, mock.AnythingOfType("string")).Return(nil)

	// Execute
	err := service.ForgotPassword(ctx, email)

	// Assert
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockVerificationTokenRepo.AssertExpectations(t)
	mockEmailPublisher.AssertExpectations(t)
}

func TestUserService_ForgotPassword_InvalidEmail(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	service := NewUserService(mockUserRepo, nil, nil, nil, nil, &config.Config{})

	ctx := context.Background()

	// Execute with empty email
	err := service.ForgotPassword(ctx, "")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidEmail, err)

	// Mock should not be called
	mockUserRepo.AssertNotCalled(t, "GetUserByEmail", mock.Anything, mock.Anything)
}

func TestUserService_ForgotPassword_UserNotFound(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	service := NewUserService(mockUserRepo, nil, nil, nil, nil, &config.Config{})

	ctx := context.Background()
	email := "notfound@example.com"

	// Mock expectations - return nil for security (don't reveal if user exists)
	mockUserRepo.On("GetUserByEmail", ctx, email).Return(nil, errors.New("record not found"))

	// Execute
	err := service.ForgotPassword(ctx, email)

	// Assert - should not error for security reasons
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_ForgotPassword_UserNotVerified(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	service := NewUserService(mockUserRepo, nil, nil, nil, nil, &config.Config{})

	ctx := context.Background()
	email := "unverified@example.com"

	unverifiedUser := &entity.UserEntity{
		ID:         1,
		Email:      email,
		IsVerified: false,
	}

	// Mock expectations - return nil for security (don't reveal verification status)
	mockUserRepo.On("GetUserByEmail", ctx, email).Return(unverifiedUser, nil)

	// Execute
	err := service.ForgotPassword(ctx, email)

	// Assert - should not error for security reasons
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_ResetPassword_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	service := NewUserService(mockUserRepo, nil, nil, mockVerificationTokenRepo, nil, &config.Config{})

	ctx := context.Background()
	token := "valid-reset-token"
	newPassword := "newpassword123"
	passwordConfirmation := "newpassword123"

	resetToken := &entity.VerificationTokenEntity{
		UserID:    1,
		Token:     token,
		TokenType: "password_reset",
	}

	// Mock expectations
	mockVerificationTokenRepo.On("GetVerificationToken", ctx, token).Return(resetToken, nil)
	mockUserRepo.On("UpdateUserPassword", ctx, int64(1), mock.AnythingOfType("string")).Return(nil)
	mockVerificationTokenRepo.On("DeleteVerificationToken", ctx, token).Return(nil)

	// Execute
	err := service.ResetPassword(ctx, token, newPassword, passwordConfirmation)

	// Assert
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockVerificationTokenRepo.AssertExpectations(t)
}

func TestUserService_ResetPassword_InvalidToken(t *testing.T) {
	// Setup
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	service := NewUserService(nil, nil, nil, mockVerificationTokenRepo, nil, &config.Config{})

	ctx := context.Background()
	token := "invalid-token"

	// Mock expectations
	mockVerificationTokenRepo.On("GetVerificationToken", ctx, token).Return(nil, errors.New("record not found"))

	// Execute
	err := service.ResetPassword(ctx, token, "newpass123", "newpass123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "invalid or expired reset token", err.Error())
	mockVerificationTokenRepo.AssertExpectations(t)
}

func TestUserService_ResetPassword_PasswordMismatch(t *testing.T) {
	// Setup
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	service := NewUserService(nil, nil, nil, mockVerificationTokenRepo, nil, &config.Config{})

	ctx := context.Background()
	token := "valid-token"
	newPassword := "newpass123"
	passwordConfirmation := "differentpass123"

	// Execute - password validation happens before token validation
	err := service.ResetPassword(ctx, token, newPassword, passwordConfirmation)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "password confirmation does not match", err.Error())

	// Mock should not be called because password validation fails first
	mockVerificationTokenRepo.AssertNotCalled(t, "GetVerificationToken", mock.Anything, mock.Anything)
}

func TestUserService_ResetPassword_PasswordTooShort(t *testing.T) {
	// Setup
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	service := NewUserService(nil, nil, nil, mockVerificationTokenRepo, nil, &config.Config{})

	ctx := context.Background()
	token := "valid-token"
	newPassword := "short"
	passwordConfirmation := "short"

	// Execute - password validation happens before token validation
	err := service.ResetPassword(ctx, token, newPassword, passwordConfirmation)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "password must be at least 8 characters long", err.Error())

	// Mock should not be called because password validation fails first
	mockVerificationTokenRepo.AssertNotCalled(t, "GetVerificationToken", mock.Anything, mock.Anything)
}

func TestUserService_ResetPassword_WrongTokenType(t *testing.T) {
	// Setup
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	service := NewUserService(nil, nil, nil, mockVerificationTokenRepo, nil, &config.Config{})

	ctx := context.Background()
	token := "email-verification-token"
	newPassword := "newpassword123"
	passwordConfirmation := "newpassword123"

	wrongTypeToken := &entity.VerificationTokenEntity{
		UserID:    1,
		Token:     token,
		TokenType: "email_verification", // Wrong type
	}

	// Mock expectations
	mockVerificationTokenRepo.On("GetVerificationToken", ctx, token).Return(wrongTypeToken, nil)

	// Execute
	err := service.ResetPassword(ctx, token, newPassword, passwordConfirmation)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "invalid token type", err.Error())
	mockVerificationTokenRepo.AssertExpectations(t)
}

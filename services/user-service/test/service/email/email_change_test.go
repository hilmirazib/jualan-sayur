package main

import (
	"context"
	"errors"
	"testing"
	"time"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"
	"user-service/test/service/mocks"
	"user-service/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_VerifyEmailChange_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	mockVerificationTokenRepo := new(mocks.MockVerificationTokenRepository)
	service := service.NewAuthService(mockUserRepo, nil, nil, mockVerificationTokenRepo, nil, nil, nil)

	ctx := context.Background()
	token := "valid-email-change-token"
	userID := int64(1)
	newEmail := "newemail@example.com"

	// Mock verification token
	verificationToken := &entity.VerificationTokenEntity{
		UserID:    userID,
		Token:     token,
		TokenType: "email_change",
		NewEmail:  newEmail,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Mock expectations
	mockVerificationTokenRepo.On("GetVerificationToken", ctx, token).Return(verificationToken, nil)
	mockUserRepo.On("UpdateUserEmail", ctx, userID, newEmail).Return(nil)
	mockUserRepo.On("UpdateUserVerificationStatus", ctx, userID, true).Return(nil)
	mockVerificationTokenRepo.On("DeleteVerificationToken", ctx, token).Return(nil)

	// Execute
	err := service.VerifyEmailChange(ctx, token)

	// Assert
	assert.NoError(t, err)
	mockVerificationTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_VerifyEmailChange_InvalidToken(t *testing.T) {
	// Setup
	mockVerificationTokenRepo := new(mocks.MockVerificationTokenRepository)
	service := service.NewAuthService(nil, nil, nil, mockVerificationTokenRepo, nil, nil, nil)

	ctx := context.Background()
	token := "invalid-token"

	// Mock expectations
	mockVerificationTokenRepo.On("GetVerificationToken", ctx, token).Return(nil, errors.New("record not found"))

	// Execute
	err := service.VerifyEmailChange(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "invalid or expired verification token", err.Error())
	mockVerificationTokenRepo.AssertExpectations(t)
}

func TestAuthService_VerifyEmailChange_WrongTokenType(t *testing.T) {
	// Setup
	mockVerificationTokenRepo := new(mocks.MockVerificationTokenRepository)
	service := service.NewAuthService(nil, nil, nil, mockVerificationTokenRepo, nil, nil, nil)

	ctx := context.Background()
	token := "wrong-type-token"

	// Mock verification token with wrong type
	verificationToken := &entity.VerificationTokenEntity{
		UserID:    1,
		Token:     token,
		TokenType: "password_reset", // Wrong type
		NewEmail:  "newemail@example.com",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Mock expectations
	mockVerificationTokenRepo.On("GetVerificationToken", ctx, token).Return(verificationToken, nil)

	// Execute
	err := service.VerifyEmailChange(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "invalid token type", err.Error())
	mockVerificationTokenRepo.AssertExpectations(t)
}

func TestAuthService_VerifyEmailChange_MissingNewEmail(t *testing.T) {
	// Setup
	mockVerificationTokenRepo := new(mocks.MockVerificationTokenRepository)
	service := service.NewAuthService(nil, nil, nil, mockVerificationTokenRepo, nil, nil, nil)

	ctx := context.Background()
	token := "missing-email-token"

	// Mock verification token without new email
	verificationToken := &entity.VerificationTokenEntity{
		UserID:    1,
		Token:     token,
		TokenType: "email_change",
		NewEmail:  "", // Missing new email
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Mock expectations
	mockVerificationTokenRepo.On("GetVerificationToken", ctx, token).Return(verificationToken, nil)

	// Execute
	err := service.VerifyEmailChange(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "invalid token data", err.Error())
	mockVerificationTokenRepo.AssertExpectations(t)
}

func TestAuthService_VerifyEmailChange_UpdateEmailFailure(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	mockVerificationTokenRepo := new(mocks.MockVerificationTokenRepository)
	service := service.NewAuthService(mockUserRepo, nil, nil, mockVerificationTokenRepo, nil, nil, nil)

	ctx := context.Background()
	token := "update-failure-token"
	userID := int64(1)
	newEmail := "newemail@example.com"

	// Mock verification token
	verificationToken := &entity.VerificationTokenEntity{
		UserID:    userID,
		Token:     token,
		TokenType: "email_change",
		NewEmail:  newEmail,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Mock expectations
	mockVerificationTokenRepo.On("GetVerificationToken", ctx, token).Return(verificationToken, nil)
	mockUserRepo.On("UpdateUserEmail", ctx, userID, newEmail).Return(errors.New("database error"))

	// Execute
	err := service.VerifyEmailChange(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "failed to update email", err.Error())
	mockVerificationTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "UpdateUserVerificationStatus", mock.Anything, mock.Anything, mock.Anything)
	mockVerificationTokenRepo.AssertNotCalled(t, "DeleteVerificationToken", mock.Anything, mock.Anything)
}

func TestAuthService_CompleteEmailChangeFlow(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	mockSessionRepo := new(mocks.MockSessionRepository)
	mockJWTUtil := new(mocks.MockJWTUtil)
	mockVerificationTokenRepo := new(mocks.MockVerificationTokenRepository)
	mockEmailPublisher := new(mocks.MockEmailPublisher)
	service := service.NewAuthService(mockUserRepo, mockSessionRepo, mockJWTUtil, mockVerificationTokenRepo, mockEmailPublisher, nil, nil)

	ctx := context.Background()
	userID := int64(1)
	oldEmail := "old@example.com"
	newEmail := "new@example.com"
	name := "John Doe"
	phone := "081234567890"
	address := "Jakarta"
	lat := -6.2088
	lng := 106.8456
	photo := "https://example.com/photo.jpg"

	// Step 1: Update Profile (email change initiated)
	currentUser := &entity.UserEntity{
		ID:    userID,
		Email: oldEmail,
		Photo: photo,
	}

	// Mock expectations for UpdateProfile
	mockUserRepo.On("GetUserByID", ctx, userID).Return(currentUser, nil)
	mockUserRepo.On("GetUserByEmailIncludingUnverified", ctx, newEmail).Return(nil, errors.New("record not found"))
	mockVerificationTokenRepo.On("CreateVerificationToken", ctx, mock.AnythingOfType("*entity.VerificationTokenEntity")).Return(nil).Run(func(args mock.Arguments) {
		token := args.Get(1).(*entity.VerificationTokenEntity) // args[0] is ctx, args[1] is token
		assert.Equal(t, userID, token.UserID)
		assert.Equal(t, "email_change", token.TokenType)
		assert.Equal(t, newEmail, token.NewEmail)
	})
	mockEmailPublisher.On("SendEmailChangeVerificationEmail", ctx, newEmail, mock.AnythingOfType("string")).Return(nil)
	mockUserRepo.On("UpdateUserVerificationStatus", ctx, userID, false).Return(nil)
	mockUserRepo.On("UpdateUserProfile", ctx, userID, name, oldEmail, phone, address, lat, lng, photo).Return(nil)

	// Execute UpdateProfile
	err := service.UpdateProfile(ctx, userID, name, newEmail, phone, address, lat, lng, photo)
	assert.NoError(t, err)

	// Step 2: Verify Email Change
	token := "verification-token"
	verificationToken := &entity.VerificationTokenEntity{
		UserID:    userID,
		Token:     token,
		TokenType: "email_change",
		NewEmail:  newEmail,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Mock expectations for VerifyEmailChange
	mockVerificationTokenRepo.On("GetVerificationToken", ctx, token).Return(verificationToken, nil)
	mockUserRepo.On("UpdateUserEmail", ctx, userID, newEmail).Return(nil)
	mockUserRepo.On("UpdateUserVerificationStatus", ctx, userID, true).Return(nil)
	mockVerificationTokenRepo.On("DeleteVerificationToken", ctx, token).Return(nil)

	// Execute VerifyEmailChange
	err = service.VerifyEmailChange(ctx, token)
	assert.NoError(t, err)

	// Step 3: Try to login with new email (should succeed)
	hashedPassword, _ := utils.HashPassword("password123")
	updatedUser := &entity.UserEntity{
		ID:       userID,
		Email:    newEmail, // Email has been updated
		Password: hashedPassword,
		RoleName: "Customer",
		Name:     name,
		IsVerified: true, // User is verified again
	}

	mockUserRepo.On("GetUserByEmail", ctx, newEmail).Return(updatedUser, nil)
	mockJWTUtil.On("GenerateJWTWithSession", userID, newEmail, "Customer", mock.AnythingOfType("string")).Return("jwt-token", nil)
	mockSessionRepo.On("StoreToken", ctx, userID, mock.AnythingOfType("string"), "jwt-token").Return(nil)

	// Execute SignIn with new email
	user, jwtToken, err := service.SignIn(ctx, entity.UserEntity{
		Email:    newEmail,
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, jwtToken)
	assert.Equal(t, newEmail, user.Email)

	// Verify all expectations
	mockUserRepo.AssertExpectations(t)
	mockVerificationTokenRepo.AssertExpectations(t)
	mockEmailPublisher.AssertExpectations(t)
}

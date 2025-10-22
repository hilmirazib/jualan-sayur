package main

import (
	"context"
	"errors"
	"testing"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"
	"user-service/test/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_UpdateProfile_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)

	ctx := context.Background()
	userID := int64(1)
	name := "John Doe Updated"
	email := "newemail@example.com"
	phone := "081234567890"
	address := "Jl. Sudirman No. 123, Jakarta"
	lat := -6.2088
	lng := 106.8456
	photo := "https://example.com/new-photo.jpg"

	// Current user data (same email, no change)
	currentUser := &entity.UserEntity{
		ID:    userID,
		Email: email, // Same email, no change
		Photo: "https://example.com/old-photo.jpg",
	}

	// Mock expectations
	mockUserRepo.On("GetUserByID", ctx, userID).Return(currentUser, nil)
	mockUserRepo.On("UpdateUserProfile", ctx, userID, name, email, phone, address, lat, lng, photo).Return(nil)

	// Execute
	err := service.UpdateProfile(ctx, userID, name, email, phone, address, lat, lng, photo)

	// Assert
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_UpdateProfile_EmailAlreadyExists(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)

	ctx := context.Background()
	userID := int64(1)
	name := "John Doe"
	email := "existing@example.com"
	phone := "08123456789"
	address := "Jakarta"
	lat := -6.2088
	lng := 106.8456
	photo := "https://example.com/photo.jpg"

	// Current user data (different email)
	currentUser := &entity.UserEntity{
		ID:    userID,
		Email: "old@example.com", // Different email
		Photo: "https://example.com/photo.jpg",
	}

	// Mock expectations
	mockUserRepo.On("GetUserByID", ctx, userID).Return(currentUser, nil)
	existingUser := &entity.UserEntity{ID: 2, Email: email}
	mockUserRepo.On("GetUserByEmailIncludingUnverified", ctx, email).Return(existingUser, nil)

	// Execute
	err := service.UpdateProfile(ctx, userID, name, email, phone, address, lat, lng, photo)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "email already exists", err.Error())
	mockUserRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "UpdateUserProfile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthService_UpdateProfile_SameUserEmail(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)

	ctx := context.Background()
	userID := int64(1)
	name := "John Doe Updated"
	email := "sameuser@example.com"
	phone := "081234567890"
	address := "Jl. Sudirman No. 123, Jakarta"
	lat := -6.2088
	lng := 106.8456
	photo := "https://example.com/new-photo.jpg"

	// Current user data (same email, no change)
	currentUser := &entity.UserEntity{
		ID:    userID,
		Email: email, // Same email, no change
		Photo: "https://example.com/old-photo.jpg",
	}

	// Mock expectations
	mockUserRepo.On("GetUserByID", ctx, userID).Return(currentUser, nil)
	mockUserRepo.On("UpdateUserProfile", ctx, userID, name, email, phone, address, lat, lng, photo).Return(nil)

	// Execute
	err := service.UpdateProfile(ctx, userID, name, email, phone, address, lat, lng, photo)

	// Assert
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_UpdateProfile_InvalidEmail(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)

	ctx := context.Background()
	userID := int64(1)
	name := "John Doe"
	email := "invalid-email-format"
	phone := "08123456789"
	address := "Jakarta"
	lat := -6.2088
	lng := 106.8456
	photo := "https://example.com/photo.jpg"

	// Execute - should fail email validation before calling repository
	err := service.UpdateProfile(ctx, userID, name, email, phone, address, lat, lng, photo)

	// Assert
	assert.Error(t, err)
	// Email validation error should be returned
	mockUserRepo.AssertNotCalled(t, "GetUserByEmailIncludingUnverified", mock.Anything, mock.Anything)
	mockUserRepo.AssertNotCalled(t, "UpdateUserProfile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthService_UpdateProfile_EmptyEmail(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)

	ctx := context.Background()
	userID := int64(1)
	name := "John Doe"
	email := "" // Empty email
	phone := "08123456789"
	address := "Jakarta"
	lat := -6.2088
	lng := 106.8456
	photo := "https://example.com/photo.jpg"

	// Execute - should fail email validation
	err := service.UpdateProfile(ctx, userID, name, email, phone, address, lat, lng, photo)

	// Assert
	assert.Error(t, err)
	mockUserRepo.AssertNotCalled(t, "GetUserByEmailIncludingUnverified", mock.Anything, mock.Anything)
	mockUserRepo.AssertNotCalled(t, "UpdateUserProfile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthService_UpdateProfile_DatabaseError(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	mockVerificationTokenRepo := new(mocks.MockVerificationTokenRepository)
	mockEmailPublisher := new(mocks.MockEmailPublisher)
	mockStorage := new(mocks.MockStorage)
	service := service.NewAuthService(mockUserRepo, nil, nil, mockVerificationTokenRepo, mockEmailPublisher, nil, mockStorage)

	ctx := context.Background()
	userID := int64(1)
	name := "John Doe Updated"
	email := "newemail@example.com"
	phone := "081234567890"
	address := "Jl. Sudirman No. 123, Jakarta"
	lat := -6.2088
	lng := 106.8456
	photo := "https://example.com/new-photo.jpg"

	// Current user data (different email)
	currentUser := &entity.UserEntity{
		ID:    userID,
		Email: "old@example.com", // Different email
		Photo: "https://test.supabase.co/storage/v1/object/public/profile-images/old-photo.jpg",
	}

	// Mock expectations
	mockUserRepo.On("GetUserByID", ctx, userID).Return(currentUser, nil)
	mockUserRepo.On("GetUserByEmailIncludingUnverified", ctx, email).Return(nil, errors.New("record not found"))
	mockStorage.On("DeleteFile", mock.Anything, "", "old-photo.jpg").Return(nil) // Photo cleanup
	// Email change flow mocks
	mockVerificationTokenRepo.On("CreateVerificationToken", ctx, mock.AnythingOfType("*entity.VerificationTokenEntity")).Return(nil)
	mockEmailPublisher.On("SendEmailChangeVerificationEmail", ctx, email, mock.AnythingOfType("string")).Return(nil)
	mockUserRepo.On("UpdateUserVerificationStatus", ctx, userID, false).Return(nil)
	// Profile update fails
	mockUserRepo.On("UpdateUserProfile", ctx, userID, name, "old@example.com", phone, address, lat, lng, photo).Return(errors.New("database connection failed"))

	// Execute
	err := service.UpdateProfile(ctx, userID, name, email, phone, address, lat, lng, photo)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "failed to update profile", err.Error())
	mockUserRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
	mockVerificationTokenRepo.AssertExpectations(t)
	mockEmailPublisher.AssertExpectations(t)
}

func TestAuthService_UpdateProfile_EmailCheckError(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)

	ctx := context.Background()
	userID := int64(1)
	name := "John Doe Updated"
	email := "newemail@example.com"
	phone := "081234567890"
	address := "Jl. Sudirman No. 123, Jakarta"
	lat := -6.2088
	lng := 106.8456
	photo := "https://example.com/new-photo.jpg"

	// Current user data (different email)
	currentUser := &entity.UserEntity{
		ID:    userID,
		Email: "old@example.com", // Different email
		Photo: "https://example.com/old-photo.jpg",
	}

	// Mock expectations - email check fails with unexpected error
	mockUserRepo.On("GetUserByID", ctx, userID).Return(currentUser, nil)
	mockUserRepo.On("GetUserByEmailIncludingUnverified", ctx, email).Return(nil, errors.New("database connection error"))

	// Execute
	err := service.UpdateProfile(ctx, userID, name, email, phone, address, lat, lng, photo)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "unable to verify email availability", err.Error())
	mockUserRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "UpdateUserProfile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

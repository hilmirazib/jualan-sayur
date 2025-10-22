package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_UploadProfileImage_Success_WithOldPhotoCleanup(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockStorage := new(MockStorage)
	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage)

	ctx := context.Background()
	userID := int64(1)
	fileReader := strings.NewReader("fake image content")
	contentType := "image/jpeg"
	filename := "test.jpg"

	oldPhotoURL := "https://test.supabase.co/storage/v1/object/public/profile-images/old-profile-uuid.jpg"
	newPhotoURL := "https://test.supabase.co/storage/v1/object/public/profile-images/new-profile-uuid.jpg"

	currentUser := &entity.UserEntity{
		ID:    userID,
		Photo: oldPhotoURL,
	}

	// Mock expectations
	mockUserRepo.On("GetUserByID", ctx, userID).Return(currentUser, nil)
	mockStorage.On("UploadFile", ctx, "", "", mock.Anything, contentType).Return(newPhotoURL, nil)
	mockUserRepo.On("UpdateUserPhoto", ctx, userID, newPhotoURL).Return(nil)
	mockStorage.On("DeleteFile", ctx, "", "old-profile-uuid.jpg").Return(nil)

	// Execute
	resultURL, err := service.UploadProfileImage(ctx, userID, fileReader, contentType, filename)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, newPhotoURL, resultURL)
	mockUserRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestAuthService_UploadProfileImage_Success_NoExistingPhoto(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockStorage := new(MockStorage)
	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage)

	ctx := context.Background()
	userID := int64(1)
	fileReader := strings.NewReader("fake image content")
	contentType := "image/jpeg"
	filename := "test.jpg"

	newPhotoURL := "https://test.supabase.co/storage/v1/object/public/profile-images/new-profile-uuid.jpg"

	currentUser := &entity.UserEntity{
		ID:    userID,
		Photo: "", // No existing photo
	}

	// Mock expectations
	mockUserRepo.On("GetUserByID", ctx, userID).Return(currentUser, nil)
	mockStorage.On("UploadFile", ctx, "", "", mock.Anything, contentType).Return(newPhotoURL, nil)
	mockUserRepo.On("UpdateUserPhoto", ctx, userID, newPhotoURL).Return(nil)
	// No delete call expected since no old photo

	// Execute
	resultURL, err := service.UploadProfileImage(ctx, userID, fileReader, contentType, filename)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, newPhotoURL, resultURL)
	mockUserRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
	mockStorage.AssertNotCalled(t, "DeleteFile", mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthService_UploadProfileImage_UploadFailure(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockStorage := new(MockStorage)
	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage)

	ctx := context.Background()
	userID := int64(1)
	fileReader := strings.NewReader("fake image content")
	contentType := "image/jpeg"
	filename := "test.jpg"

	currentUser := &entity.UserEntity{
		ID:    userID,
		Photo: "",
	}

	// Mock expectations
	mockUserRepo.On("GetUserByID", ctx, userID).Return(currentUser, nil)
	mockStorage.On("UploadFile", ctx, "", "", mock.Anything, contentType).Return("", errors.New("upload failed"))

	// Execute
	resultURL, err := service.UploadProfileImage(ctx, userID, fileReader, contentType, filename)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, resultURL)
	assert.Equal(t, "failed to upload image", err.Error())
	mockUserRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "UpdateUserPhoto", mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthService_UploadProfileImage_DatabaseUpdateFailure(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockStorage := new(MockStorage)
	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage)

	ctx := context.Background()
	userID := int64(1)
	fileReader := strings.NewReader("fake image content")
	contentType := "image/jpeg"
	filename := "test.jpg"

	newPhotoURL := "https://test.supabase.co/storage/v1/object/public/profile-images/new-profile-uuid.jpg"

	currentUser := &entity.UserEntity{
		ID:    userID,
		Photo: "",
	}

	// Mock expectations
	mockUserRepo.On("GetUserByID", ctx, userID).Return(currentUser, nil)
	mockStorage.On("UploadFile", ctx, "", "", mock.Anything, contentType).Return(newPhotoURL, nil)
	mockUserRepo.On("UpdateUserPhoto", ctx, userID, newPhotoURL).Return(errors.New("database error"))
	mockStorage.On("DeleteFile", ctx, "", "new-profile-uuid.jpg").Return(nil) // Cleanup of uploaded file

	// Execute
	resultURL, err := service.UploadProfileImage(ctx, userID, fileReader, contentType, filename)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, resultURL)
	assert.Equal(t, "failed to update profile", err.Error())
	mockUserRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestAuthService_UploadProfileImage_OldPhotoDeletionFailure(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockStorage := new(MockStorage)
	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage)

	ctx := context.Background()
	userID := int64(1)
	fileReader := strings.NewReader("fake image content")
	contentType := "image/jpeg"
	filename := "test.jpg"

	oldPhotoURL := "https://test.supabase.co/storage/v1/object/public/profile-images/old-profile-uuid.jpg"
	newPhotoURL := "https://test.supabase.co/storage/v1/object/public/profile-images/new-profile-uuid.jpg"

	currentUser := &entity.UserEntity{
		ID:    userID,
		Photo: oldPhotoURL,
	}

	// Mock expectations
	mockUserRepo.On("GetUserByID", ctx, userID).Return(currentUser, nil)
	mockStorage.On("UploadFile", ctx, "", "", mock.Anything, contentType).Return(newPhotoURL, nil)
	mockUserRepo.On("UpdateUserPhoto", ctx, userID, newPhotoURL).Return(nil)
	mockStorage.On("DeleteFile", ctx, "", "old-profile-uuid.jpg").Return(errors.New("delete failed"))

	// Execute
	resultURL, err := service.UploadProfileImage(ctx, userID, fileReader, contentType, filename)

	// Assert - Upload should still succeed even if old photo deletion fails
	assert.NoError(t, err)
	assert.Equal(t, newPhotoURL, resultURL)
	mockUserRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestAuthService_UploadProfileImage_GetUserFailure(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockStorage := new(MockStorage)
	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage)

	ctx := context.Background()
	userID := int64(1)
	fileReader := strings.NewReader("fake image content")
	contentType := "image/jpeg"
	filename := "test.jpg"

	// Mock expectations
	mockUserRepo.On("GetUserByID", ctx, userID).Return(nil, errors.New("user not found"))

	// Execute
	resultURL, err := service.UploadProfileImage(ctx, userID, fileReader, contentType, filename)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, resultURL)
	assert.Equal(t, "failed to get user data", err.Error())
	mockUserRepo.AssertExpectations(t)
	mockStorage.AssertNotCalled(t, "UploadFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

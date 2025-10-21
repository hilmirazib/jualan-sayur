package service

import (
	"context"
	"errors"
	"strings"
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
	mockStorage := new(MockStorage)
	service := NewUserService(mockRepo, nil, nil, nil, nil, nil, mockStorage, &config.Config{})

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
	mockStorage := new(MockStorage)
	service := NewUserService(mockRepo, nil, nil, nil, nil, nil, mockStorage, &config.Config{})

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
	mockStorage := new(MockStorage)
	service := NewUserService(mockUserRepo, nil, nil, mockVerificationTokenRepo, mockEmailPublisher, nil, mockStorage, &config.Config{})

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
	mockStorage := new(MockStorage)
	service := NewUserService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage, &config.Config{})

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
	mockStorage := new(MockStorage)
	service := NewUserService(mockUserRepo, nil, nil, mockVerificationTokenRepo, nil, nil, mockStorage, &config.Config{})

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
	mockStorage := new(MockStorage)
	mockConfig := &config.Config{
		App: config.App{
			JwtSecretKey: "test-secret-key",
			JwtIssuer:    "test-issuer",
		},
	}
	service := NewUserService(mockUserRepo, mockSessionRepo, mockJWTUtil, nil, nil, nil, mockStorage, mockConfig)

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

func TestAuthService_UploadProfileImage_Success_WithOldPhotoCleanup(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockStorage := new(MockStorage)
	service := NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage)

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
	service := NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage)

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
	service := NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage)

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
	service := NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage)

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
	service := NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage)

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
	service := NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage)

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

func TestAuthService_ExtractObjectNameFromURL(t *testing.T) {
	// Setup
	service := &AuthService{}

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "Valid Supabase URL",
			url:      "https://test.supabase.co/storage/v1/object/public/profile-images/profile-uuid.jpg",
			expected: "profile-uuid.jpg",
		},
		{
			name:     "Valid URL with different bucket",
			url:      "https://project.supabase.co/storage/v1/object/public/avatars/user-123.png",
			expected: "user-123.png",
		},
		{
			name:     "Invalid URL - missing storage path",
			url:      "https://test.supabase.co/invalid/path",
			expected: "",
		},
		{
			name:     "Invalid URL - no object name",
			url:      "https://test.supabase.co/storage/v1/object/public/bucket/",
			expected: "",
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.extractObjectNameFromURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
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

func TestAuthService_UpdateProfile_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)

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
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)

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
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)

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
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)

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
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)

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
	mockUserRepo := new(MockUserRepository)
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	mockEmailPublisher := new(MockEmailPublisher)
	mockStorage := new(MockStorage)
	service := NewAuthService(mockUserRepo, nil, nil, mockVerificationTokenRepo, mockEmailPublisher, nil, mockStorage)

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
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)

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

func TestAuthService_VerifyEmailChange_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	service := NewAuthService(mockUserRepo, nil, nil, mockVerificationTokenRepo, nil, nil, nil)

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
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	service := NewAuthService(nil, nil, nil, mockVerificationTokenRepo, nil, nil, nil)

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
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	service := NewAuthService(nil, nil, nil, mockVerificationTokenRepo, nil, nil, nil)

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
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	service := NewAuthService(nil, nil, nil, mockVerificationTokenRepo, nil, nil, nil)

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
	mockUserRepo := new(MockUserRepository)
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	service := NewAuthService(mockUserRepo, nil, nil, mockVerificationTokenRepo, nil, nil, nil)

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
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockJWTUtil := new(MockJWTUtil)
	mockVerificationTokenRepo := new(MockVerificationTokenRepository)
	mockEmailPublisher := new(MockEmailPublisher)
	service := NewAuthService(mockUserRepo, mockSessionRepo, mockJWTUtil, mockVerificationTokenRepo, mockEmailPublisher, nil, nil)

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

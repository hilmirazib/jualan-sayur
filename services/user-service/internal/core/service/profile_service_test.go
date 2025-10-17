package service

import (
	"context"
	"errors"
	"testing"
	"user-service/config"
	"user-service/internal/core/domain/entity"

	"github.com/stretchr/testify/assert"
)

func TestUserService_GetProfile_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockStorage := new(MockStorage)
	service := NewUserService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage, &config.Config{})

	ctx := context.Background()
	userID := int64(1)

	expectedUser := &entity.UserEntity{
		ID:       userID,
		Email:    "test@example.com",
		Name:     "Test User",
		RoleName: "Customer",
		Phone:    "08123456789",
		Address:  "Jl. Test No. 123",
		Lat:      "-6.2088",
		Lng:      "106.8456",
		Photo:    "https://example.com/photo.jpg",
	}

	// Mock expectations
	mockUserRepo.On("GetUserByID", ctx, userID).Return(expectedUser, nil)

	// Execute
	user, err := service.GetProfile(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser, user)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_GetProfile_UserNotFound(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockStorage := new(MockStorage)
	service := NewUserService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage, &config.Config{})

	ctx := context.Background()
	userID := int64(999)

	// Mock expectations - user not found
	mockUserRepo.On("GetUserByID", ctx, userID).Return(nil, errors.New("record not found"))

	// Execute
	user, err := service.GetProfile(ctx, userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "user not found", err.Error())
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_GetProfile_DatabaseError(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockStorage := new(MockStorage)
	service := NewUserService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage, &config.Config{})

	ctx := context.Background()
	userID := int64(1)

	// Mock expectations - database connection error
	mockUserRepo.On("GetUserByID", ctx, userID).Return(nil, errors.New("database connection failed"))

	// Execute
	user, err := service.GetProfile(ctx, userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "database connection failed", err.Error())
	mockUserRepo.AssertExpectations(t)
}

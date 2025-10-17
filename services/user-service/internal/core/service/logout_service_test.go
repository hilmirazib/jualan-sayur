package service

import (
	"context"
	"errors"
	"testing"
	"user-service/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_Logout_Success(t *testing.T) {
	// Setup
	mockSessionRepo := new(MockSessionRepository)
	mockBlacklistRepo := new(MockBlacklistTokenRepository)
	service := NewUserService(nil, mockSessionRepo, nil, nil, nil, mockBlacklistRepo, &config.Config{})

	ctx := context.Background()
	userID := int64(1)
	sessionID := "session-123"
	tokenString := "jwt-token-123"
	tokenExpiresAt := int64(1640995200) // 2022-01-01 00:00:00 UTC

	// Mock expectations
	mockSessionRepo.On("DeleteToken", ctx, userID, sessionID).Return(nil)
	mockBlacklistRepo.On("AddToBlacklist", ctx, mock.AnythingOfType("string"), tokenExpiresAt).Return(nil)

	// Execute
	err := service.Logout(ctx, userID, sessionID, tokenString, tokenExpiresAt)

	// Assert
	assert.NoError(t, err)
	mockSessionRepo.AssertExpectations(t)
	mockBlacklistRepo.AssertExpectations(t)
}

func TestUserService_Logout_SessionDeletionFails(t *testing.T) {
	// Setup
	mockSessionRepo := new(MockSessionRepository)
	service := NewUserService(nil, mockSessionRepo, nil, nil, nil, nil, &config.Config{})

	ctx := context.Background()
	userID := int64(1)
	sessionID := "session-123"

	// Mock expectations - session deletion fails
	mockSessionRepo.On("DeleteToken", ctx, userID, sessionID).Return(errors.New("redis connection failed"))

	// Execute
	err := service.Logout(ctx, userID, sessionID, "", 0)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "failed to logout", err.Error())
	mockSessionRepo.AssertExpectations(t)
}

func TestUserService_Logout_BlacklistFailureIgnored(t *testing.T) {
	// Setup
	mockSessionRepo := new(MockSessionRepository)
	mockBlacklistRepo := new(MockBlacklistTokenRepository)
	service := NewUserService(nil, mockSessionRepo, nil, nil, nil, mockBlacklistRepo, &config.Config{})

	ctx := context.Background()
	userID := int64(1)
	sessionID := "session-123"
	tokenString := "jwt-token-123"
	tokenExpiresAt := int64(1640995200)

	// Mock expectations
	mockSessionRepo.On("DeleteToken", ctx, userID, sessionID).Return(nil)
	mockBlacklistRepo.On("AddToBlacklist", ctx, mock.AnythingOfType("string"), tokenExpiresAt).Return(errors.New("database connection failed"))

	// Execute - should succeed even if blacklist fails
	err := service.Logout(ctx, userID, sessionID, tokenString, tokenExpiresAt)

	// Assert - logout should succeed despite blacklist failure
	assert.NoError(t, err)
	mockSessionRepo.AssertExpectations(t)
	mockBlacklistRepo.AssertExpectations(t)
}

func TestUserService_Logout_WithoutToken(t *testing.T) {
	// Setup
	mockSessionRepo := new(MockSessionRepository)
	service := NewUserService(nil, mockSessionRepo, nil, nil, nil, nil, &config.Config{})

	ctx := context.Background()
	userID := int64(1)
	sessionID := "session-123"

	// Mock expectations
	mockSessionRepo.On("DeleteToken", ctx, userID, sessionID).Return(nil)

	// Execute - logout without token (backward compatibility)
	err := service.Logout(ctx, userID, sessionID, "", 0)

	// Assert
	assert.NoError(t, err)
	mockSessionRepo.AssertExpectations(t)
}

func TestUserService_Logout_WithoutExpiration(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockBlacklistRepo := new(MockBlacklistTokenRepository)
	service := NewUserService(mockUserRepo, mockSessionRepo, nil, nil, nil, mockBlacklistRepo, &config.Config{})

	ctx := context.Background()
	userID := int64(1)
	sessionID := "session-123"
	tokenString := "jwt-token-123"

	// Mock expectations
	mockSessionRepo.On("DeleteToken", ctx, userID, sessionID).Return(nil)
	// Blacklist should not be called when expiresAt is 0
	mockBlacklistRepo.AssertNotCalled(t, "AddToBlacklist", mock.Anything, mock.Anything, mock.Anything)

	// Execute - logout with token but no expiration
	err := service.Logout(ctx, userID, sessionID, tokenString, 0)

	// Assert
	assert.NoError(t, err)
	mockSessionRepo.AssertExpectations(t)
}

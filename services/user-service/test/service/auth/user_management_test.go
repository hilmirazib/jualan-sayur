package main

import (
	"context"
	"errors"
	"testing"
	"user-service/config"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"
	"user-service/test/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_CreateUserAccount_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	mockVerificationTokenRepo := new(mocks.MockVerificationTokenRepository)
	mockEmailPublisher := new(mocks.MockEmailPublisher)
	mockStorage := new(mocks.MockStorage)
	service := service.NewUserService(mockUserRepo, nil, nil, mockVerificationTokenRepo, mockEmailPublisher, nil, mockStorage, &config.Config{})

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
	mockUserRepo := new(mocks.MockUserRepository)
	mockStorage := new(mocks.MockStorage)
	service := service.NewUserService(mockUserRepo, nil, nil, nil, nil, nil, mockStorage, &config.Config{})

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
	mockUserRepo := new(mocks.MockUserRepository)
	mockVerificationTokenRepo := new(mocks.MockVerificationTokenRepository)
	mockStorage := new(mocks.MockStorage)
	service := service.NewUserService(mockUserRepo, nil, nil, mockVerificationTokenRepo, nil, nil, mockStorage, &config.Config{})

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

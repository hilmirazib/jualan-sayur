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

// Mock repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserEntity), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *entity.UserEntity) (*entity.UserEntity, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserEntity), args.Error(1)
}

func (m *MockUserRepository) GetRoleByName(ctx context.Context, name string) (*entity.RoleEntity, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.RoleEntity), args.Error(1)
}

func TestUserService_SignIn_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	mockConfig := &config.Config{
		App: config.App{
			JwtSecretKey: "test-secret-key",
			JwtIssuer:    "test-issuer",
		},
	}
	service := &UserService{
		userRepo: mockRepo,
		config:   mockConfig,
	}

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"

	// Hash the password for the mock user
	hashedPassword, _ := utils.HashPassword(password)
	expectedUser := &entity.UserEntity{
		ID:       1,
		Email:    email,
		Password: hashedPassword,
		RoleName: "user",
	}

	// Mock expectations
	mockRepo.On("GetUserByEmail", ctx, email).Return(expectedUser, nil)

	// Execute
	user, token, err := service.SignIn(ctx, entity.UserEntity{
		Email:    email,
		Password: password,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Email, user.Email)

	mockRepo.AssertExpectations(t)
}

func TestUserService_SignIn_UserNotFound(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	service := &UserService{
		userRepo: mockRepo,
		config:   &config.Config{},
	}

	ctx := context.Background()
	email := "notfound@example.com"

	// Mock expectations - return nil user and error
	mockRepo.On("GetUserByEmail", ctx, email).Return(nil, errors.New("user not found"))

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
	service := &UserService{
		userRepo: mockRepo,
		config:   &config.Config{},
	}

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

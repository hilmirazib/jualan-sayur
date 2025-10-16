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

func (m *MockUserRepository) UpdateUserVerificationStatus(ctx context.Context, userID int64, isVerified bool) error {
	args := m.Called(ctx, userID, isVerified)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByEmailIncludingUnverified(ctx context.Context, email string) (*entity.UserEntity, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserEntity), args.Error(1)
}

// Mock session repository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) StoreToken(ctx context.Context, userID int64, sessionID, token string) error {
	args := m.Called(ctx, userID, sessionID, token)
	return args.Error(0)
}

func (m *MockSessionRepository) GetToken(ctx context.Context, userID int64, sessionID string) (string, error) {
	args := m.Called(ctx, userID, sessionID)
	return args.String(0), args.Error(1)
}

func (m *MockSessionRepository) DeleteToken(ctx context.Context, userID int64, sessionID string) error {
	args := m.Called(ctx, userID, sessionID)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteAllUserTokens(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) ValidateToken(ctx context.Context, userID int64, sessionID string, token string) bool {
	args := m.Called(ctx, userID, sessionID, token)
	return args.Bool(0)
}

func (m *MockSessionRepository) GetUserSessions(ctx context.Context, userID int64) ([]entity.SessionInfo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.SessionInfo), args.Error(1)
}

// Mock JWT utility
type MockJWTUtil struct {
	mock.Mock
}

func (m *MockJWTUtil) GenerateJWT(userID int64, email, roleName string) (string, error) {
	args := m.Called(userID, email, roleName)
	return args.String(0), args.Error(1)
}

func (m *MockJWTUtil) GenerateJWTWithSession(userID int64, email, role, sessionID string) (string, error) {
	args := m.Called(userID, email, role, sessionID)
	return args.String(0), args.Error(1)
}

func (m *MockJWTUtil) ValidateJWT(tokenString string) (*utils.JWTClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.JWTClaims), args.Error(1)
}

// Mock verification token repository
type MockVerificationTokenRepository struct {
	mock.Mock
}

func (m *MockVerificationTokenRepository) CreateVerificationToken(ctx context.Context, token *entity.VerificationTokenEntity) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockVerificationTokenRepository) GetVerificationToken(ctx context.Context, token string) (*entity.VerificationTokenEntity, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.VerificationTokenEntity), args.Error(1)
}

func (m *MockVerificationTokenRepository) DeleteVerificationToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// Mock email publisher
type MockEmailPublisher struct {
	mock.Mock
}

func (m *MockEmailPublisher) SendVerificationEmail(ctx context.Context, email, token string) error {
	args := m.Called(ctx, email, token)
	return args.Error(0)
}

func (m *MockEmailPublisher) SendPasswordResetEmail(ctx context.Context, email, token string) error {
	args := m.Called(ctx, email, token)
	return args.Error(0)
}

func TestUserService_SignIn_UserNotFound(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, nil, nil, nil, nil, &config.Config{})

	ctx := context.Background()
	email := "notfound@example.com"

	// Mock expectations - return nil user and error
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

	// Mock expectations - existing user found
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

	// Hash the password for the mock admin user
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

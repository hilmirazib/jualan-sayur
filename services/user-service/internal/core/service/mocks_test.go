package service

import (
	"context"
	"user-service/internal/core/domain/entity"
	"user-service/utils"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository mocks the user repository
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

func (m *MockUserRepository) UpdateUserPassword(ctx context.Context, userID int64, hashedPassword string) error {
	args := m.Called(ctx, userID, hashedPassword)
	return args.Error(0)
}

// MockSessionRepository mocks the session repository
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

// MockJWTUtil mocks the JWT utility
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

// MockVerificationTokenRepository mocks the verification token repository
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

// MockEmailPublisher mocks the email publisher
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

// MockBlacklistTokenRepository mocks the blacklist token repository
type MockBlacklistTokenRepository struct {
	mock.Mock
}

func (m *MockBlacklistTokenRepository) AddToBlacklist(ctx context.Context, tokenHash string, expiresAt int64) error {
	args := m.Called(ctx, tokenHash, expiresAt)
	return args.Error(0)
}

func (m *MockBlacklistTokenRepository) IsTokenBlacklisted(ctx context.Context, tokenHash string) bool {
	args := m.Called(ctx, tokenHash)
	return args.Bool(0)
}

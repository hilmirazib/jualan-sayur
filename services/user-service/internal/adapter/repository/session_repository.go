package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"user-service/config"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/port"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type SessionRepository struct {
	redisClient *redis.Client
	config      *config.Config
}

func NewSessionRepository(redisClient *redis.Client, cfg *config.Config) port.SessionInterface {
	return &SessionRepository{
		redisClient: redisClient,
		config:      cfg,
	}
}

func (s *SessionRepository) StoreToken(ctx context.Context, userID int64, sessionID string, token string) error {
	key := s.getSessionKey(userID, sessionID)

	err := s.redisClient.Set(ctx, key, token, 24*time.Hour).Err()
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Str("session_id", sessionID).Msg("[SessionRepository-StoreToken] Failed to store token")
		return err
	}

	// Add session to user's active sessions list
	userSessionsKey := s.getUserSessionsKey(userID)
	sessionInfo := entity.SessionInfo{
		SessionID: sessionID,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	sessionData, err := json.Marshal(sessionInfo)
	if err != nil {
		log.Error().Err(err).Msg("[SessionRepository-StoreToken] Failed to marshal session info")
		return err
	}

	err = s.redisClient.HSet(ctx, userSessionsKey, sessionID, sessionData).Err()
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Str("session_id", sessionID).Msg("[SessionRepository-StoreToken] Failed to store session info")
		return err
	}

	// Set expiration for user sessions key as well
	s.redisClient.Expire(ctx, userSessionsKey, 24*time.Hour)

	log.Info().Int64("user_id", userID).Str("session_id", sessionID).Msg("[SessionRepository-StoreToken] Token stored successfully")
	return nil
}

// GetToken retrieves JWT token from Redis
func (s *SessionRepository) GetToken(ctx context.Context, userID int64, sessionID string) (string, error) {
	key := s.getSessionKey(userID, sessionID)

	token, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Warn().Int64("user_id", userID).Str("session_id", sessionID).Msg("[SessionRepository-GetToken] Token not found")
		return "", fmt.Errorf("token not found")
	}
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Str("session_id", sessionID).Msg("[SessionRepository-GetToken] Failed to get token")
		return "", err
	}

	return token, nil
}

// DeleteToken removes JWT token from Redis
func (s *SessionRepository) DeleteToken(ctx context.Context, userID int64, sessionID string) error {
	key := s.getSessionKey(userID, sessionID)
	userSessionsKey := s.getUserSessionsKey(userID)

	// Delete token
	err := s.redisClient.Del(ctx, key).Err()
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Str("session_id", sessionID).Msg("[SessionRepository-DeleteToken] Failed to delete token")
		return err
	}

	// Remove from user sessions
	err = s.redisClient.HDel(ctx, userSessionsKey, sessionID).Err()
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Str("session_id", sessionID).Msg("[SessionRepository-DeleteToken] Failed to delete session info")
		return err
	}

	log.Info().Int64("user_id", userID).Str("session_id", sessionID).Msg("[SessionRepository-DeleteToken] Token deleted successfully")
	return nil
}

// DeleteAllUserTokens removes all tokens for a user
func (s *SessionRepository) DeleteAllUserTokens(ctx context.Context, userID int64) error {
	userSessionsKey := s.getUserSessionsKey(userID)

	// Get all session IDs for user
	sessionIDs, err := s.redisClient.HKeys(ctx, userSessionsKey).Result()
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("[SessionRepository-DeleteAllUserTokens] Failed to get session keys")
		return err
	}

	// Delete all individual session tokens
	for _, sessionID := range sessionIDs {
		key := s.getSessionKey(userID, sessionID)
		if err := s.redisClient.Del(ctx, key).Err(); err != nil {
			log.Error().Err(err).Int64("user_id", userID).Str("session_id", sessionID).Msg("[SessionRepository-DeleteAllUserTokens] Failed to delete token")
		}
	}

	// Delete user sessions hash
	err = s.redisClient.Del(ctx, userSessionsKey).Err()
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("[SessionRepository-DeleteAllUserTokens] Failed to delete user sessions")
		return err
	}

	log.Info().Int64("user_id", userID).Int("sessions_deleted", len(sessionIDs)).Msg("[SessionRepository-DeleteAllUserTokens] All user tokens deleted")
	return nil
}

// ValidateToken checks if token exists and matches
func (s *SessionRepository) ValidateToken(ctx context.Context, userID int64, sessionID string, token string) bool {
	storedToken, err := s.GetToken(ctx, userID, sessionID)
	if err != nil {
		return false
	}

	return storedToken == token
}

// GetUserSessions returns all active sessions for a user
func (s *SessionRepository) GetUserSessions(ctx context.Context, userID int64) ([]entity.SessionInfo, error) {
	userSessionsKey := s.getUserSessionsKey(userID)

	sessionData, err := s.redisClient.HGetAll(ctx, userSessionsKey).Result()
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("[SessionRepository-GetUserSessions] Failed to get user sessions")
		return nil, err
	}

	var sessions []entity.SessionInfo
	for _, data := range sessionData {
		var session entity.SessionInfo
		if err := json.Unmarshal([]byte(data), &session); err != nil {
			log.Error().Err(err).Int64("user_id", userID).Msg("[SessionRepository-GetUserSessions] Failed to unmarshal session data")
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Helper methods
func (s *SessionRepository) getSessionKey(userID int64, sessionID string) string {
	return fmt.Sprintf("session:%d:%s", userID, sessionID)
}

func (s *SessionRepository) getUserSessionsKey(userID int64) string {
	return fmt.Sprintf("user_sessions:%d", userID)
}

// GenerateSessionID generates a unique session ID
func GenerateSessionID() string {
	return uuid.New().String()
}

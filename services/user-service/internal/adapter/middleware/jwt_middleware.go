package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"user-service/config"
	"user-service/internal/core/port"
	"user-service/utils"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// JWTMiddleware creates JWT authentication middleware with Redis session validation and blacklist check
func JWTMiddleware(cfg *config.Config, sessionRepo port.SessionInterface, blacklistRepo port.BlacklistTokenInterface) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Warn().Msg("[JWTMiddleware] Missing authorization header")
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"message": "Authorization header required",
					"data":    nil,
				})
			}

			// Check Bearer token format
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				log.Warn().Str("auth_header", authHeader).Msg("[JWTMiddleware] Invalid authorization header format")
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"message": "Invalid authorization header format. Use: Bearer <token>",
					"data":    nil,
				})
			}

			tokenString := tokenParts[1]

			// Validate JWT signature first
			claims, err := utils.ValidateJWT(cfg, tokenString)
			if err != nil {
				log.Warn().Err(err).Msg("[JWTMiddleware] Invalid JWT signature")
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"message": "Invalid or expired token",
					"data":    nil,
				})
			}

			// Check if token is blacklisted
			if blacklistRepo != nil {
				hash := sha256.Sum256([]byte(tokenString))
				tokenHash := hex.EncodeToString(hash[:])

				if blacklistRepo.IsTokenBlacklisted(c.Request().Context(), tokenHash) {
					log.Warn().
						Int64("user_id", claims.UserID).
						Str("session_id", claims.SessionID).
						Msg("[JWTMiddleware] Token is blacklisted")
					return c.JSON(http.StatusUnauthorized, map[string]interface{}{
						"message": "Token has been revoked",
						"data":    nil,
					})
				}
			}

			// Validate session in Redis
			if claims.SessionID != "" {
				isValid := sessionRepo.ValidateToken(c.Request().Context(), claims.UserID, claims.SessionID, tokenString)
				if !isValid {
					log.Warn().
						Int64("user_id", claims.UserID).
						Str("session_id", claims.SessionID).
						Msg("[JWTMiddleware] Session not found in Redis")
					return c.JSON(http.StatusUnauthorized, map[string]interface{}{
						"message": "Session expired or invalid",
						"data":    nil,
					})
				}
			} else {
				// For backward compatibility, if no session_id in token, still allow
				log.Warn().
					Int64("user_id", claims.UserID).
					Msg("[JWTMiddleware] Token without session_id (backward compatibility)")
			}

			// Set user information in context
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_role", claims.RoleName)
			c.Set("session_id", claims.SessionID)
			c.Set("exp", claims.ExpiresAt.Unix()) // Set expiration time for logout

			log.Info().
				Int64("user_id", claims.UserID).
				Str("email", claims.Email).
				Str("role", claims.RoleName).
				Str("session_id", claims.SessionID).
				Msg("[JWTMiddleware] Token and session validated successfully")

			return next(c)
		}
	}
}

// OptionalJWTMiddleware creates optional JWT middleware (doesn't fail if no token)
func OptionalJWTMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				// No token provided, continue without authentication
				return next(c)
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				// Invalid format, continue without authentication
				return next(c)
			}

			tokenString := tokenParts[1]

			// Try to validate token
			claims, err := utils.ValidateJWT(cfg, tokenString)
			if err != nil {
				// Invalid token, continue without authentication
				log.Debug().Err(err).Msg("[OptionalJWTMiddleware] Invalid token, continuing without auth")
				return next(c)
			}

			// Set user information in context
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_role", claims.RoleName)

			return next(c)
		}
	}
}

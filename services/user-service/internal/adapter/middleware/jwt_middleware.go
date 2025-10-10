package middleware

import (
	"net/http"
	"strings"
	"user-service/config"
	"user-service/utils"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// JWTMiddleware creates JWT authentication middleware
func JWTMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Warn().Msg("[JWTMiddleware] Missing authorization header")
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "Authorization header required",
				})
			}

			// Check Bearer token format
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				log.Warn().Str("auth_header", authHeader).Msg("[JWTMiddleware] Invalid authorization header format")
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "Invalid authorization header format. Use: Bearer <token>",
				})
			}

			tokenString := tokenParts[1]

			// Validate token
			claims, err := utils.ValidateJWT(cfg, tokenString)
			if err != nil {
				log.Warn().Err(err).Msg("[JWTMiddleware] Invalid token")
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "Invalid or expired token",
				})
			}

			// Set user information in context
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_role", claims.RoleName)

			log.Info().
				Int64("user_id", claims.UserID).
				Str("email", claims.Email).
				Str("role", claims.RoleName).
				Msg("[JWTMiddleware] Token validated successfully")

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

package utils

import (
	"time"
	"user-service/config"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	RoleName  string `json:"role_name"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token for the user
func GenerateJWT(cfg *config.Config, userID int64, email, roleName string) (string, error) {
	return GenerateJWTWithSession(cfg, userID, email, roleName, "")
}

// GenerateJWTWithSession generates a JWT token with session ID for the user
func GenerateJWTWithSession(cfg *config.Config, userID int64, email, roleName, sessionID string) (string, error) {
	// Token expiration time (24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create the JWT claims
	claims := &JWTClaims{
		UserID:    userID,
		Email:     email,
		RoleName:  roleName,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.App.JwtIssuer,
			Subject:   email,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString([]byte(cfg.App.JwtSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates and parses a JWT token
func ValidateJWT(cfg *config.Config, tokenString string) (*JWTClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(cfg.App.JwtSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// Extract claims
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

// JWTUtil implements JWTInterface
type JWTUtil struct {
	config *config.Config
}

// NewJWTUtil creates a new JWTUtil instance
func NewJWTUtil(cfg *config.Config) *JWTUtil {
	return &JWTUtil{
		config: cfg,
	}
}

// GenerateJWT generates a JWT token for the user
func (j *JWTUtil) GenerateJWT(userID int64, email, roleName string) (string, error) {
	return GenerateJWT(j.config, userID, email, roleName)
}

// GenerateJWTWithSession generates a JWT token with session ID for the user
func (j *JWTUtil) GenerateJWTWithSession(userID int64, email, roleName, sessionID string) (string, error) {
	return GenerateJWTWithSession(j.config, userID, email, roleName, sessionID)
}

// ValidateJWT validates and parses a JWT token
func (j *JWTUtil) ValidateJWT(tokenString string) (*JWTClaims, error) {
	return ValidateJWT(j.config, tokenString)
}

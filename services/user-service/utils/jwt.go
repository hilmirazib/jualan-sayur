package utils

import (
	"time"
	"user-service/config"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	RoleName  string `json:"role_name"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

func GenerateJWT(cfg *config.Config, userID int64, email, roleName string) (string, error) {
	return GenerateJWTWithSession(cfg, userID, email, roleName, "")
}

func GenerateJWTWithSession(cfg *config.Config, userID int64, email, roleName, sessionID string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(cfg.App.JwtSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWT(cfg *config.Config, tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(cfg.App.JwtSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

type JWTUtil struct {
	config *config.Config
}

func NewJWTUtil(cfg *config.Config) *JWTUtil {
	return &JWTUtil{
		config: cfg,
	}
}

func (j *JWTUtil) GenerateJWT(userID int64, email, roleName string) (string, error) {
	return GenerateJWT(j.config, userID, email, roleName)
}

func (j *JWTUtil) GenerateJWTWithSession(userID int64, email, roleName, sessionID string) (string, error) {
	return GenerateJWTWithSession(j.config, userID, email, roleName, sessionID)
}

func (j *JWTUtil) ValidateJWT(tokenString string) (*JWTClaims, error) {
	return ValidateJWT(j.config, tokenString)
}

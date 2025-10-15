package port

import "user-service/utils"

type JWTInterface interface {
	GenerateJWT(userID int64, email, roleName string) (string, error)
	GenerateJWTWithSession(userID int64, email, roleName, sessionID string) (string, error)
	ValidateJWT(tokenString string) (*utils.JWTClaims, error)
}

package handler

import (
	"net/http"
	"user-service/internal/core/port"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type AdminHandlerInterface interface {
	AdminCheck(ctx echo.Context) error
}

type AdminHandler struct {
	userService port.UserServiceInterface
}

func (a *AdminHandler) AdminCheck(c echo.Context) error {
	userID := c.Get("user_id").(int64)
	email := c.Get("user_email").(string)
	role := c.Get("user_role").(string)
	sessionID := c.Get("session_id").(string)

	resp := map[string]interface{}{
		"message": "Authentication successful",
		"data": map[string]interface{}{
			"user_id":    userID,
			"email":      email,
			"role":       role,
			"session_id": sessionID,
		},
	}

	log.Info().
		Int64("user_id", userID).
		Str("email", email).
		Str("role", role).
		Str("session_id", sessionID).
		Msg("[AdminHandler-AdminCheck] Admin authentication check successful")

	return c.JSON(http.StatusOK, resp)
}

func NewAdminHandler(userService port.UserServiceInterface) AdminHandlerInterface {
	return &AdminHandler{
		userService: userService,
	}
}

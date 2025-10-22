package handler

import (
	"net/http"
	"user-service/internal/core/port"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type RoleHandlerInterface interface {
	GetAllRoles(c echo.Context) error
}

type RoleHandler struct {
	roleService port.RoleServiceInterface
}

func (h *RoleHandler) GetAllRoles(c echo.Context) error {
	search := c.QueryParam("search")

	roles, err := h.roleService.GetAllRoles(c.Request().Context(), search)
	if err != nil {
		log.Error().Err(err).Str("search", search).Msg("[RoleHandler-GetAllRoles] Failed to get roles")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to retrieve roles",
			"data":    nil,
		})
	}

	// Transform to response format
	var roleData []map[string]interface{}
	for _, role := range roles {
		roleData = append(roleData, map[string]interface{}{
			"id":   role.ID,
			"name": role.Name,
		})
	}

	log.Info().Int("count", len(roles)).Str("search", search).Msg("[RoleHandler-GetAllRoles] Roles retrieved successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Roles retrieved successfully",
		"data":    roleData,
	})
}

func NewRoleHandler(roleService port.RoleServiceInterface) RoleHandlerInterface {
	return &RoleHandler{
		roleService: roleService,
	}
}

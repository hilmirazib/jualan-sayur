package handler

import (
	"fmt"
	"net/http"
	"user-service/internal/core/port"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type RoleHandlerInterface interface {
	GetAllRoles(c echo.Context) error
	GetRoleByID(c echo.Context) error
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

func (h *RoleHandler) GetRoleByID(c echo.Context) error {
	idParam := c.Param("id")

	// Convert string ID to int64
	var id int64
	if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil {
		log.Warn().Str("id_param", idParam).Msg("[RoleHandler-GetRoleByID] Invalid ID format")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid role ID format",
			"data":    nil,
		})
	}

	role, err := h.roleService.GetRoleByID(c.Request().Context(), id)
	if err != nil {
		log.Error().Err(err).Int64("role_id", id).Msg("[RoleHandler-GetRoleByID] Failed to get role by ID")

		if err.Error() == "record not found" {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"message": "Role not found",
				"data":    nil,
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to retrieve role",
			"data":    nil,
		})
	}

	// Transform users to response format
	var userData []map[string]interface{}
	for _, user := range role.Users {
		userData = append(userData, map[string]interface{}{
			"id":   user.ID,
			"name": user.Name,
		})
	}

	// Response data
	roleData := map[string]interface{}{
		"id":    role.ID,
		"name":  role.Name,
		"users": userData,
	}

	log.Info().Int64("role_id", id).Int("users_count", len(userData)).Msg("[RoleHandler-GetRoleByID] Role retrieved successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Role retrieved successfully",
		"data":    roleData,
	})
}

func NewRoleHandler(roleService port.RoleServiceInterface) RoleHandlerInterface {
	return &RoleHandler{
		roleService: roleService,
	}
}

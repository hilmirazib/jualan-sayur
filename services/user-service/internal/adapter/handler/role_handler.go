package handler

import (
	"fmt"
	"net/http"
	"strings"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/core/port"

	myvalidator "user-service/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type RoleHandlerInterface interface {
	GetAllRoles(c echo.Context) error
	GetRoleByID(c echo.Context) error
	CreateRole(c echo.Context) error
	UpdateRole(c echo.Context) error
	DeleteRole(c echo.Context) error
}

type RoleHandler struct {
	roleService port.RoleServiceInterface
	validator   *myvalidator.Validator
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

func (h *RoleHandler) CreateRole(c echo.Context) error {
	// Bind request
	var req request.CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		log.Warn().Err(err).Msg("[RoleHandler-CreateRole] Failed to bind request")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid request format",
			"data":    nil,
		})
	}

	// Validate request
	if err := h.validator.Validate(&req); err != nil {
		log.Error().Err(err).Msg("[RoleHandler-CreateRole] Validation failed")
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"message": err.Error(),
			"data":    nil,
		})
	}

	// Create role
	role, err := h.roleService.CreateRole(c.Request().Context(), req.Name)
	if err != nil {
		log.Error().Err(err).Str("role_name", req.Name).Msg("[RoleHandler-CreateRole] Failed to create role")

		// Check for duplicate role error
		if err.Error() == "role with name 'Super Admin' already exists" ||
		   err.Error() == "role with name 'Customer' already exists" ||
		   strings.Contains(err.Error(), "already exists") {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to create role",
			"data":    nil,
		})
	}

	log.Info().Int64("role_id", role.ID).Str("role_name", role.Name).Msg("[RoleHandler-CreateRole] Role created successfully")
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Role created successfully",
		"data":    nil,
	})
}

func (h *RoleHandler) UpdateRole(c echo.Context) error {
	// Get role ID from URL parameter
	idParam := c.Param("id")

	// Convert string ID to int64
	var id int64
	if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil {
		log.Warn().Str("id_param", idParam).Msg("[RoleHandler-UpdateRole] Invalid ID format")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid role ID format",
			"data":    nil,
		})
	}

	// Bind request
	var req request.CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		log.Warn().Err(err).Int64("role_id", id).Msg("[RoleHandler-UpdateRole] Failed to bind request")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid request format",
			"data":    nil,
		})
	}

	// Validate request
	if err := h.validator.Validate(&req); err != nil {
		log.Error().Err(err).Int64("role_id", id).Msg("[RoleHandler-UpdateRole] Validation failed")
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"message": err.Error(),
			"data":    nil,
		})
	}

	// Update role
	role, err := h.roleService.UpdateRole(c.Request().Context(), id, req.Name)
	if err != nil {
		log.Error().Err(err).Int64("role_id", id).Str("role_name", req.Name).Msg("[RoleHandler-UpdateRole] Failed to update role")

		// Check for specific errors
		if err.Error() == "role not found" {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"message": "Role not found",
				"data":    nil,
			})
		}

		if strings.Contains(err.Error(), "already exists") {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to update role",
			"data":    nil,
		})
	}

	log.Info().Int64("role_id", id).Str("role_name", role.Name).Msg("[RoleHandler-UpdateRole] Role updated successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Role updated successfully",
		"data":    nil,
	})
}

func (h *RoleHandler) DeleteRole(c echo.Context) error {
	// Get role ID from URL parameter
	idParam := c.Param("id")

	// Convert string ID to int64
	var id int64
	if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil {
		log.Warn().Str("id_param", idParam).Msg("[RoleHandler-DeleteRole] Invalid ID format")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid role ID format",
			"data":    nil,
		})
	}

	// Delete role
	err := h.roleService.DeleteRole(c.Request().Context(), id)
	if err != nil {
		log.Error().Err(err).Int64("role_id", id).Msg("[RoleHandler-DeleteRole] Failed to delete role")

		// Check for specific errors
		if err.Error() == "role not found" {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"message": "Role not found",
				"data":    nil,
			})
		}

		if strings.Contains(err.Error(), "currently assigned to users") {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to delete role",
			"data":    nil,
		})
	}

	log.Info().Int64("role_id", id).Msg("[RoleHandler-DeleteRole] Role deleted successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Role deleted successfully",
		"data":    nil,
	})
}

func NewRoleHandler(roleService port.RoleServiceInterface) RoleHandlerInterface {
	return &RoleHandler{
		roleService: roleService,
		validator:   myvalidator.NewValidator(),
	}
}

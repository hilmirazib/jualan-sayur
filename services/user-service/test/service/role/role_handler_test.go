package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"user-service/internal/core/domain/entity"
	"user-service/internal/adapter/handler"
	"user-service/test/service/mocks"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRoleHandler_GetAllRoles_Success(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}
	expectedRoles := []entity.RoleEntity{
		{ID: 1, Name: "Super Admin"},
		{ID: 2, Name: "Customer"},
	}
	mockRoleService.On("GetAllRoles", mock.Anything, "").Return(expectedRoles, nil)

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.GetAllRoles(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Roles retrieved successfully", response["message"])

	data := response["data"].([]interface{})
	assert.Len(t, data, 2)

	// Check first role
	role1 := data[0].(map[string]interface{})
	assert.Equal(t, float64(1), role1["id"])
	assert.Equal(t, "Super Admin", role1["name"])

	// Check second role
	role2 := data[1].(map[string]interface{})
	assert.Equal(t, float64(2), role2["id"])
	assert.Equal(t, "Customer", role2["name"])

	mockRoleService.AssertExpectations(t)
}

func TestRoleHandler_GetAllRoles_WithSearch(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles?search=admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().URL.RawQuery = "search=admin"

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}
	expectedRoles := []entity.RoleEntity{
		{ID: 1, Name: "Super Admin"},
	}
	mockRoleService.On("GetAllRoles", mock.Anything, "admin").Return(expectedRoles, nil)

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.GetAllRoles(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Roles retrieved successfully", response["message"])

	data := response["data"].([]interface{})
	assert.Len(t, data, 1)

	role := data[0].(map[string]interface{})
	assert.Equal(t, float64(1), role["id"])
	assert.Equal(t, "Super Admin", role["name"])

	mockRoleService.AssertExpectations(t)
}

func TestRoleHandler_GetAllRoles_EmptyResult(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles?search=nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().URL.RawQuery = "search=nonexistent"

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}
	expectedRoles := []entity.RoleEntity{}
	mockRoleService.On("GetAllRoles", mock.Anything, "nonexistent").Return(expectedRoles, nil)

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.GetAllRoles(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Roles retrieved successfully", response["message"])

	// When no roles are found, data should be null (empty slice marshals to null)
	assert.Nil(t, response["data"])

	mockRoleService.AssertExpectations(t)
}

func TestRoleHandler_GetAllRoles_ServiceError(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}
	mockRoleService.On("GetAllRoles", mock.Anything, "").Return(nil, assert.AnError)

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.GetAllRoles(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to retrieve roles", response["message"])
	assert.Nil(t, response["data"])

	mockRoleService.AssertExpectations(t)
}

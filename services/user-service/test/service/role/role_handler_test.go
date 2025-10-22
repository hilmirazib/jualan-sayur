package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"user-service/internal/core/domain/entity"
	"user-service/internal/adapter/handler"
	"user-service/test/service/mocks"
	validatorUtils "user-service/utils/validator"

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

func TestRoleHandler_GetRoleByID_Success(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/admin/roles/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}
	expectedRole := &entity.RoleEntity{
		ID:   1,
		Name: "Super Admin",
		Users: []entity.UserEntity{
			{ID: 1, Name: "Admin User", Email: "admin@example.com"},
			{ID: 2, Name: "Another Admin", Email: "admin2@example.com"},
		},
	}
	mockRoleService.On("GetRoleByID", mock.Anything, int64(1)).Return(expectedRole, nil)

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.GetRoleByID(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Role retrieved successfully", response["message"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["id"])
	assert.Equal(t, "Super Admin", data["name"])

	users := data["users"].([]interface{})
	assert.Len(t, users, 2)

	user1 := users[0].(map[string]interface{})
	assert.Equal(t, float64(1), user1["id"])
	assert.Equal(t, "Admin User", user1["name"])

	user2 := users[1].(map[string]interface{})
	assert.Equal(t, float64(2), user2["id"])
	assert.Equal(t, "Another Admin", user2["name"])

	mockRoleService.AssertExpectations(t)
}

func TestRoleHandler_GetRoleByID_InvalidID(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles/abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/admin/roles/:id")
	c.SetParamNames("id")
	c.SetParamValues("abc")

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.GetRoleByID(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid role ID format", response["message"])
	assert.Nil(t, response["data"])

	mockRoleService.AssertNotCalled(t, "GetRoleByID", mock.Anything, mock.Anything)
}

func TestRoleHandler_GetRoleByID_NotFound(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles/999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/admin/roles/:id")
	c.SetParamNames("id")
	c.SetParamValues("999")

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}
	mockRoleService.On("GetRoleByID", mock.Anything, int64(999)).Return(nil, errors.New("record not found"))

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.GetRoleByID(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Role not found", response["message"])
	assert.Nil(t, response["data"])

	mockRoleService.AssertExpectations(t)
}

func TestRoleHandler_GetRoleByID_ServiceError(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/admin/roles/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}
	mockRoleService.On("GetRoleByID", mock.Anything, int64(1)).Return(nil, assert.AnError)

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.GetRoleByID(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to retrieve role", response["message"])
	assert.Nil(t, response["data"])

	mockRoleService.AssertExpectations(t)
}

func TestRoleHandler_CreateRole_Success(t *testing.T) {
	// Setup Echo with validator
	e := echo.New()
	e.Validator = validatorUtils.NewValidator() // Add validator to Echo instance

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", strings.NewReader(`{"name":"Manager"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}
	expectedRole := &entity.RoleEntity{
		ID:   3,
		Name: "Manager",
	}
	mockRoleService.On("CreateRole", mock.Anything, "Manager").Return(expectedRole, nil)

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.CreateRole(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Role created successfully", response["message"])
	assert.Nil(t, response["data"])

	mockRoleService.AssertExpectations(t)
}

func TestRoleHandler_CreateRole_InvalidJSON(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", strings.NewReader(`invalid json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.CreateRole(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request format", response["message"])
	assert.Nil(t, response["data"])

	mockRoleService.AssertNotCalled(t, "CreateRole", mock.Anything, mock.Anything)
}

func TestRoleHandler_CreateRole_ValidationFailed(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", strings.NewReader(`{"name":""}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.CreateRole(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Name is required", response["message"])
	assert.Nil(t, response["data"])

	mockRoleService.AssertNotCalled(t, "CreateRole", mock.Anything, mock.Anything)
}

func TestRoleHandler_UpdateRole_Success(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/roles/1", strings.NewReader(`{"name":"Updated Admin"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/admin/roles/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}
	updatedRole := &entity.RoleEntity{
		ID:   1,
		Name: "Updated Admin",
	}
	mockRoleService.On("UpdateRole", mock.Anything, int64(1), "Updated Admin").Return(updatedRole, nil)

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.UpdateRole(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Role updated successfully", response["message"])
	assert.Nil(t, response["data"])

	mockRoleService.AssertExpectations(t)
}

func TestRoleHandler_UpdateRole_InvalidID(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/roles/abc", strings.NewReader(`{"name":"Updated Admin"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/admin/roles/:id")
	c.SetParamNames("id")
	c.SetParamValues("abc")

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.UpdateRole(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid role ID format", response["message"])
	assert.Nil(t, response["data"])

	mockRoleService.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleHandler_UpdateRole_InvalidJSON(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/roles/1", strings.NewReader(`invalid json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/admin/roles/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.UpdateRole(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request format", response["message"])
	assert.Nil(t, response["data"])

	mockRoleService.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleHandler_UpdateRole_NotFound(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/roles/999", strings.NewReader(`{"name":"Updated Admin"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/admin/roles/:id")
	c.SetParamNames("id")
	c.SetParamValues("999")

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}
	mockRoleService.On("UpdateRole", mock.Anything, int64(999), "Updated Admin").Return(nil, errors.New("role not found"))

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.UpdateRole(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Role not found", response["message"])
	assert.Nil(t, response["data"])

	mockRoleService.AssertExpectations(t)
}

func TestRoleHandler_UpdateRole_ValidationFailed(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/roles/1", strings.NewReader(`{"name":""}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/admin/roles/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.UpdateRole(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Name is required", response["message"])
	assert.Nil(t, response["data"])

	mockRoleService.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleHandler_UpdateRole_DuplicateName(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/roles/1", strings.NewReader(`{"name":"Customer"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/admin/roles/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}
	mockRoleService.On("UpdateRole", mock.Anything, int64(1), "Customer").Return(nil, errors.New("role with name 'Customer' already exists"))

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.UpdateRole(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "role with name 'Customer' already exists", response["message"])
	assert.Nil(t, response["data"])

	mockRoleService.AssertExpectations(t)
}

func TestRoleHandler_UpdateRole_ServiceError(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/roles/1", strings.NewReader(`{"name":"Updated Admin"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/admin/roles/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Setup mocks
	mockRoleService := &mocks.MockRoleService{}
	mockRoleService.On("UpdateRole", mock.Anything, int64(1), "Updated Admin").Return(nil, assert.AnError)

	// Test handler
	roleHandler := handler.NewRoleHandler(mockRoleService)
	err := roleHandler.UpdateRole(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to update role", response["message"])
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

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"user-service/internal/adapter/middleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSuperAdminMiddleware_Success(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set Super Admin role in context
	c.Set("user_role", "Super Admin")

	// Setup middleware
	middlewareFunc := middleware.SuperAdminMiddleware()
	handler := middlewareFunc(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	// Test middleware
	err := handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

func TestSuperAdminMiddleware_Forbidden_CustomerRole(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set Customer role in context (not Super Admin)
	c.Set("user_role", "Customer")

	// Setup middleware
	middlewareFunc := middleware.SuperAdminMiddleware()
	handler := middlewareFunc(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	// Test middleware
	err := handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	// Check response body
	expectedResponse := `{"message":"Access denied","data":null}`
	assert.JSONEq(t, expectedResponse, rec.Body.String())
}

func TestSuperAdminMiddleware_Forbidden_NoRole(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Don't set any role in context

	// Setup middleware
	middlewareFunc := middleware.SuperAdminMiddleware()
	handler := middlewareFunc(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	// Test middleware
	err := handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	// Check response body
	expectedResponse := `{"message":"Access denied","data":null}`
	assert.JSONEq(t, expectedResponse, rec.Body.String())
}

func TestSuperAdminMiddleware_Forbidden_InvalidRole(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set invalid role in context
	c.Set("user_role", "InvalidRole")

	// Setup middleware
	middlewareFunc := middleware.SuperAdminMiddleware()
	handler := middlewareFunc(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	// Test middleware
	err := handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	// Check response body
	expectedResponse := `{"message":"Access denied","data":null}`
	assert.JSONEq(t, expectedResponse, rec.Body.String())
}

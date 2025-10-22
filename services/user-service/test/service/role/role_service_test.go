package main

import (
	"context"
	"errors"
	"testing"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"
	"user-service/test/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRoleService_GetAllRoles_Success(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	expectedRoles := []entity.RoleEntity{
		{ID: 1, Name: "Super Admin"},
		{ID: 2, Name: "Customer"},
	}
	mockRoleRepo.On("GetAllRoles", mock.Anything, "").Return(expectedRoles, nil)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	roles, err := roleService.GetAllRoles(context.Background(), "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRoles, roles)
	assert.Len(t, roles, 2)
	mockRoleRepo.AssertExpectations(t)
}

func TestRoleService_GetAllRoles_WithSearch(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	searchTerm := "admin"
	expectedRoles := []entity.RoleEntity{
		{ID: 1, Name: "Super Admin"},
	}
	mockRoleRepo.On("GetAllRoles", mock.Anything, searchTerm).Return(expectedRoles, nil)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	roles, err := roleService.GetAllRoles(context.Background(), searchTerm)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRoles, roles)
	assert.Len(t, roles, 1)
	assert.Equal(t, "Super Admin", roles[0].Name)
	mockRoleRepo.AssertExpectations(t)
}

func TestRoleService_GetAllRoles_RepositoryError(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	expectedError := errors.New("database connection failed")
	mockRoleRepo.On("GetAllRoles", mock.Anything, "").Return(nil, expectedError)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	roles, err := roleService.GetAllRoles(context.Background(), "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, roles)
	assert.Equal(t, expectedError, err)
	mockRoleRepo.AssertExpectations(t)
}

func TestRoleService_GetAllRoles_EmptyResult(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	expectedRoles := []entity.RoleEntity{}
	mockRoleRepo.On("GetAllRoles", mock.Anything, "nonexistent").Return(expectedRoles, nil)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	roles, err := roleService.GetAllRoles(context.Background(), "nonexistent")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRoles, roles)
	assert.Len(t, roles, 0)
	mockRoleRepo.AssertExpectations(t)
}

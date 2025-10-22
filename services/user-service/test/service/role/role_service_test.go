package main

import (
	"context"
	"errors"
	"strings"
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

func TestRoleService_UpdateRole_Success(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	roleID := int64(1)
	newName := "Updated Admin"
	existingRole := &entity.RoleEntity{
		ID:   roleID,
		Name: "Super Admin",
	}

	// Mock get role by ID (role exists)
	mockRoleRepo.On("GetRoleByID", mock.Anything, roleID).Return(existingRole, nil)

	// Mock existing roles check (no duplicates)
	mockRoleRepo.On("GetAllRoles", mock.Anything, "").Return([]entity.RoleEntity{
		{ID: 1, Name: "Super Admin"},
		{ID: 2, Name: "Customer"},
	}, nil)

	// Mock update role
	updatedRole := &entity.RoleEntity{
		ID:   roleID,
		Name: newName,
	}
	mockRoleRepo.On("UpdateRole", mock.Anything, roleID, mock.AnythingOfType("*entity.RoleEntity")).Return(updatedRole, nil)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.UpdateRole(context.Background(), roleID, newName)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, updatedRole, role)
	assert.Equal(t, newName, role.Name)
	mockRoleRepo.AssertExpectations(t)
}

func TestRoleService_UpdateRole_NotFound(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	roleID := int64(999)
	newName := "Updated Admin"

	// Mock get role by ID (role not found)
	mockRoleRepo.On("GetRoleByID", mock.Anything, roleID).Return(nil, errors.New("record not found"))

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.UpdateRole(context.Background(), roleID, newName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, "role not found", err.Error())
	mockRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertNotCalled(t, "GetAllRoles", mock.Anything, mock.Anything)
	mockRoleRepo.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleService_UpdateRole_EmptyName(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.UpdateRole(context.Background(), 1, "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, "role name cannot be empty", err.Error())
	mockRoleRepo.AssertNotCalled(t, "GetRoleByID", mock.Anything, mock.Anything)
	mockRoleRepo.AssertNotCalled(t, "GetAllRoles", mock.Anything, mock.Anything)
	mockRoleRepo.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleService_UpdateRole_WhitespaceName(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.UpdateRole(context.Background(), 1, "   ")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, "role name cannot be empty", err.Error())
	mockRoleRepo.AssertNotCalled(t, "GetRoleByID", mock.Anything, mock.Anything)
	mockRoleRepo.AssertNotCalled(t, "GetAllRoles", mock.Anything, mock.Anything)
	mockRoleRepo.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleService_UpdateRole_NameTooShort(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.UpdateRole(context.Background(), 1, "A")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, "role name must be between 2 and 50 characters", err.Error())
	mockRoleRepo.AssertNotCalled(t, "GetRoleByID", mock.Anything, mock.Anything)
	mockRoleRepo.AssertNotCalled(t, "GetAllRoles", mock.Anything, mock.Anything)
	mockRoleRepo.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleService_UpdateRole_NameTooLong(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	longName := strings.Repeat("A", 51)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.UpdateRole(context.Background(), 1, longName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, "role name must be between 2 and 50 characters", err.Error())
	mockRoleRepo.AssertNotCalled(t, "GetRoleByID", mock.Anything, mock.Anything)
	mockRoleRepo.AssertNotCalled(t, "GetAllRoles", mock.Anything, mock.Anything)
	mockRoleRepo.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleService_UpdateRole_DuplicateName(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	roleID := int64(1)
	newName := "Customer"
	existingRole := &entity.RoleEntity{
		ID:   roleID,
		Name: "Super Admin",
	}

	// Mock get role by ID (role exists)
	mockRoleRepo.On("GetRoleByID", mock.Anything, roleID).Return(existingRole, nil)

	// Mock existing roles check (contains duplicate)
	mockRoleRepo.On("GetAllRoles", mock.Anything, "").Return([]entity.RoleEntity{
		{ID: 1, Name: "Super Admin"},
		{ID: 2, Name: "Customer"},
	}, nil)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.UpdateRole(context.Background(), roleID, newName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, "role with name 'Customer' already exists", err.Error())
	mockRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleService_UpdateRole_GetRoleByIDError(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	roleID := int64(1)
	newName := "Updated Admin"
	expectedError := errors.New("database connection failed")

	// Mock get role by ID error
	mockRoleRepo.On("GetRoleByID", mock.Anything, roleID).Return(nil, expectedError)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.UpdateRole(context.Background(), roleID, newName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, expectedError, err)
	mockRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertNotCalled(t, "GetAllRoles", mock.Anything, mock.Anything)
	mockRoleRepo.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleService_UpdateRole_CheckExistingRolesError(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	roleID := int64(1)
	newName := "Updated Admin"
	existingRole := &entity.RoleEntity{
		ID:   roleID,
		Name: "Super Admin",
	}
	expectedError := errors.New("database connection failed")

	// Mock get role by ID (role exists)
	mockRoleRepo.On("GetRoleByID", mock.Anything, roleID).Return(existingRole, nil)

	// Mock existing roles check error
	mockRoleRepo.On("GetAllRoles", mock.Anything, "").Return(nil, expectedError)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.UpdateRole(context.Background(), roleID, newName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, expectedError, err)
	mockRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleService_UpdateRole_RepositoryError(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	roleID := int64(1)
	newName := "Updated Admin"
	existingRole := &entity.RoleEntity{
		ID:   roleID,
		Name: "Super Admin",
	}
	expectedError := errors.New("database connection failed")

	// Mock get role by ID (role exists)
	mockRoleRepo.On("GetRoleByID", mock.Anything, roleID).Return(existingRole, nil)

	// Mock existing roles check (no duplicates)
	mockRoleRepo.On("GetAllRoles", mock.Anything, "").Return([]entity.RoleEntity{
		{ID: 1, Name: "Super Admin"},
		{ID: 2, Name: "Customer"},
	}, nil)

	// Mock update role error
	mockRoleRepo.On("UpdateRole", mock.Anything, roleID, mock.AnythingOfType("*entity.RoleEntity")).Return(nil, expectedError)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.UpdateRole(context.Background(), roleID, newName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
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

func TestRoleService_GetRoleByID_Success(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	expectedRole := &entity.RoleEntity{
		ID:   1,
		Name: "Super Admin",
		Users: []entity.UserEntity{
			{ID: 1, Name: "Admin User", Email: "admin@example.com"},
		},
	}
	mockRoleRepo.On("GetRoleByID", mock.Anything, int64(1)).Return(expectedRole, nil)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.GetRoleByID(context.Background(), 1)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRole, role)
	assert.Equal(t, "Super Admin", role.Name)
	assert.Len(t, role.Users, 1)
	mockRoleRepo.AssertExpectations(t)
}

func TestRoleService_GetRoleByID_NotFound(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	mockRoleRepo.On("GetRoleByID", mock.Anything, int64(999)).Return(nil, errors.New("record not found"))

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.GetRoleByID(context.Background(), 999)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, "record not found", err.Error())
	mockRoleRepo.AssertExpectations(t)
}

func TestRoleService_GetRoleByID_RepositoryError(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	expectedError := errors.New("database connection failed")
	mockRoleRepo.On("GetRoleByID", mock.Anything, int64(1)).Return(nil, expectedError)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.GetRoleByID(context.Background(), 1)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, expectedError, err)
	mockRoleRepo.AssertExpectations(t)
}

func TestRoleService_CreateRole_Success(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	roleName := "Manager"
	expectedRole := &entity.RoleEntity{
		ID:   3,
		Name: roleName,
	}

	// Mock existing roles (no duplicates)
	mockRoleRepo.On("GetAllRoles", mock.Anything, "").Return([]entity.RoleEntity{
		{ID: 1, Name: "Super Admin"},
		{ID: 2, Name: "Customer"},
	}, nil)

	// Mock create role
	mockRoleRepo.On("CreateRole", mock.Anything, mock.AnythingOfType("*entity.RoleEntity")).Return(expectedRole, nil)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.CreateRole(context.Background(), roleName)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRole, role)
	assert.Equal(t, roleName, role.Name)
	mockRoleRepo.AssertExpectations(t)
}

func TestRoleService_CreateRole_EmptyName(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.CreateRole(context.Background(), "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, "role name cannot be empty", err.Error())
	mockRoleRepo.AssertNotCalled(t, "GetAllRoles", mock.Anything, mock.Anything)
	mockRoleRepo.AssertNotCalled(t, "CreateRole", mock.Anything, mock.Anything)
}

func TestRoleService_CreateRole_WhitespaceName(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.CreateRole(context.Background(), "   ")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, "role name cannot be empty", err.Error())
	mockRoleRepo.AssertNotCalled(t, "GetAllRoles", mock.Anything, mock.Anything)
	mockRoleRepo.AssertNotCalled(t, "CreateRole", mock.Anything, mock.Anything)
}

func TestRoleService_CreateRole_NameTooShort(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.CreateRole(context.Background(), "A")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, "role name must be between 2 and 50 characters", err.Error())
	mockRoleRepo.AssertNotCalled(t, "GetAllRoles", mock.Anything, mock.Anything)
	mockRoleRepo.AssertNotCalled(t, "CreateRole", mock.Anything, mock.Anything)
}

func TestRoleService_CreateRole_NameTooLong(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	longName := strings.Repeat("A", 51)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.CreateRole(context.Background(), longName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, "role name must be between 2 and 50 characters", err.Error())
	mockRoleRepo.AssertNotCalled(t, "GetAllRoles", mock.Anything, mock.Anything)
	mockRoleRepo.AssertNotCalled(t, "CreateRole", mock.Anything, mock.Anything)
}

func TestRoleService_CreateRole_DuplicateName(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	roleName := "Super Admin"

	// Mock existing roles (contains duplicate)
	mockRoleRepo.On("GetAllRoles", mock.Anything, "").Return([]entity.RoleEntity{
		{ID: 1, Name: "Super Admin"},
		{ID: 2, Name: "Customer"},
	}, nil)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.CreateRole(context.Background(), roleName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, "role with name 'Super Admin' already exists", err.Error())
	mockRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertNotCalled(t, "CreateRole", mock.Anything, mock.Anything)
}

func TestRoleService_CreateRole_CheckExistingRolesError(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	expectedError := errors.New("database connection failed")

	mockRoleRepo.On("GetAllRoles", mock.Anything, "").Return(nil, expectedError)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.CreateRole(context.Background(), "Manager")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, expectedError, err)
	mockRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertNotCalled(t, "CreateRole", mock.Anything, mock.Anything)
}

func TestRoleService_CreateRole_RepositoryError(t *testing.T) {
	// Setup
	mockRoleRepo := &mocks.MockRoleRepository{}
	expectedError := errors.New("database connection failed")

	// Mock existing roles (no duplicates)
	mockRoleRepo.On("GetAllRoles", mock.Anything, "").Return([]entity.RoleEntity{
		{ID: 1, Name: "Super Admin"},
		{ID: 2, Name: "Customer"},
	}, nil)

	// Mock create role error
	mockRoleRepo.On("CreateRole", mock.Anything, mock.AnythingOfType("*entity.RoleEntity")).Return(nil, expectedError)

	// Test service
	roleService := service.NewRoleService(mockRoleRepo)
	role, err := roleService.CreateRole(context.Background(), "Manager")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, expectedError, err)
	mockRoleRepo.AssertExpectations(t)
}

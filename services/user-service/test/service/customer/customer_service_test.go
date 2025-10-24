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
	"gorm.io/gorm"
)

func TestAuthService_GetCustomers_Success(t *testing.T) {
	// Setup
	mockUserRepo := &mocks.MockUserRepository{}
	expectedCustomers := []entity.UserEntity{
		{
			ID:     1,
			Name:   "John Customer",
			Email:  "john@example.com",
			Phone:  "+628987654321",
			Photo:  "https://example.com/photo.jpg",
			Address: "Jakarta",
			Lat:    -6.2088,
			Lng:    106.8456,
			IsVerified: true,
		},
		{
			ID:     2,
			Name:   "Jane Customer",
			Email:  "jane@example.com",
			Phone:  "+628123456789",
			Photo:  "https://example.com/photo2.jpg",
			Address: "Bandung",
			Lat:    -6.9175,
			Lng:    107.6191,
			IsVerified: true,
		},
	}
	expectedTotalCount := int64(2)

	mockUserRepo.On("GetCustomers", mock.Anything, "", 1, 10, "").Return(expectedCustomers, expectedTotalCount, nil)

	// Test service
	authService := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)
	customers, pagination, err := authService.GetCustomers(context.Background(), "", 1, 10, "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedCustomers, customers)
	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, expectedTotalCount, pagination.TotalCount)
	assert.Equal(t, 10, pagination.PerPage)
	assert.Equal(t, 1, pagination.TotalPage)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_GetCustomers_WithSearch(t *testing.T) {
	// Setup
	mockUserRepo := &mocks.MockUserRepository{}
	searchTerm := "john"
	expectedCustomers := []entity.UserEntity{
		{
			ID:     1,
			Name:   "John Customer",
			Email:  "john@example.com",
			Phone:  "+628987654321",
			Photo:  "https://example.com/photo.jpg",
			Address: "Jakarta",
			Lat:    -6.2088,
			Lng:    106.8456,
			IsVerified: true,
		},
	}
	expectedTotalCount := int64(1)

	mockUserRepo.On("GetCustomers", mock.Anything, searchTerm, 1, 10, "").Return(expectedCustomers, expectedTotalCount, nil)

	// Test service
	authService := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)
	customers, pagination, err := authService.GetCustomers(context.Background(), searchTerm, 1, 10, "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedCustomers, customers)
	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, expectedTotalCount, pagination.TotalCount)
	assert.Equal(t, 10, pagination.PerPage)
	assert.Equal(t, 1, pagination.TotalPage)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_GetCustomers_WithPagination(t *testing.T) {
	// Setup
	mockUserRepo := &mocks.MockUserRepository{}
	expectedCustomers := []entity.UserEntity{
		{
			ID:     3,
			Name:   "Bob Customer",
			Email:  "bob@example.com",
			Phone:  "+628111111111",
			Photo:  "https://example.com/photo3.jpg",
			Address: "Surabaya",
			Lat:    -7.2575,
			Lng:    112.7521,
			IsVerified: true,
		},
	}
	expectedTotalCount := int64(25)
	page := 2
	limit := 5

	mockUserRepo.On("GetCustomers", mock.Anything, "", page, limit, "").Return(expectedCustomers, expectedTotalCount, nil)

	// Test service
	authService := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)
	customers, pagination, err := authService.GetCustomers(context.Background(), "", page, limit, "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedCustomers, customers)
	assert.Equal(t, page, pagination.Page)
	assert.Equal(t, expectedTotalCount, pagination.TotalCount)
	assert.Equal(t, limit, pagination.PerPage)
	assert.Equal(t, 5, pagination.TotalPage) // 25 total / 5 per page = 5 pages
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_GetCustomers_RepositoryError(t *testing.T) {
	// Setup
	mockUserRepo := &mocks.MockUserRepository{}
	expectedError := errors.New("database connection failed")
	mockUserRepo.On("GetCustomers", mock.Anything, "", 1, 10, "").Return(nil, int64(0), expectedError)

	// Test service
	authService := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)
	customers, pagination, err := authService.GetCustomers(context.Background(), "", 1, 10, "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, customers)
	assert.Nil(t, pagination)
	assert.Equal(t, "failed to retrieve customers", err.Error())
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_GetCustomers_EmptyResult(t *testing.T) {
	// Setup
	mockUserRepo := &mocks.MockUserRepository{}
	expectedCustomers := []entity.UserEntity{}
	expectedTotalCount := int64(0)

	mockUserRepo.On("GetCustomers", mock.Anything, "nonexistent", 1, 10, "").Return(expectedCustomers, expectedTotalCount, nil)

	// Test service
	authService := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)
	customers, pagination, err := authService.GetCustomers(context.Background(), "nonexistent", 1, 10, "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedCustomers, customers)
	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, expectedTotalCount, pagination.TotalCount)
	assert.Equal(t, 10, pagination.PerPage)
	assert.Equal(t, 0, pagination.TotalPage)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_GetCustomerByID_Success(t *testing.T) {
	// Setup
	mockUserRepo := &mocks.MockUserRepository{}
	expectedCustomer := &entity.UserEntity{
		ID:       1,
		Name:     "John Customer",
		Email:    "john@example.com",
		Phone:    "+628987654321",
		Photo:    "https://example.com/photo.jpg",
		Address:  "Jakarta",
		Lat:      -6.2088,
		Lng:      106.8456,
		RoleName: "Customer",
		RoleID:   2,
		IsVerified: true,
	}
	customerID := int64(1)

	mockUserRepo.On("GetCustomerByID", mock.Anything, customerID).Return(expectedCustomer, nil)

	// Test service
	authService := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)
	customer, err := authService.GetCustomerByID(context.Background(), customerID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedCustomer, customer)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_GetCustomerByID_NotFound(t *testing.T) {
	// Setup
	mockUserRepo := &mocks.MockUserRepository{}
	customerID := int64(999)
	expectedError := errors.New("customer not found")

	mockUserRepo.On("GetCustomerByID", mock.Anything, customerID).Return(nil, gorm.ErrRecordNotFound)

	// Test service
	authService := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)
	customer, err := authService.GetCustomerByID(context.Background(), customerID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, customer)
	assert.Equal(t, expectedError.Error(), err.Error())
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_GetCustomerByID_RepositoryError(t *testing.T) {
	// Setup
	mockUserRepo := &mocks.MockUserRepository{}
	customerID := int64(1)
	expectedError := errors.New("database connection failed")

	mockUserRepo.On("GetCustomerByID", mock.Anything, customerID).Return(nil, expectedError)

	// Test service
	authService := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, nil)
	customer, err := authService.GetCustomerByID(context.Background(), customerID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, customer)
	assert.Equal(t, expectedError, err)
	mockUserRepo.AssertExpectations(t)
}

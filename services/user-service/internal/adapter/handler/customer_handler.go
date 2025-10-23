package handler

import (
	"net/http"
	"strconv"
	"user-service/internal/core/port"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type CustomerHandlerInterface interface {
	GetCustomers(c echo.Context) error
}

type CustomerHandler struct {
	userService port.UserServiceInterface
}

func (h *CustomerHandler) GetCustomers(c echo.Context) error {
	// Get query parameters
	search := c.QueryParam("search")
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")
	orderBy := c.QueryParam("orderBy")

	// Parse page
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Parse limit
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Get customers from service
	customers, pagination, err := h.userService.GetCustomers(c.Request().Context(), search, page, limit, orderBy)
	if err != nil {
		log.Error().Err(err).Str("search", search).Int("page", page).Int("limit", limit).Msg("[CustomerHandler-GetCustomers] Failed to get customers")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to retrieve customers",
			"data":    nil,
		})
	}

	// Transform customers to response format
	var customerData []map[string]interface{}
	for _, customer := range customers {
		customerData = append(customerData, map[string]interface{}{
			"id":     customer.ID,
			"name":   customer.Name,
			"photo":  customer.Photo,
			"email":  customer.Email,
			"phone":  customer.Phone,
		})
	}

	log.Info().Int("count", len(customers)).Int64("total_count", pagination.TotalCount).Str("search", search).Int("page", page).Int("limit", limit).Msg("[CustomerHandler-GetCustomers] Customers retrieved successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Customers retrieved successfully",
		"data":    customerData,
		"pagination": map[string]interface{}{
			"page":        pagination.Page,
			"total_count": pagination.TotalCount,
			"per_page":    pagination.PerPage,
			"total_page":  pagination.TotalPage,
		},
	})
}

func NewCustomerHandler(userService port.UserServiceInterface) CustomerHandlerInterface {
	return &CustomerHandler{
		userService: userService,
	}
}

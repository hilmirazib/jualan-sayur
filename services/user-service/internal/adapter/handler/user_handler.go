package handler

import (
	"net/http"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/port"
	"user-service/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type UserHandlerInterface interface {
	SignIn(ctx echo.Context) error
	CreateUserAccount(ctx echo.Context) error
	AdminCheck(ctx echo.Context) error
}

type UserHandler struct {
	userService port.UserServiceInterface
	validator   *validator.Validator
}

func (u *UserHandler) SignIn(c echo.Context) error {
	var (
		req        = request.SignInRequest{}
		resp       = response.DefaultResponse{}
		respSignIn = response.SignInResponse{}
		ctx        = c.Request().Context()
	)

	// Bind request
	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("[UserHandler-SignIn] Failed to bind request")
		resp.Message = "Invalid request format"
		return c.JSON(http.StatusBadRequest, resp)
	}

	// Validate request using go-playground/validator
	if err := u.validator.Validate(&req); err != nil {
		log.Error().Err(err).Msg("[UserHandler-SignIn] Validation failed")
		resp.Message = err.Error()
		return c.JSON(http.StatusBadRequest, resp)
	}

	// Convert request to entity
	userEntity := entity.UserEntity{
		Email:    req.Email,
		Password: req.Password,
	}

	// Call service
	user, token, err := u.userService.SignIn(ctx, userEntity)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("[UserHandler-SignIn] Sign in failed")

		// Handle different error types
		switch err.Error() {
		case "user not found":
			resp.Message = "User not found"
			return c.JSON(http.StatusNotFound, resp)
		case "incorrect password":
			resp.Message = "Incorrect password"
			return c.JSON(http.StatusUnauthorized, resp)
		case "failed to generate token":
			resp.Message = "Authentication failed"
			return c.JSON(http.StatusInternalServerError, resp)
		default:
			resp.Message = "Internal server error"
			return c.JSON(http.StatusInternalServerError, resp)
		}
	}

	// Build response
	respSignIn.AccessToken = token
	respSignIn.Role = user.RoleName
	respSignIn.ID = user.ID
	respSignIn.Name = user.Name
	respSignIn.Email = user.Email
	respSignIn.Phone = user.Phone
	respSignIn.Lat = user.Lat
	respSignIn.Lng = user.Lng

	resp.Message = "Sign in successful"
	resp.Data = respSignIn

	log.Info().Str("email", req.Email).Int64("user_id", user.ID).Msg("[UserHandler-SignIn] User signed in successfully")

	return c.JSON(http.StatusOK, resp)
}

// AdminCheck handles admin authentication check
func (u *UserHandler) AdminCheck(c echo.Context) error {
	// Get user information from context (set by middleware)
	userID := c.Get("user_id").(int64)
	email := c.Get("user_email").(string)
	role := c.Get("user_role").(string)
	sessionID := c.Get("session_id").(string)

	// Build response
	resp := map[string]interface{}{
		"message": "Authentication successful",
		"data": map[string]interface{}{
			"user_id":    userID,
			"email":      email,
			"role":       role,
			"session_id": sessionID,
		},
	}

	log.Info().
		Int64("user_id", userID).
		Str("email", email).
		Str("role", role).
		Str("session_id", sessionID).
		Msg("[UserHandler-AdminCheck] Admin authentication check successful")

	return c.JSON(http.StatusOK, resp)
}

// CreateUserAccount handles user account creation
func (u *UserHandler) CreateUserAccount(c echo.Context) error {
	var (
		req         = request.CreateUserAccountRequest{}
		resp        = response.DefaultResponse{}
		respCreate  = response.CreateUserAccountResponse{}
		ctx         = c.Request().Context()
	)

	// Bind request
	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("[UserHandler-CreateUserAccount] Failed to bind request")
		resp.Message = "Invalid request format"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	// Validate request using go-playground/validator
	if err := u.validator.Validate(&req); err != nil {
		log.Error().Err(err).Msg("[UserHandler-CreateUserAccount] Validation failed")
		resp.Message = "Validation failed"
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	// Call service
	err := u.userService.CreateUserAccount(ctx, req.Email, req.Name, req.Password, req.PasswordConfirmation)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("[UserHandler-CreateUserAccount] Account creation failed")

		// Handle different error types
		switch err.Error() {
		case "invalid email format":
			resp.Message = "Invalid email format"
			resp.Data = nil
			return c.JSON(http.StatusUnprocessableEntity, resp)
		case "password is required", "password must be at least 8 characters long", "password confirmation does not match":
			resp.Message = err.Error()
			resp.Data = nil
			return c.JSON(http.StatusUnprocessableEntity, resp)
		case "email already exists":
			resp.Message = "Email already exists"
			resp.Data = nil
			return c.JSON(http.StatusConflict, resp)
		case "failed to create account", "failed to generate verification token", "failed to create verification token":
			resp.Message = "Failed to create account"
			resp.Data = nil
			return c.JSON(http.StatusInternalServerError, resp)
		default:
			resp.Message = "Internal server error"
			resp.Data = nil
			return c.JSON(http.StatusInternalServerError, resp)
		}
	}

	// Build response
	respCreate.Name = req.Name
	respCreate.Email = req.Email

	resp.Message = "Account created successfully. Please check your email for verification."
	resp.Data = respCreate

	log.Info().Str("email", req.Email).Msg("[UserHandler-CreateUserAccount] User account created successfully")

	return c.JSON(http.StatusCreated, resp)
}

func NewUserHandler(userService port.UserServiceInterface) UserHandlerInterface {
	return &UserHandler{
		userService: userService,
		validator:   validator.NewValidator(),
	}
}

package handler

import (
	"net/http"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/port"

	"github.com/go-playground/validator/v10"
	myvalidator "user-service/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type UserHandlerInterface interface {
	SignIn(ctx echo.Context) error
	CreateUserAccount(ctx echo.Context) error
	VerifyUserAccount(ctx echo.Context) error
	ForgotPassword(ctx echo.Context) error
	ResetPassword(ctx echo.Context) error
	AdminCheck(ctx echo.Context) error
}

type UserHandler struct {
	userService port.UserServiceInterface
	validator   *myvalidator.Validator
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

		// Handle specific validation errors with clear messages
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				fieldName := fieldError.Field()
				tag := fieldError.Tag()

				switch fieldName {
				case "Email":
					if tag == "email" {
						resp.Message = "Invalid email format"
						resp.Data = nil
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
					if tag == "required" {
						resp.Message = "Email is required"
						resp.Data = nil
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
				case "Name":
					if tag == "required" {
						resp.Message = "Name is required"
						resp.Data = nil
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
					if tag == "min" {
						resp.Message = "Name must be at least 2 characters long"
						resp.Data = nil
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
					if tag == "max" {
						resp.Message = "Name must not exceed 100 characters"
						resp.Data = nil
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
				case "Password":
					if tag == "required" {
						resp.Message = "Password is required"
						resp.Data = nil
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
					if tag == "min" {
						resp.Message = "Password must be at least 8 characters long"
						resp.Data = nil
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
				case "PasswordConfirmation":
					if tag == "required" {
						resp.Message = "Password confirmation is required"
						resp.Data = nil
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
					if tag == "eqfield" {
						resp.Message = "Password confirmation does not match"
						resp.Data = nil
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
				}
			}
		}

		// Fallback for other validation errors
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

// VerifyUserAccount handles email verification
func (u *UserHandler) VerifyUserAccount(c echo.Context) error {
	var (
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
	)

	// Get token from query parameter
	token := c.QueryParam("token")
	if token == "" {
		log.Warn().Msg("[UserHandler-VerifyUserAccount] Missing verification token")
		resp.Message = "Verification token is required"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	// Call service
	err := u.userService.VerifyUserAccount(ctx, token)
	if err != nil {
		log.Error().Err(err).Str("token", token).Msg("[UserHandler-VerifyUserAccount] Account verification failed")

		// Handle different error types
		switch err.Error() {
		case "invalid or expired verification token":
			resp.Message = "Invalid or expired verification token"
			resp.Data = nil
			return c.JSON(http.StatusBadRequest, resp)
		case "failed to verify token", "failed to verify account":
			resp.Message = "Failed to verify account"
			resp.Data = nil
			return c.JSON(http.StatusInternalServerError, resp)
		default:
			resp.Message = "Internal server error"
			resp.Data = nil
			return c.JSON(http.StatusInternalServerError, resp)
		}
	}

	resp.Message = "Account verified successfully. You can now sign in."
	resp.Data = nil

	log.Info().Str("token", token).Msg("[UserHandler-VerifyUserAccount] User account verified successfully")

	return c.JSON(http.StatusOK, resp)
}

// ForgotPassword handles password reset request
func (u *UserHandler) ForgotPassword(c echo.Context) error {
	var (
		req  = request.ForgotPasswordRequest{}
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
	)

	// Bind request
	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("[UserHandler-ForgotPassword] Failed to bind request")
		resp.Message = "Invalid request format"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	// Validate request using go-playground/validator
	if err := u.validator.Validate(&req); err != nil {
		log.Error().Err(err).Msg("[UserHandler-ForgotPassword] Validation failed")
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	// Call service
	err := u.userService.ForgotPassword(ctx, req.Email)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("[UserHandler-ForgotPassword] Password reset request failed")

		// Handle different error types
		switch err.Error() {
		case "invalid email format":
			resp.Message = "Invalid email format"
			resp.Data = nil
			return c.JSON(http.StatusUnprocessableEntity, resp)
		case "failed to process request", "failed to generate reset token", "failed to create reset token":
			resp.Message = "Failed to process request"
			resp.Data = nil
			return c.JSON(http.StatusInternalServerError, resp)
		default:
			resp.Message = "Internal server error"
			resp.Data = nil
			return c.JSON(http.StatusInternalServerError, resp)
		}
	}

	resp.Message = "If an account with this email exists, you will receive a password reset link."
	resp.Data = nil

	log.Info().Str("email", req.Email).Msg("[UserHandler-ForgotPassword] Password reset request processed successfully")

	return c.JSON(http.StatusOK, resp)
}

// ResetPassword handles password reset with token
func (u *UserHandler) ResetPassword(c echo.Context) error {
	var (
		req  = request.ResetPasswordRequest{}
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
	)

	// Bind request
	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("[UserHandler-ResetPassword] Failed to bind request")
		resp.Message = "Invalid request format"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	// Validate request using go-playground/validator
	if err := u.validator.Validate(&req); err != nil {
		log.Error().Err(err).Msg("[UserHandler-ResetPassword] Validation failed")

		// Handle specific validation errors with clear messages
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				fieldName := fieldError.Field()
				tag := fieldError.Tag()

				switch fieldName {
				case "Token":
					if tag == "required" {
						resp.Message = "Reset token is required"
						resp.Data = nil
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
				case "Password":
					if tag == "required" {
						resp.Message = "Password is required"
						resp.Data = nil
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
					if tag == "min" {
						resp.Message = "Password must be at least 8 characters long"
						resp.Data = nil
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
				case "PasswordConfirmation":
					if tag == "required" {
						resp.Message = "Password confirmation is required"
						resp.Data = nil
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
					if tag == "eqfield" {
						resp.Message = "Password confirmation does not match"
						resp.Data = nil
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
				}
			}
		}

		// Fallback for other validation errors
		resp.Message = "Validation failed"
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	// Call service
	err := u.userService.ResetPassword(ctx, req.Token, req.Password, req.PasswordConfirmation)
	if err != nil {
		log.Error().Err(err).Str("token", req.Token).Msg("[UserHandler-ResetPassword] Password reset failed")

		// Handle different error types
		switch err.Error() {
		case "invalid or expired reset token":
			resp.Message = "Invalid or expired reset token"
			resp.Data = nil
			return c.JSON(http.StatusBadRequest, resp)
		case "invalid token type":
			resp.Message = "Invalid token type"
			resp.Data = nil
			return c.JSON(http.StatusBadRequest, resp)
		case "password is required", "password must be at least 8 characters long", "password confirmation does not match":
			resp.Message = err.Error()
			resp.Data = nil
			return c.JSON(http.StatusUnprocessableEntity, resp)
		case "failed to validate token", "failed to process password", "failed to update password":
			resp.Message = "Failed to reset password"
			resp.Data = nil
			return c.JSON(http.StatusInternalServerError, resp)
		default:
			resp.Message = "Internal server error"
			resp.Data = nil
			return c.JSON(http.StatusInternalServerError, resp)
		}
	}

	resp.Message = "Password reset successfully. You can now sign in with your new password."
	resp.Data = nil

	log.Info().Str("token", req.Token).Msg("[UserHandler-ResetPassword] Password reset successfully")

	return c.JSON(http.StatusOK, resp)
}

func NewUserHandler(userService port.UserServiceInterface) UserHandlerInterface {
	return &UserHandler{
		userService: userService,
		validator:   myvalidator.NewValidator(),
	}
}

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

type AuthHandlerInterface interface {
	SignIn(ctx echo.Context) error
	CreateUserAccount(ctx echo.Context) error
	VerifyUserAccount(ctx echo.Context) error
	ForgotPassword(ctx echo.Context) error
	ResetPassword(ctx echo.Context) error
}

type AuthHandler struct {
	userService port.UserServiceInterface
	validator   *myvalidator.Validator
}

func (a *AuthHandler) SignIn(c echo.Context) error {
	var (
		req        = request.SignInRequest{}
		resp       = response.DefaultResponse{}
		respSignIn = response.SignInResponse{}
		ctx        = c.Request().Context()
	)

	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("[AuthHandler-SignIn] Failed to bind request")
		resp.Message = "Invalid request format"
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := a.validator.Validate(&req); err != nil {
		log.Error().Err(err).Msg("[AuthHandler-SignIn] Validation failed")
		resp.Message = err.Error()
		return c.JSON(http.StatusBadRequest, resp)
	}

	userEntity := entity.UserEntity{
		Email:    req.Email,
		Password: req.Password,
	}

	user, token, err := a.userService.SignIn(ctx, userEntity)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("[AuthHandler-SignIn] Sign in failed")

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

	log.Info().Str("email", req.Email).Int64("user_id", user.ID).Msg("[AuthHandler-SignIn] User signed in successfully")

	return c.JSON(http.StatusOK, resp)
}

func (a *AuthHandler) CreateUserAccount(c echo.Context) error {
	var (
		req         = request.CreateUserAccountRequest{}
		resp        = response.DefaultResponse{}
		respCreate  = response.CreateUserAccountResponse{}
		ctx         = c.Request().Context()
	)

	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("[AuthHandler-CreateUserAccount] Failed to bind request")
		resp.Message = "Invalid request format"
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := a.validator.Validate(&req); err != nil {
		log.Error().Err(err).Msg("[AuthHandler-CreateUserAccount] Validation failed")

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				fieldName := fieldError.Field()
				tag := fieldError.Tag()

				switch fieldName {
				case "Email":
					if tag == "email" {
						resp.Message = "Invalid email format"
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
					if tag == "required" {
						resp.Message = "Email is required"
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
				case "Name":
					if tag == "required" {
						resp.Message = "Name is required"
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
					if tag == "min" {
						resp.Message = "Name must be at least 2 characters long"
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
					if tag == "max" {
						resp.Message = "Name must not exceed 100 characters"
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
				case "Password":
					if tag == "required" {
						resp.Message = "Password is required"
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
					if tag == "min" {
						resp.Message = "Password must be at least 8 characters long"
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
				case "PasswordConfirmation":
					if tag == "required" {
						resp.Message = "Password confirmation is required"
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
					if tag == "eqfield" {
						resp.Message = "Password confirmation does not match"
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
				}
			}
		}

		resp.Message = "Validation failed"
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	err := a.userService.CreateUserAccount(ctx, req.Email, req.Name, req.Password, req.PasswordConfirmation)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("[AuthHandler-CreateUserAccount] Account creation failed")

		switch err.Error() {
		case "invalid email format":
			resp.Message = "Invalid email format"
			return c.JSON(http.StatusUnprocessableEntity, resp)
		case "password is required", "password must be at least 8 characters long", "password confirmation does not match":
			resp.Message = err.Error()
			return c.JSON(http.StatusUnprocessableEntity, resp)
		case "email already exists":
			resp.Message = "Email already exists"
			return c.JSON(http.StatusConflict, resp)
		case "failed to create account", "failed to generate verification token", "failed to create verification token":
			resp.Message = "Failed to create account"
			return c.JSON(http.StatusInternalServerError, resp)
		default:
			resp.Message = "Internal server error"
			return c.JSON(http.StatusInternalServerError, resp)
		}
	}

	respCreate.Name = req.Name
	respCreate.Email = req.Email
	resp.Message = "Account created successfully. Please check your email for verification."
	resp.Data = respCreate

	log.Info().Str("email", req.Email).Msg("[AuthHandler-CreateUserAccount] User account created successfully")

	return c.JSON(http.StatusCreated, resp)
}

func (a *AuthHandler) VerifyUserAccount(c echo.Context) error {
	var (
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
	)

	token := c.QueryParam("token")
	if token == "" {
		log.Warn().Msg("[AuthHandler-VerifyUserAccount] Missing verification token")
		resp.Message = "Verification token is required"
		return c.JSON(http.StatusBadRequest, resp)
	}

	err := a.userService.VerifyUserAccount(ctx, token)
	if err != nil {
		log.Error().Err(err).Str("token", token).Msg("[AuthHandler-VerifyUserAccount] Account verification failed")

		switch err.Error() {
		case "invalid or expired verification token":
			resp.Message = "Invalid or expired verification token"
			return c.JSON(http.StatusBadRequest, resp)
		case "failed to verify token", "failed to verify account":
			resp.Message = "Failed to verify account"
			return c.JSON(http.StatusInternalServerError, resp)
		default:
			resp.Message = "Internal server error"
			return c.JSON(http.StatusInternalServerError, resp)
		}
	}

	resp.Message = "Account verified successfully. You can now sign in."
	log.Info().Str("token", token).Msg("[AuthHandler-VerifyUserAccount] User account verified successfully")

	return c.JSON(http.StatusOK, resp)
}

func (a *AuthHandler) ForgotPassword(c echo.Context) error {
	var (
		req  = request.ForgotPasswordRequest{}
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
	)

	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("[AuthHandler-ForgotPassword] Failed to bind request")
		resp.Message = "Invalid request format"
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := a.validator.Validate(&req); err != nil {
		log.Error().Err(err).Msg("[AuthHandler-ForgotPassword] Validation failed")
		resp.Message = err.Error()
		return c.JSON(http.StatusBadRequest, resp)
	}

	err := a.userService.ForgotPassword(ctx, req.Email)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("[AuthHandler-ForgotPassword] Password reset request failed")

		switch err.Error() {
		case "invalid email format":
			resp.Message = "Invalid email format"
			return c.JSON(http.StatusUnprocessableEntity, resp)
		case "failed to process request", "failed to generate reset token", "failed to create reset token":
			resp.Message = "Failed to process request"
			return c.JSON(http.StatusInternalServerError, resp)
		default:
			resp.Message = "Internal server error"
			return c.JSON(http.StatusInternalServerError, resp)
		}
	}

	resp.Message = "If an account with this email exists, you will receive a password reset link."
	log.Info().Str("email", req.Email).Msg("[AuthHandler-ForgotPassword] Password reset request processed successfully")

	return c.JSON(http.StatusOK, resp)
}

func (a *AuthHandler) ResetPassword(c echo.Context) error {
	var (
		req  = request.ResetPasswordRequest{}
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
	)

	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("[AuthHandler-ResetPassword] Failed to bind request")
		resp.Message = "Invalid request format"
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := a.validator.Validate(&req); err != nil {
		log.Error().Err(err).Msg("[AuthHandler-ResetPassword] Validation failed")

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				fieldName := fieldError.Field()
				tag := fieldError.Tag()

				switch fieldName {
				case "Token":
					if tag == "required" {
						resp.Message = "Reset token is required"
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
				case "Password":
					if tag == "required" {
						resp.Message = "Password is required"
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
					if tag == "min" {
						resp.Message = "Password must be at least 8 characters long"
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
				case "PasswordConfirmation":
					if tag == "required" {
						resp.Message = "Password confirmation is required"
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
					if tag == "eqfield" {
						resp.Message = "Password confirmation does not match"
						return c.JSON(http.StatusUnprocessableEntity, resp)
					}
				}
			}
		}

		resp.Message = "Validation failed"
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	err := a.userService.ResetPassword(ctx, req.Token, req.Password, req.PasswordConfirmation)
	if err != nil {
		log.Error().Err(err).Str("token", req.Token).Msg("[AuthHandler-ResetPassword] Password reset failed")

		switch err.Error() {
		case "invalid or expired reset token":
			resp.Message = "Invalid or expired reset token"
			return c.JSON(http.StatusBadRequest, resp)
		case "invalid token type":
			resp.Message = "Invalid token type"
			return c.JSON(http.StatusBadRequest, resp)
		case "password is required", "password must be at least 8 characters long", "password confirmation does not match":
			resp.Message = err.Error()
			return c.JSON(http.StatusUnprocessableEntity, resp)
		case "failed to validate token", "failed to process password", "failed to update password":
			resp.Message = "Failed to reset password"
			return c.JSON(http.StatusInternalServerError, resp)
		default:
			resp.Message = "Internal server error"
			return c.JSON(http.StatusInternalServerError, resp)
		}
	}

	resp.Message = "Password reset successfully. You can now sign in with your new password."
	log.Info().Str("token", req.Token).Msg("[AuthHandler-ResetPassword] Password reset successfully")

	return c.JSON(http.StatusOK, resp)
}

func NewAuthHandler(userService port.UserServiceInterface) AuthHandlerInterface {
	return &AuthHandler{
		userService: userService,
		validator:   myvalidator.NewValidator(),
	}
}

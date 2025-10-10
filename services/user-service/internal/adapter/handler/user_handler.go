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

func NewUserHandler(userService port.UserServiceInterface) UserHandlerInterface {
	return &UserHandler{
		userService: userService,
		validator:   validator.NewValidator(),
	}
}

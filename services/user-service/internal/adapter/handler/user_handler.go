package handler

import (
	"user-service/internal/core/port"
)

type UserHandlerInterface interface {
	AuthHandlerInterface
	AdminHandlerInterface
}

type UserHandler struct {
	AuthHandlerInterface
	AdminHandlerInterface
}

func NewUserHandler(userService port.UserServiceInterface) UserHandlerInterface {
	return &UserHandler{
		AuthHandlerInterface:  NewAuthHandler(userService),
		AdminHandlerInterface: NewAdminHandler(userService),
	}
}

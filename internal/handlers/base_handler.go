package handlers

import (
	TokenService "github.com/aerosystems/nix-junior-chat-back/internal/services/token_service"
	"os"

	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
)

type BaseHandler struct {
	userRepo     models.UserRepository
	messageRepo  models.MessageRepository
	tokenService TokenService.Service
}

// Response is the type used for sending JSON around
type Response struct {
	Error   bool   `json:"error" example:"false"`
	Message string `json:"message" example:"success operation"`
	Data    any    `json:"data,omitempty"`
}

func NewBaseHandler(
	userRepo models.UserRepository,
	messageRepo models.MessageRepository,
	tokenService TokenService.Service,
) *BaseHandler {
	return &BaseHandler{
		userRepo:     userRepo,
		messageRepo:  messageRepo,
		tokenService: tokenService,
	}
}

// SuccessResponse takes a response status code and arbitrary data and writes a json response to the client in depends on Header Accept
func SuccessResponse(c echo.Context, statusCode int, message string, data any) error {
	payload := Response{
		Error:   false,
		Message: message,
		Data:    data,
	}
	return c.JSON(statusCode, payload)
}

// ErrorResponse takes a response status code and arbitrary data and writes a json response to the client in depends on Header Accept and APP_ENV environment variable(has two possible values: dev and prod)
// - APP_ENV=dev responds debug info level of error
// - APP_ENV=prod responds just message about error [DEFAULT]
func ErrorResponse(c echo.Context, statusCode int, message string, err error) error {
	payload := Response{
		Error:   true,
		Message: message,
		Data:    err.Error(),
	}

	if os.Getenv("APP_ENV") == "prod" {
		payload.Data = nil
	}

	return c.JSON(statusCode, payload)
}

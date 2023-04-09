package handlers

import (
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
)

type BaseHandler struct {
	userRepo   models.UserRepository
	tokensRepo models.TokensRepository
}

// Response is the type used for sending JSON around
type Response struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func NewBaseHandler(
	userRepo models.UserRepository,
	tokensRepo models.TokensRepository,
) *BaseHandler {
	return &BaseHandler{
		userRepo:   userRepo,
		tokensRepo: tokensRepo,
	}
}

// WriteResponse takes a response status code and arbitrary data and writes a json response to the client in depends on Header Accept
func WriteResponse(c echo.Context, statusCode int, payload any) error {
	return c.JSON(statusCode, payload)
}

func NewErrorPayload(err error) Response {
	return Response{
		Error:   true,
		Message: err.Error(),
	}
}

package handlers

import (
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
)

type BaseHandler struct {
	userRepo   models.UserRepository
	codeRepo   models.CodeRepository
	tokensRepo models.TokensRepository
}

// Response is the type used for sending JSON around
type Response struct {
	Error   bool   `json:"error" xml:"error"`
	Message string `json:"message" xml:"message"`
	Data    any    `json:"data,omitempty" xml:"data,omitempty"`
}

func NewBaseHandler(
	userRepo models.UserRepository,
	codeRepo models.CodeRepository,
	tokensRepo models.TokensRepository,
) *BaseHandler {
	return &BaseHandler{
		userRepo:   userRepo,
		codeRepo:   codeRepo,
		tokensRepo: tokensRepo,
	}
}

// WriteResponse takes a response status code and arbitrary data and writes a xml/json response to the client in depends on Header Accept
func WriteResponse(c echo.Context, statusCode int, payload any) error {
	acceptHeaders := c.Request().Header["Accept"]
	if Contains(acceptHeaders, "application/xml") {
		return c.XML(statusCode, payload)
	}
	return c.JSON(statusCode, payload)
}

func NewErrorPayload(err error) Response {
	return Response{
		Error:   true,
		Message: err.Error(),
	}
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

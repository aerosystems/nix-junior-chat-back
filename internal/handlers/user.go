package handlers

import (
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
)

// User godoc
// @Summary Get user data
// @Description Get user data
// @Tags user
// @Accept  json
// @Produce application/json
// @Param Authorization header string true "should contain Access Token, with the Bearer started"
// @Success 200 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user [get]
func (h *BaseHandler) User(c echo.Context) error {
	user := c.Get("user").(*models.User)

	return SuccessResponse(c, http.StatusOK, "successfully found user data", user)
}
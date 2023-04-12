package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/aerosystems/nix-junior-chat-back/internal/models"
)

// Logout godoc
// @Summary logout user
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param Authorization header string true "should contain Access Token, with the Bearer started"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/logout [post]
func (h *BaseHandler) Logout(c echo.Context) error {
	// receive AccessToken Claims from context middleware
	accessTokenClaims, ok := c.Get("user").(*models.AccessTokenClaims)
	if !ok {
		err := errors.New("internal transport token error")
		return ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
	}

	err := h.tokensRepo.DropCacheTokens(*accessTokenClaims)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error clearing cache token", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("user %s successfully logged out", accessTokenClaims.AccessUUID), accessTokenClaims)
}

package handlers

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"

	"github.com/aerosystems/nix-junior-chat-back/internal/models"
)

// Logout godoc
// @Summary logout user
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param Authorization header string true "should contain Access Token, with the Bearer started"
// @Success 202 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Router /user/logout [post]
func (h *BaseHandler) Logout(c echo.Context) error {
	// receive AccessToken Claims from context middleware
	accessTokenClaims, ok := c.Get("user").(*models.AccessTokenClaims)
	if !ok {
		err := errors.New("token is untracked")
		return WriteResponse(c, http.StatusUnauthorized, NewErrorPayload(err))
	}

	err := h.tokensRepo.DropCacheTokens(*accessTokenClaims)
	if err != nil {
		return WriteResponse(c, http.StatusUnauthorized, NewErrorPayload(err))
	}

	payload := Response{
		Error:   false,
		Message: fmt.Sprintf("User %s successfully logged out", accessTokenClaims.AccessUUID),
		Data:    accessTokenClaims,
	}
	return WriteResponse(c, http.StatusAccepted, payload)
}

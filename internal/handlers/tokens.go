package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type RefreshTokenRequestBody struct {
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

type TokensResponseBody struct {
	AccessToken  string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

// RefreshTokens godoc
// @Summary refresh pair JWT token_service
// @Tags token_service
// @Accept  json
// @Produce application/json
// @Param login body RefreshTokenRequestBody true "raw request body, should contain Refresh Token"
// @Success 200 {object} Response{data=TokensResponseBody}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/token/refresh [post]
func (h *BaseHandler) RefreshTokens(c echo.Context) error {
	var requestPayload RefreshTokenRequestBody

	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	if requestPayload == (RefreshTokenRequestBody{}) {
		err := fmt.Errorf("empty request body")
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	// validate & parse refresh token claims
	refreshTokenClaims, err := h.tokenService.DecodeRefreshToken(requestPayload.RefreshToken)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid refresh token", err)
	}

	if value, _ := h.tokenService.GetCacheValue(refreshTokenClaims.RefreshUUID); value == nil {
		err := fmt.Errorf("refresh token %s not found in cache", refreshTokenClaims.RefreshUUID)
		return ErrorResponse(c, http.StatusUnauthorized, "refresh token expired", err)
	}

	// drop Refresh Token from Redis Cache
	_ = h.tokenService.DropCacheKey(refreshTokenClaims.RefreshUUID)

	// create a pair of JWT token_service
	ts, err := h.tokenService.CreateToken(refreshTokenClaims.UserID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error creating token_service", err)
	}

	// add refresh token UUID to cache
	err = h.tokenService.CreateCacheKey(refreshTokenClaims.UserID, ts)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error creating cache key", err)
	}

	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("token_service successfuly refreshed for User %d", refreshTokenClaims.UserID), tokens)
}

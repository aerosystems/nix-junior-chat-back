package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
)

type RefreshTokenRequestBody struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

// RefreshTokens godoc
// @Summary refresh pair JWT tokens
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param login body handlers.RefreshTokenRequestBody true "raw request body, should contain Refresh Token"
// @Param Authorization header string true "should contain Access Token, with the Bearer started"
// @Success 200 {object} Response{data=TokensResponseBody}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/token/refresh [post]
func (h *BaseHandler) RefreshTokens(c echo.Context) error {
	// receive AccessToken Claims from context middleware
	accessTokenClaims, ok := c.Get("user").(*models.AccessTokenClaims)
	if !ok {
		err := errors.New("internal transport token error")
		return ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
	}

	// getting Refresh Token from Redis cache
	cacheJSON, _ := h.tokensRepo.GetCacheValue(accessTokenClaims.AccessUUID)
	accessTokenCache := new(models.AccessTokenCache)
	err := json.Unmarshal([]byte(*cacheJSON), accessTokenCache)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error unmarshaling cache token", err)
	}
	cacheRefreshTokenUUID := accessTokenCache.RefreshUUID

	var requestPayload RefreshTokenRequestBody

	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	// validate & parse refresh token claims
	refreshTokenClaims, err := h.tokensRepo.DecodeRefreshToken(requestPayload.RefreshToken)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "error decoding refresh token", err)
	}
	requestRefreshTokenUUID := refreshTokenClaims.RefreshUUID

	// drop Access & Refresh Tokens from Redis Cache
	_ = h.tokensRepo.DropCacheTokens(*accessTokenClaims)

	// compare RefreshToken UUID from Redis cache & Request body
	if requestRefreshTokenUUID != cacheRefreshTokenUUID {
		// drop request RefreshToken UUID from cache
		_ = h.tokensRepo.DropCacheKey(requestRefreshTokenUUID)
		err := fmt.Errorf("refresh token %s in request body does not match refresh token which publish access token", requestRefreshTokenUUID)
		return ErrorResponse(c, http.StatusBadRequest, "hmmm... refresh token in request body does not match refresh token which publish access token", err)
	}

	// create a pair of JWT tokens
	ts, err := h.tokensRepo.CreateToken(refreshTokenClaims.UserID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error creating tokens", err)
	}

	// add refresh token UUID to cache
	err = h.tokensRepo.CreateCacheKey(refreshTokenClaims.UserID, ts)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error creating cache key", err)
	}

	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("tokens successfuly refreshed for User %d", refreshTokenClaims.UserID), tokens)
}

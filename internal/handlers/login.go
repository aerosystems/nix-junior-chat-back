package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/aerosystems/nix-junior-chat-back/internal/helpers"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type LoginRequestBody struct {
	Username string `json:"username" example:"username"`
	Password string `json:"password" example:"P@ssw0rd"`
}

// Login godoc
// @Summary login user by credentials
// @Description Username should contain:
// @Description - lower, upper case latin letters and digits
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Description Password should contain:
// @Description - minimum of one small case letter
// @Description - minimum of one upper case letter
// @Description - minimum of one digit
// @Description - minimum of one special character
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Description Response contain pair JWT tokens, use /v1/tokens/refresh for updating them
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param login body LoginRequestBody true "raw request body"
// @Success 200 {object} Response{data=TokensResponseBody}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/login [post]
func (h *BaseHandler) Login(c echo.Context) error {
	var requestPayload LoginRequestBody

	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	if err := helpers.ValidateUsername(requestPayload.Username); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid username", err)
	}

	if err := helpers.ValidatePassword(requestPayload.Password); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid password", err)
	}

	//checking if user is existing
	user, err := h.userRepo.FindByUsername(requestPayload.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		return ErrorResponse(c, http.StatusBadRequest, "error while finding user", err)
	}
	if user == nil {
		err := fmt.Errorf("user with username %s does not exist", requestPayload.Username)
		return ErrorResponse(c, http.StatusNotFound, "user not found", err)
	}

	valid, err := h.userRepo.PasswordMatches(user, requestPayload.Password)
	if err != nil || !valid {
		err := errors.New("invalid credentials")
		return ErrorResponse(c, http.StatusUnauthorized, err.Error(), err)
	}

	// create pair JWT tokens
	ts, err := h.tokensRepo.CreateToken(user.ID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error creating tokens", err)
	}

	// add refresh token UUID to cache
	if err = h.tokensRepo.CreateCacheKey(user.ID, ts); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error creating cache tokens", err)
	}

	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("user %s is logged in", requestPayload.Username), tokens)
}

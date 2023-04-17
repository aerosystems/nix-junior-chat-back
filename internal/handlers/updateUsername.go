package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/aerosystems/nix-junior-chat-back/internal/helpers"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
)

type UpdateUsernameRequestBody struct {
	Username string `json:"username" example:"username"`
}

// UpdateUsername godoc
// @Summary update username
// @Description Username should contain:
// @Description - lower, upper case latin letters and digits
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Tags user
// @Accept  json
// @Produce application/json
// @Param Authorization header string true "should contain Access Token, with the Bearer started"
// @Param username body UpdateUsernameRequestBody true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/update-username [put]
func (h *BaseHandler) UpdateUsername(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	fmt.Println(user)
	if !ok {
		err := errors.New("internal transport token error")
		return ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
	}

	var requestPayload UpdateUsernameRequestBody
	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	if err := helpers.ValidateUsername(requestPayload.Username); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "username is incorrect", err)
	}

	if tmpUser, _ := h.userRepo.FindByUsername(requestPayload.Username); tmpUser != nil {
		err := errors.New("username is already taken")
		return ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
	}

	user.Username = requestPayload.Username

	if err := h.userRepo.Update(user); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error updating user", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("username successfuly updated to %s", user.Username), nil)
}

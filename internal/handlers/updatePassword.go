package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/aerosystems/nix-junior-chat-back/internal/helpers"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
)

type UpdatePasswordRequestBody struct {
	NewPassword string `json:"new_password" example:"NewP@ssw0rd"`
	OldPassword string `json:"old_password" example:"OldP@ssw0rd"`
}

// UpdatePassword godoc
// @Summary update password
// @Description OldPassword/NewPassword should contain:
// @Description - minimum of one small case letter
// @Description - minimum of one upper case letter
// @Description - minimum of one digit
// @Description - minimum of one special character
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Tags user
// @Accept  json
// @Produce application/json
// @Param password body handlers.UpdatePasswordRequestBody true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/update-password [put]
func (h *BaseHandler) UpdatePassword(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		err := errors.New("internal transport token error")
		return ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
	}

	var requestPayload UpdatePasswordRequestBody

	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	if err := helpers.ValidatePassword(requestPayload.OldPassword); err != nil {
		prefixErr := errors.New("old password is incorrect. ")
		err = fmt.Errorf("%w%v", prefixErr, err)
		return ErrorResponse(c, http.StatusBadRequest, prefixErr.Error(), err)
	}

	if err := helpers.ValidatePassword(requestPayload.NewPassword); err != nil {
		prefixErr := errors.New("new password is incorrect. ")
		err = fmt.Errorf("%w%v", prefixErr, err)
		return ErrorResponse(c, http.StatusBadRequest, prefixErr.Error(), err)
	}

	ok, err := h.userRepo.PasswordMatches(user, requestPayload.OldPassword)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error matching passwords", err)
	}
	if !ok {
		err := errors.New("old password is incorrect")
		return ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
	}

	if err := h.userRepo.ResetPassword(user, requestPayload.NewPassword); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error updating password", err)
	}

	return SuccessResponse(c, http.StatusOK, "password successfuly updated", nil)
}

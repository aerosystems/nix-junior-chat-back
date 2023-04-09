package handlers

import (
	"fmt"
	"net/http"

	"github.com/aerosystems/nix-junior-chat-back/internal/helpers"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type RegistrationRequestBody struct {
	Username string `json:"username" example:"username"`
	Password string `json:"password" example:"P@ssw0rd"`
}

// Registration godoc
// @Summary registration user by credentials
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
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param registration body handlers.RegistrationRequestBody true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /user/register [post]
func (h *BaseHandler) Registration(c echo.Context) error {
	var requestPayload RegistrationRequestBody

	if err := c.Bind(&requestPayload); err != nil {
		return WriteResponse(c, http.StatusBadRequest, NewErrorPayload(err))
	}

	if err := helpers.ValidateUsername(requestPayload.Username); err != nil {
		return WriteResponse(c, http.StatusBadRequest, NewErrorPayload(err))
	}

	err := helpers.ValidatePassword(requestPayload.Password)
	if err != nil {
		return WriteResponse(c, http.StatusBadRequest, NewErrorPayload(err))
	}

	var payload Response

	//checking if username is existing
	user, _ := h.userRepo.FindByUsername(requestPayload.Username)
	if user != nil {
		err = fmt.Errorf("user with username %s already exists", requestPayload.Username)
		return WriteResponse(c, http.StatusBadRequest, NewErrorPayload(err))
	}

	// hashing password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestPayload.Password), 12)
	if err != nil {
		return err
	}

	// creating new user
	newUser := models.User{
		Username: requestPayload.Username,
		Password: string(hashedPassword),
	}
	err = h.userRepo.Create(&newUser)

	if err != nil {
		return WriteResponse(c, http.StatusBadRequest, NewErrorPayload(err))
	}

	// TODO makemethods for cornirm user account

	payload = Response{
		Error:   false,
		Message: fmt.Sprintf("user with username: %s successfully created", requestPayload.Username),
		Data:    nil,
	}

	return WriteResponse(c, http.StatusOK, payload)
}

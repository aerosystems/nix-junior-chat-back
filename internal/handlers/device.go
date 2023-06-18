package handlers

import (
	"errors"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/aerosystems/nix-junior-chat-back/pkg/validators"
	"github.com/labstack/echo/v4"
	"net/http"
)

type AddDeviceRequestBody struct {
	Type  string `json:"type" example:"web"`
	Token string `json:"token" example:"token"`
}

// AddDevice godoc
// @Summary Add device
// @Description Set device Token(Firebase Cloud Messaging) for push notifications
// @Tags device
// @Accept  json
// @Produce application/json
// @Param password body AddDeviceRequestBody true "raw request body"
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 409 {object} Response
// @Failure 500 {object} Response
// @Router /v1/devices [post]
func (h *BaseHandler) AddDevice(c echo.Context) error {
	user := c.Get("user").(*models.User)
	userAgent := c.Request().UserAgent()
	var rawDataDevice AddDeviceRequestBody

	if err := c.Bind(&rawDataDevice); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid device data", err)
	}

	if err := validators.ValidateDeviceType(rawDataDevice.Type); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid device type", err)
	}

	for _, device := range user.Devices {
		if device.Token == rawDataDevice.Token {
			return ErrorResponse(c, http.StatusConflict, "device with this token already exists", errors.New("device with this token already exists"))
		}
	}

	user.Devices = append(user.Devices, &models.Device{
		Type:  rawDataDevice.Type,
		Token: rawDataDevice.Token,
		Name:  userAgent,
	})

	if err := h.userRepo.UpdateWithAssociations(user); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error updating user", err)
	}

	return SuccessResponse(c, http.StatusOK, "successfully set device", user)
}

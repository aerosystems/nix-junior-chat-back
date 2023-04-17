package handlers

import (
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
)

type BlockRequestBody struct {
	UserID int `json:"userId" example:"1"`
}

// Block godoc
// @Summary Block user
// @Tags relationship
// @Accept  json
// @Produce application/json
// @Param Authorization header string true "should contain Access Token, with the Bearer started"
// @Param follow body BlockRequestBody true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/block [post]
func (h *BaseHandler) Block(c echo.Context) error {
	user := c.Get("user").(*models.User)
	var requestPayload BlockRequestBody

	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request payload", err)
	}

	blockedUser, err := h.userRepo.FindByID(requestPayload.UserID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}
	if blockedUser == nil {
		err := fmt.Errorf("user with id %d not found", requestPayload.UserID)
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}
	if blockedUser.ID == user.ID {
		err := fmt.Errorf("user with id %d is the same as current user", requestPayload.UserID)
		return ErrorResponse(c, http.StatusBadRequest, "you can't block yourself", err)
	}
	for _, item := range user.BlockedUsers {
		if item.ID == blockedUser.ID {
			err := fmt.Errorf("user with id %d is already blocked", requestPayload.UserID)
			return ErrorResponse(c, http.StatusBadRequest, "user is already blocked", err)
		}
	}

	user.BlockedUsers = append(user.BlockedUsers, blockedUser)

	if err := h.userRepo.Update(user); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error updating user", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("successfully blacklisted user %s (id: %d)", blockedUser.Username, blockedUser.ID), user)
}

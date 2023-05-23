package handlers

import (
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

// Block godoc
// @Summary Block user
// @Tags relationship
// @Accept  json
// @Produce application/json
// @Param	id	path	int	true	"Blocked User ID"
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/block/{id} [post]
func (h *BaseHandler) Block(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("id")
	blockedUserID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid blocked userId", err)
	}
	blockedUser, err := h.userRepo.FindByID(blockedUserID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}
	if blockedUser == nil {
		err := fmt.Errorf("user with id %d not found", blockedUserID)
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}
	if blockedUser.ID == user.ID {
		err := fmt.Errorf("user with id %d is the same as current user", blockedUserID)
		return ErrorResponse(c, http.StatusBadRequest, "you can't block yourself", err)
	}
	for _, item := range user.BlockedUsers {
		if item.ID == blockedUser.ID {
			err := fmt.Errorf("user with id %d is already blocked", blockedUserID)
			return ErrorResponse(c, http.StatusBadRequest, "user is already blocked", err)
		}
	}

	user.BlockedUsers = append(user.BlockedUsers, blockedUser)

	if err := h.userRepo.Update(user); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error updating user", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("successfully blacklisted user %s (id: %d)", blockedUser.Username, blockedUser.ID), user)
}

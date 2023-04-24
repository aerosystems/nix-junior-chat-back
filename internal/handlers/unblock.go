package handlers

import (
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

// Unblock godoc
// @Summary Unblock user
// @Tags relationship
// @Accept  json
// @Produce application/json
// @Param	id	path	int	true	"Unblocked User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/block/{id} [delete]
func (h *BaseHandler) Unblock(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("id")
	blockedUserID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid unfollowed userId", err)
	}

	for i, item := range user.BlockedUsers {
		if item.ID == blockedUserID {
			followedUsers := append(user.BlockedUsers[:i], user.BlockedUsers[i+1:]...)
			err := h.userRepo.ReplaceBlockedUsers(user, followedUsers)
			if err != nil {
				return ErrorResponse(c, http.StatusInternalServerError, "failed to unblock user", err)
			}

			updatedUser, _ := h.userRepo.FindByID(user.ID)
			updatedUser.ModifyImage()
			return SuccessResponse(c, http.StatusOK, "successfully unblocked user", updatedUser)
		}
	}
	err = fmt.Errorf("user with id %d is not blocked", blockedUserID)
	return ErrorResponse(c, http.StatusNotFound, "user is not blocked", err)
}

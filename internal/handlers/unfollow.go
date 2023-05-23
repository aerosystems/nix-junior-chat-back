package handlers

import (
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

// Unfollow godoc
// @Summary Unfollow user
// @Tags relationship
// @Accept  json
// @Produce application/json
// @Param	id	path	int	true	"Followed User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/follow/{id} [delete]
func (h *BaseHandler) Unfollow(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("id")
	followedUserID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid unfollowed userId", err)
	}

	for i, item := range user.FollowedUsers {
		if item.ID == followedUserID {
			followedUsers := append(user.FollowedUsers[:i], user.FollowedUsers[i+1:]...)
			err := h.userRepo.ReplaceFollowedUsers(user, followedUsers)
			if err != nil {
				return ErrorResponse(c, http.StatusInternalServerError, "failed to unfollow user", err)
			}

			return SuccessResponse(c, http.StatusOK, "successfully unfollowed user", user)
		}
	}
	err = fmt.Errorf("user with id %d is not followed", followedUserID)
	return ErrorResponse(c, http.StatusNotFound, "user is not followed", err)
}

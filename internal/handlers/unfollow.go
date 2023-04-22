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
		fmt.Println(i, item.ID)
		if item.ID == followedUserID {
			//if err := h.userRepo.DeleteFollowedUser(user, item); err != nil {
			//	return ErrorResponse(c, http.StatusInternalServerError, "error updating user", err)
			//}

			var followedUsers []*models.User
			followedUsers = append(user.FollowedUsers[:i], user.FollowedUsers[i+1:]...)

			h.userRepo.ReplaceFollowedUsers(user, followedUsers)

			updatedUser, _ := h.userRepo.FindByID(user.ID)
			updatedUser.ModifyImage()
			return SuccessResponse(c, http.StatusOK, "successfully unfollowed user", updatedUser)
		}
	}
	err = fmt.Errorf("user with id %d is not followed", followedUserID)
	return ErrorResponse(c, http.StatusNotFound, "user is not followed", err)
}

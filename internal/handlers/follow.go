package handlers

import (
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

// Follow godoc
// @Summary Follow user
// @Tags relationship
// @Accept  json
// @Produce application/json
// @Param	id	path	int	true	"Followed User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/follow/{id} [post]
func (h *BaseHandler) Follow(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("id")
	followedUserID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}

	followedUser, err := h.userRepo.FindByID(followedUserID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}
	if followedUser == nil {
		err := fmt.Errorf("user with id %d not found", followedUserID)
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}
	if followedUser.ID == user.ID {
		err := fmt.Errorf("user with id %d is the same as current user", followedUserID)
		return ErrorResponse(c, http.StatusBadRequest, "you can't follow yourself", err)
	}
	for _, item := range user.FollowedUsers {
		if item.ID == followedUser.ID {
			err := fmt.Errorf("user with id %d is already followed", followedUserID)
			return ErrorResponse(c, http.StatusBadRequest, "user is already followed", err)
		}
	}

	user.FollowedUsers = append(user.FollowedUsers, followedUser)

	if err := h.userRepo.Update(user); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error updating user", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("successfully followed user %s (id: %d)", followedUser.Username, followedUser.ID), user)
}

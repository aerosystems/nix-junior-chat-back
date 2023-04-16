package handlers

import (
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
)

type FollowRequest struct {
	UserID int `json:"userId" example:"1"`
}

// Follow godoc
// @Summary Follow user
// @Tags relationship
// @Accept  json
// @Produce application/json
// @Param Authorization header string true "should contain Access Token, with the Bearer started"
// @Param follow body FollowRequest true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/follow [post]
func (h *BaseHandler) Follow(c echo.Context) error {
	user := c.Get("user").(*models.User)
	var requestPayload FollowRequest

	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request payload", err)
	}

	followedUser, err := h.userRepo.FindByID(requestPayload.UserID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}
	if followedUser == nil {
		err := fmt.Errorf("user with id %d not found", requestPayload.UserID)
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}
	if followedUser.ID == user.ID {
		err := fmt.Errorf("user with id %d is the same as current user", requestPayload.UserID)
		return ErrorResponse(c, http.StatusBadRequest, "you can't follow yourself", err)
	}
	for _, item := range user.FollowedUsers {
		if item.ID == followedUser.ID {
			err := fmt.Errorf("user with id %d is already followed", requestPayload.UserID)
			return ErrorResponse(c, http.StatusBadRequest, "user is already followed", err)
		}
	}

	user.FollowedUsers = append(user.FollowedUsers, followedUser)

	if err := h.userRepo.Update(user); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error updating user", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("successfully followed user %s (id: %d)", followedUser.Username, followedUser.ID), user)
}

package handlers

import (
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

// DeleteChat godoc
// @Summary DeleteChat user
// @Tags relationship
// @Accept  json
// @Produce application/json
// @Param	id	path	int	true	"Chat User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/chat/{id} [delete]
func (h *BaseHandler) DeleteChat(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("id")
	chatUserID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid chat's userId", err)
	}

	for i, item := range user.Chats {
		if item.ID == chatUserID {
			chatUsers := append(user.Chats[:i], user.Chats[i+1:]...)
			err := h.userRepo.ReplaceChatUsers(user, chatUsers)
			if err != nil {
				return ErrorResponse(c, http.StatusInternalServerError, "failed to delete chat with user", err)
			}

			return SuccessResponse(c, http.StatusOK, "successfully deleted chat with user", user)
		}
	}
	err = fmt.Errorf("chat with userId %d does not exist", chatUserID)
	return ErrorResponse(c, http.StatusNotFound, "chat with user does not exist", err)
}

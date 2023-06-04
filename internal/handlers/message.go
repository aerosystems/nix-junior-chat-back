package handlers

import (
	"errors"
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
	"reflect"
	"strconv"
)

// GetMessages godoc
// @Summary Get messages from chat by ChatId
// @Tags chat
// @Accept  json
// @Produce application/json
// @Param	chat_id	path	int	true	"Chat ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=[]models.Message}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/chat/messages/{chat_id} [get]
func (h *BaseHandler) GetMessages(c echo.Context) error {
	user := c.Get("user").(*models.User)
	chatID, err := strconv.Atoi(c.Param("chat_id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid chat id", err)
	}
	chat, err := h.chatRepo.FindByID(chatID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid chat", err)
	}
	if chat == nil || reflect.DeepEqual(*chat, models.Chat{}) {
		err := fmt.Errorf("chat with id %d not found", chatID)
		return ErrorResponse(c, http.StatusBadRequest, "chat not found", err)
	}
	for _, u := range chat.Users {
		if u.ID == user.ID {
			limit := 10
			limitStr := c.QueryParam("limit")
			if limitStr != "" {
				var err error
				limit, err = strconv.Atoi(limitStr)
				if err != nil {
					return ErrorResponse(c, http.StatusBadRequest, "invalid limit, limit must be integer", err)
				}
				if limit < 1 || limit > 1000 {
					err := errors.New("invalid limit, available limits: '1-1000'")
					return ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
				}
			}

			from := 0
			fromStr := c.QueryParam("from")
			if fromStr != "" {
				var err error
				from, err = strconv.Atoi(fromStr)
				if err != nil {
					return ErrorResponse(c, http.StatusBadRequest, "invalid limit, limit must be integer", err)
				}
			}

			messages, err := h.messageRepo.GetMessages(chatID, from, limit)
			if err != nil {
				return ErrorResponse(c, http.StatusInternalServerError, "failed to get messages", err)
			}
			return SuccessResponse(c, http.StatusOK, "messages found successfully", messages)
		}
	}

	err = fmt.Errorf("user with id %d not found in chat with id %d", user.ID, chatID)
	return ErrorResponse(c, http.StatusBadRequest, "user not found in chat", err)

}

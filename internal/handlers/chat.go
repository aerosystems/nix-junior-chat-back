package handlers

import (
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	ChatService "github.com/aerosystems/nix-junior-chat-back/internal/services/chat_service"
	"github.com/aerosystems/nix-junior-chat-back/pkg/redisclient"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type MessageResponseBody struct {
	Content     string `json:"content" example:"bla-bla-bla"`
	RecipientID int    `json:"recipientId" example:"1"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // TODO: remove this line in production
}

// Chat godoc
// @Summary Chat [WebSocket]
// @Description Chat with users based on WebSocket
// @Tags chat
// @Param token query string true "Access JWT Token"
// @Param chat body MessageResponseBody true "body should contain content and recipient_id for sending message"
// @Failure 401 {object} Response
// @Router /ws/chat [get]
func (h *BaseHandler) Chat(c echo.Context) error {
	clientREDIS := redisclient.NewClient()
	token := c.QueryParam("token")
	accessTokenClaims, err := h.tokenService.DecodeAccessToken(token)
	if err != nil {
		return ErrorResponse(c, 401, "invalid token", err)
	}

	_, err = h.tokenService.GetCacheValue(accessTokenClaims.AccessUUID)
	if err != nil {
		return ErrorResponse(c, 401, "invalid token", err)
	}

	user, err := h.userRepo.FindByID(accessTokenClaims.UserID)
	if err != nil {
		return ErrorResponse(c, 401, "user not found", err)
	}

	c.Logger().Info(fmt.Sprintf("client %d connected", user.ID))

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ChatService.OnConnect(user, ws, clientREDIS); err != nil {
		ChatService.HandleWSError(err, ws)
	}

	closeCh := ChatService.OnDisconnect(user, ws)

	ChatService.OnChannelMessage(ws, user)

loop:
	for {
		select {
		case <-closeCh:
			break loop
		default:
			ChatService.OnClientMessage(ws, user, clientREDIS)
		}
	}

	return nil
}

// DeleteChat godoc
// @Summary Delete Chat by ChatID
// @Tags user
// @Accept  json
// @Produce application/json
// @Param	chat_id	path	int	true	"Chat ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 403 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/chat/{chat_id} [delete]
func (h *BaseHandler) DeleteChat(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("chat_id")
	chatID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid chat's userId", err)
	}

	chat, err := h.chatRepo.FindByID(chatID)
	if err != nil {
		return ErrorResponse(c, http.StatusNotFound, "chat not found", err)
	}

	if chat.Type != "private" {
		return ErrorResponse(c, http.StatusForbidden, "does not have access to delete this chat", nil)
	}

	for _, item := range chat.Users {
		if item.ID == user.ID {
			err := h.chatRepo.Delete(chat)
			if err != nil {
				return ErrorResponse(c, http.StatusInternalServerError, "failed to delete chat", err)
			}

			return SuccessResponse(c, http.StatusOK, "successfully deleted chat", user)
		}
	}

	return ErrorResponse(c, http.StatusForbidden, "does not have access to delete this chat", nil)
}

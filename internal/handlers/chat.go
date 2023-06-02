package handlers

import (
	"fmt"
	ChatService "github.com/aerosystems/nix-junior-chat-back/internal/services/chat_service"
	"github.com/aerosystems/nix-junior-chat-back/pkg/redisclient"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"net/http"
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

package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/aerosystems/nix-junior-chat-back/pkg/myredis"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type MessageResponseBody struct {
	Content     string `json:"content" example:"bla-bla-bla"`
	RecipientID int    `json:"recipientId" example:"1"`
}

// Storage for clients
var clients []*models.Client

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
	clientREDIS := myredis.NewClient()
	token := c.QueryParam("token")
	accessTokenClaims, err := h.tokensRepo.DecodeAccessToken(token)
	if err != nil {
		return ErrorResponse(c, 401, "invalid token", err)
	}

	_, err = h.tokensRepo.GetCacheValue(accessTokenClaims.AccessUUID)
	if err != nil {
		return ErrorResponse(c, 401, "invalid token", err)
	}

	sender, err := h.userRepo.FindByID(accessTokenClaims.UserID)
	if err != nil {
		return ErrorResponse(c, 401, "user not found", err)
	}

	c.Logger().Info(fmt.Sprintf("client %d connected", sender.ID))
	clientREDIS.Set(fmt.Sprintf("is-online:%d", sender.ID), true, time.Minute)

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	client := &models.Client{
		WS:   ws,
		User: *sender,
	}

	clients = append(clients, client)

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			for i, client := range clients {
				if client.WS == ws {
					clients = append(clients[:i], clients[i+1:]...)
				}
				c.Logger().Error(fmt.Errorf("client %d disconnected", client.User.ID))
				clientREDIS.Del(fmt.Sprintf("is-online:%d", client.User.ID))
				sender.LastActive = time.Now().Unix()
				if err := h.userRepo.Update(sender); err != nil {
					c.Logger().Error(err)
				}
				break
			}
			c.Logger().Error(err)
			break
		}

		for _, client := range clients {
			var responseMessage MessageResponseBody
			if err := json.Unmarshal(msg, &responseMessage); err != nil {
				reply := models.NewErrorMessage("invalid message format", *sender)
				client.WS.WriteMessage(websocket.TextMessage, reply.Json())
				c.Logger().Error(err)
				continue
			}

			recipient, err := h.userRepo.FindByID(responseMessage.RecipientID)
			if err != nil {
				reply := models.NewErrorMessage("invalid recipient id", *sender)
				client.WS.WriteMessage(websocket.TextMessage, reply.Json())
				c.Logger().Error(err)
				continue
			}
			message := models.NewTextMessage(responseMessage.Content, *sender, responseMessage.RecipientID)

			if client.User.ID == message.RecipientID {
				client.WS.WriteMessage(websocket.TextMessage, message.Json())
				h.messageRepo.Create(message)
				// Adding chat to sender
				status := false
				for _, item := range sender.Chats {
					if item.ID == recipient.ID {
						status = true
						break
					}
				}
				if !status {
					sender.Chats = append(sender.Chats, recipient)
					h.userRepo.Update(sender)
				}
				// Adding chat to recipient
				status = false
				for _, item := range recipient.Chats {
					if item.ID == sender.ID {
						status = true
						break
					}
				}
				if !status {
					recipient.Chats = append(recipient.Chats, sender)
					h.userRepo.Update(recipient)
				}
			}
		}
	}
	return nil
}

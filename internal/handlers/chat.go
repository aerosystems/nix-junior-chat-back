package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
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
}

// Chat godoc
// @Summary Chat [WebSocket]
// @Description Chat with users based on WebSocket
// @Tags chat
// @Param Authorization header string true "should contain Access Token, with the Bearer started"
// @Param chat body MessageResponseBody true "body should contain content and recipient_id for sending message"
// @Failure 401 {object} Response
// @Router /ws/chat [get]
func (h *BaseHandler) Chat(c echo.Context) error {
	sender, ok := c.Get("user").(*models.User)
	if !ok {
		err := errors.New("internal transport token error")
		c.Logger().Error(err)
	}
	c.Logger().Info(fmt.Sprintf("client %d connected", sender.ID))

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
				break
			}
			c.Logger().Error(err)
			break
		}

		for _, client := range clients {
			var responseMessage MessageResponseBody
			if err := json.Unmarshal(msg, &responseMessage); err != nil {
				reply := models.NewErrorMessage("invalid message format", sender.ID)
				client.WS.WriteMessage(websocket.TextMessage, reply.Json())
				c.Logger().Error(err)
				continue
			}

			recipient, err := h.userRepo.FindByID(responseMessage.RecipientID)
			if err != nil {
				reply := models.NewErrorMessage("invalid recipient id", sender.ID)
				client.WS.WriteMessage(websocket.TextMessage, reply.Json())
				c.Logger().Error(err)
				continue
			}
			message := models.NewTextMessage(responseMessage.Content, sender.ID, responseMessage.RecipientID)

			if client.User.ID == message.RecipientID {
				client.WS.WriteMessage(websocket.TextMessage, message.Json())
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

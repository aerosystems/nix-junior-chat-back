package ChatService

import (
	"encoding/json"
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
	"strconv"
	"time"
)

var connectedClients = make(map[int]*Client)

type msg struct {
	Content string `json:"content,omitempty"`
	ChatID  int    `json:"chatId,omitempty"`
	Err     string `json:"err,omitempty"`
}

func OnConnect(user *models.User, conn *websocket.Conn, rdb *redis.Client) error {
	fmt.Println("connected from:", conn.RemoteAddr(), "client:", user.Username, "id:", user.ID)

	u, err := Connect(rdb, user)
	if err != nil {
		return err
	}
	connectedClients[user.ID] = u
	return nil
}

func OnDisconnect(user *models.User, conn *websocket.Conn) chan struct{} {

	closeCh := make(chan struct{})

	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Println("connection closed for client", user.Username, "id:", user.ID)

		u := connectedClients[user.ID]
		if err := u.Disconnect(); err != nil {
			return err
		}
		delete(connectedClients, user.ID)
		close(closeCh)
		return nil
	})

	return closeCh
}

func OnClientMessage(conn *websocket.Conn, rdb *redis.Client, messageRepo models.MessageRepository, user *models.User) {

	var msg msg

	fmt.Println("message from:", conn.RemoteAddr(), "client:", user.Username, "id:", user.ID)

	if err := conn.ReadJSON(&msg); err != nil {
		HandleWSError(err, conn)
		return
	}

	newMessage := models.Message{
		Type:      "text",
		ChatID:    msg.ChatID,
		Content:   msg.Content,
		SenderID:  user.ID,
		Sender:    *user,
		Status:    "sent",
		CreatedAt: time.Now().Unix(),
	}
	newMessageJSON, err := json.Marshal(newMessage)
	if err != nil {
		HandleWSError(err, conn)
		return
	}

	channel := strconv.Itoa(msg.ChatID)

	if err := Chat(rdb, channel, string(newMessageJSON)); err != nil {
		HandleWSError(err, conn)
	}

	err = messageRepo.Create(&newMessage)
}

func OnChannelMessage(conn *websocket.Conn, messageRepo models.MessageRepository, user *models.User) {

	c := connectedClients[user.ID]

	go func() {
		for m := range c.MessageChan {

			var message models.Message
			if err := json.Unmarshal([]byte(m.Payload), &message); err != nil {
				HandleWSError(err, conn)
			}

			if message.SenderID != user.ID {
				message.Status = "delivered"
				err := messageRepo.Update(&message)
				if err != nil {
					fmt.Println(err)
				}
				if err := conn.WriteJSON(message); err != nil {
					fmt.Println(err)
				}
			}
		}
	}()
}

func HandleWSError(err error, conn *websocket.Conn) {
	_ = conn.WriteJSON(msg{Err: err.Error()})
}

package ChatService

import (
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
)

var connectedClients = make(map[int]*Client)

type msg struct {
	Content string `json:"content,omitempty"`
	Channel string `json:"channel,omitempty"`
	Command int    `json:"command,omitempty"`
	Err     string `json:"err,omitempty"`
}

const (
	commandSubscribe = iota
	commandUnsubscribe
	commandChat
)

func OnConnect(user *models.User, conn *websocket.Conn, rdb *redis.Client) error {
	fmt.Println("connected from:", conn.RemoteAddr(), "client:", user.Username, "id:", user.ID)

	u, err := Connect(rdb, user.ID)
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

func OnClientMessage(conn *websocket.Conn, user *models.User, rdb *redis.Client) {

	var msg msg

	fmt.Println("message from:", conn.RemoteAddr(), "client:", user.Username, "id:", user.ID)

	if err := conn.ReadJSON(&msg); err != nil {
		HandleWSError(err, conn)
		return
	}

	u := connectedClients[user.ID]

	switch msg.Command {
	case commandSubscribe:
		if err := u.Subscribe(rdb, msg.Channel); err != nil {
			HandleWSError(err, conn)
		}
	case commandUnsubscribe:
		if err := u.Unsubscribe(rdb, msg.Channel); err != nil {
			HandleWSError(err, conn)
		}
	case commandChat:
		if err := Chat(rdb, msg.Channel, msg.Content); err != nil {
			HandleWSError(err, conn)
		}
	}
}

func OnChannelMessage(conn *websocket.Conn, user *models.User) {

	c := connectedClients[user.ID]

	go func() {
		for m := range c.MessageChan {

			msg := msg{
				Content: m.Payload,
				Channel: m.Channel,
			}

			if err := conn.WriteJSON(msg); err != nil {
				fmt.Println(err)
			}
		}

	}()
}

func HandleWSError(err error, conn *websocket.Conn) {
	_ = conn.WriteJSON(msg{Err: err.Error()})
}

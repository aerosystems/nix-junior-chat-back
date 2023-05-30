package ChatService

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader websocket.Upgrader

var connectedClients = make(map[string]*Client)

func H(rdb *redis.Client, fn func(http.ResponseWriter, *http.Request, *redis.Client)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, rdb)
	}
}

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

func ChatWebSocketHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		handleWSError(err, conn)
		return
	}

	err = onConnect(r, conn, rdb)
	if err != nil {
		handleWSError(err, conn)
		return
	}

	closeCh := onDisconnect(r, conn, rdb)

	onChannelMessage(conn, r)

loop:
	for {
		select {
		case <-closeCh:
			break loop
		default:
			onClientMessage(conn, r, rdb)
		}
	}
}

func onConnect(r *http.Request, conn *websocket.Conn, rdb *redis.Client) error {
	clientname := r.URL.Query()["clientname"][0]
	fmt.Println("connected from:", conn.RemoteAddr(), "client:", clientname)

	u, err := Connect(rdb, clientname)
	if err != nil {
		return err
	}
	connectedClients[clientname] = u
	return nil
}

func onDisconnect(r *http.Request, conn *websocket.Conn, rdb *redis.Client) chan struct{} {

	closeCh := make(chan struct{})

	clientname := r.URL.Query()["clientname"][0]

	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Println("connection closed for client", clientname)

		u := connectedClients[clientname]
		if err := u.Disconnect(); err != nil {
			return err
		}
		delete(connectedClients, clientname)
		close(closeCh)
		return nil
	})

	return closeCh
}

func onClientMessage(conn *websocket.Conn, r *http.Request, rdb *redis.Client) {

	var msg msg

	if err := conn.ReadJSON(&msg); err != nil {
		handleWSError(err, conn)
		return
	}

	clientname := r.URL.Query()["clientname"][0]
	u := connectedClients[clientname]

	switch msg.Command {
	case commandSubscribe:
		if err := u.Subscribe(rdb, msg.Channel); err != nil {
			handleWSError(err, conn)
		}
	case commandUnsubscribe:
		if err := u.Unsubscribe(rdb, msg.Channel); err != nil {
			handleWSError(err, conn)
		}
	case commandChat:
		if err := Chat(rdb, msg.Channel, msg.Content); err != nil {
			handleWSError(err, conn)
		}
	}
}

func onChannelMessage(conn *websocket.Conn, r *http.Request) {

	clientname := r.URL.Query()["clientname"][0]
	u := connectedClients[clientname]

	go func() {
		for m := range u.MessageChan {

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

func handleWSError(err error, conn *websocket.Conn) {
	_ = conn.WriteJSON(msg{Err: err.Error()})
}

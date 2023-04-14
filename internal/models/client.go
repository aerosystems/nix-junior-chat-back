package models

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	WS   *websocket.Conn
	User User
}

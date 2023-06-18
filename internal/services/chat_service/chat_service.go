package ChatService

import (
	"encoding/json"
	firebase "firebase.google.com/go"
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"strconv"
	"time"
)

type ChatService struct {
	firebaseApp *firebase.App
	rdb         *redis.Client
	userRepo    models.UserRepository
	messageRepo models.MessageRepository
	chatRepo    models.ChatRepository
}

func NewChatService(firebaseApp *firebase.App,
	rdb *redis.Client,
	userRepo models.UserRepository,
	messageRepo models.MessageRepository,
	chatRepo models.ChatRepository,
) *ChatService {
	return &ChatService{
		firebaseApp: firebaseApp,
		rdb:         rdb,
		userRepo:    userRepo,
		messageRepo: messageRepo,
		chatRepo:    chatRepo,
	}
}

var connectedClients = make(map[int]*Client)

type msg struct {
	Content string `json:"content,omitempty"`
	ChatID  int    `json:"chatId,omitempty"`
	Type    string `json:"type,omitempty"` // error, system
	Err     string `json:"error,omitempty"`
}

func (cs *ChatService) OnConnect(conn *websocket.Conn, user *models.User) error {
	fmt.Println("connected from:", conn.RemoteAddr(), "client:", user.Username, "id:", user.ID)

	u, err := Connect(cs.rdb, user)
	if err != nil {
		return err
	}
	connectedClients[user.ID] = u
	return nil
}

func (cs *ChatService) OnDisconnect(conn *websocket.Conn, user *models.User) chan struct{} {

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

func (cs *ChatService) OnClientMessage(conn *websocket.Conn, user *models.User) {
	var msg msg

	if err := conn.ReadJSON(&msg); err != nil {
		cs.HandleWSError(err, "error sending message", conn)
		return
	}

	fmt.Println("message :", msg.Content, ", client:", user.Username, "id:", user.ID)

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
		cs.HandleWSError(err, "error sending message", conn)
		return
	}

	chat, err := cs.chatRepo.FindByID(msg.ChatID)
	if err != nil {
		cs.HandleWSError(err, "error sending message", conn)
		return
	}

	// Handle case when user is blocked
	if chat.Type == "private" {
		for _, u := range chat.Users {
			if u.ID != user.ID {
				for _, b := range u.BlockedUsers {
					if user.ID == b.ID {
						err = fmt.Errorf("user %d blocks %d", u.ID, user.ID)
						cs.HandleWSError(err, fmt.Sprintf("User %s blocked you. You can't send message to this chat", b.Username), conn)
						return
					}
				}
			}
		}
	}

	channel := strconv.Itoa(msg.ChatID)

	if err := Chat(cs.rdb, channel, string(newMessageJSON)); err != nil {
		cs.HandleWSError(err, "error sending message", conn)
	}

	fmt.Println("message: ", newMessage, "sent to channel:", channel)
	err = cs.messageRepo.Create(&newMessage)
}

func (cs *ChatService) OnChannelMessage(conn *websocket.Conn, user *models.User) {

	c := connectedClients[user.ID]

	go func() {
		for m := range c.MessageChan {
			var message models.Message
			if err := json.Unmarshal([]byte(m.Payload), &message); err != nil {
				cs.HandleWSError(err, "error sending message", conn)
			}

			if message.SenderID != user.ID {

				if err := conn.WriteJSON(message); err != nil {
					log.Println(err)
				} else {
					// Send notification to all users in chat
					cs.SendNotification(&message)
				}
			}
		}
	}()
}

func (cs *ChatService) HandleWSError(err error, message string, conn *websocket.Conn) {
	// ErrorResponse takes a response status code and arbitrary data and writes a json response to the client in depends on Header Accept and APP_ENV environment variable(has two possible values: dev and prod)
	// - APP_ENV=dev responds debug info level of error
	// - APP_ENV=prod responds just message about error [DEFAULT]

	payload := msg{
		Type:    "error",
		Content: message,
	}

	if os.Getenv("APP_ENV") == "dev" {
		payload.Err = err.Error()
	}

	_ = conn.WriteJSON(payload)
}

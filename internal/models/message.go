package models

import (
	"encoding/json"
	"time"
)

type Message struct {
	ID          int    `json:"id" gorm:"primaryKey" example:"1"`
	Type        string `json:"type" example:"message"` // text, error, system
	Content     string `json:"content" example:"bla-bla-bla"`
	SenderID    int    `json:"senderId" gorm:"foreignKey:SenderID" example:"2"`
	Sender      User   `json:"sender"`
	RecipientID int    `json:"recipientId" gorm:"foreignKey:RecipientID" example:"3"`
	Recipient   User   `json:"-"`
	CreatedAt   int64  `json:"createdAt" example:"1620000000"`
}

type MessageRepository interface {
	FindByID(id int) (*Message, error)
	FindAll() (*[]Message, error)
	Create(message *Message) error
	GetMessages(senderID, recipientID, from, limit int) (*[]Message, error)
}

func NewTextMessage(content string, sender User, recipientID int) *Message {
	return &Message{
		Type:        "text",
		Content:     content,
		Sender:      sender,
		RecipientID: recipientID,
		CreatedAt:   time.Now().Unix(),
	}
}

func NewErrorMessage(content string, sender User) *Message {
	return &Message{
		Type:        "error",
		Content:     content,
		Sender:      sender,
		RecipientID: sender.ID,
		CreatedAt:   time.Now().Unix(),
	}
}

func NewSystemMessage(content string, sender User) *Message {
	return &Message{
		Type:        "system",
		Content:     content,
		Sender:      sender,
		RecipientID: sender.ID,
		CreatedAt:   time.Now().Unix(),
	}
}

func (m *Message) Json() []byte {
	result, _ := json.Marshal(m)
	return result
}

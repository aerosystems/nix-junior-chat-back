package models

import (
	"encoding/json"
	"time"
)

type ResponseMessage struct {
	Content     string `json:"content" example:"bla-bla-bla"`
	RecipientID int    `json:"recipientId" example:"1"`
}

type Message struct {
	Type        string `json:"type" example:"message"` // text, error, system
	Content     string `json:"content" example:"bla-bla-bla"`
	SenderID    int    `json:"senderId" example:"1"`
	RecipientID int    `json:"recipientId" example:"2"`
	CreatedAt   int64  `json:"createdAt" example:"1620000000"`
}

func NewTextMessage(content string, senderID int, recipientID int) *Message {
	return &Message{
		Type:        "text",
		Content:     content,
		SenderID:    senderID,
		RecipientID: recipientID,
		CreatedAt:   time.Now().Unix(),
	}
}

func NewErrorMessage(content string, senderID int) *Message {
	return &Message{
		Type:        "error",
		Content:     content,
		SenderID:    senderID,
		RecipientID: senderID,
		CreatedAt:   time.Now().Unix(),
	}
}

func NewSystemMessage(content string, senderID int) *Message {
	return &Message{
		Type:        "system",
		Content:     content,
		SenderID:    senderID,
		RecipientID: senderID,
		CreatedAt:   time.Now().Unix(),
	}
}

func (m *Message) Json() []byte {
	result, _ := json.Marshal(m)
	return result
}

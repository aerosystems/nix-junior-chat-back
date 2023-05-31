package ChatService

import (
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"time"
)

func NewTextMessage(content string, sender models.User, recipientID int) *models.Message {
	return &models.Message{
		Type:        "text",
		Content:     content,
		Sender:      sender,
		RecipientID: recipientID,
		CreatedAt:   time.Now().Unix(),
	}
}

func NewErrorMessage(content string, sender models.User) *models.Message {
	return &models.Message{
		Type:        "error",
		Content:     content,
		Sender:      sender,
		RecipientID: sender.ID,
		CreatedAt:   time.Now().Unix(),
	}
}

func NewSystemMessage(content string, sender models.User) *models.Message {
	return &models.Message{
		Type:        "system",
		Content:     content,
		Sender:      sender,
		RecipientID: sender.ID,
		CreatedAt:   time.Now().Unix(),
	}
}

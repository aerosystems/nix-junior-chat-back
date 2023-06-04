package storage

import (
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"gorm.io/gorm"
)

type MessageRepo struct {
	db *gorm.DB
}

func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{
		db: db,
	}
}

func (r *MessageRepo) FindAll() (*[]models.Message, error) {
	var messages []models.Message
	r.db.Find(&messages)
	return &messages, nil
}

func (r *MessageRepo) FindByID(id int) (*models.Message, error) {
	var message models.Message
	result := r.db.Find(&message, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &message, nil
}

func (r *MessageRepo) Create(message *models.Message) error {
	result := r.db.Create(&message)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *MessageRepo) Update(message *models.Message) error {
	result := r.db.Save(&message)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *MessageRepo) GetMessages(chatID, from, limit int) (*[]models.Message, error) {
	var messages []models.Message
	if from == 0 {
		result := r.db.Preload("Sender").Where("chat_id = ?", chatID).
			Limit(limit).
			Order("id desc").
			Find(&messages)
		if result.Error != nil {
			return nil, result.Error
		}
	} else {
		result := r.db.Preload("Sender").Where("chat_id = ? AND id < ?", chatID, from).
			Limit(limit).
			Order("id desc").
			Find(&messages)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	return &messages, nil
}

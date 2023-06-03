package storage

import (
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"gorm.io/gorm"
)

type ChatRepo struct {
	db *gorm.DB
}

func NewChatRepo(db *gorm.DB) *ChatRepo {
	return &ChatRepo{
		db: db,
	}
}

func (r *ChatRepo) FindByID(id int) (*models.Chat, error) {
	var chat models.Chat
	result := r.db.Preload("Users").Find(&chat, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &chat, nil
}

func (r *ChatRepo) Create(chat *models.Chat) error {
	result := r.db.Create(&chat)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *ChatRepo) Update(chat *models.Chat) error {
	result := r.db.Save(&chat)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *ChatRepo) Delete(chat *models.Chat) error {
	result := r.db.Delete(&chat)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

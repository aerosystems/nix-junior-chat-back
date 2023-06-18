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
	result := r.db.Preload("Users.BlockedUsers").Find(&chat, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &chat, nil
}

func (r *ChatRepo) FindByUserID(id int) (*[]models.Chat, error) {
	var chats []models.Chat
	result := r.db.Preload("Users").Joins("JOIN chat_users ON chat_users.chat_id = chats.id").
		Where("chat_users.user_id = ?", id).
		Find(&chats)
	if result.Error != nil {
		return nil, result.Error
	}
	return &chats, nil
}

func (r *ChatRepo) FindPrivateChatByUsersArray(users []*models.User) (*models.Chat, error) {
	var chat models.Chat
	result := r.db.Preload("Users").
		Where("type = ?", "private").
		Joins("JOIN chat_users ON chat_users.chat_id = chats.id").
		Where("chat_users.user_id IN ?", getUserIDs(users)).
		Group("chats.id").
		Having("COUNT(DISTINCT chat_users.user_id) = ?", len(users)).
		First(&chat)

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
	result := r.db.Table("chat_users").Where("chat_id = ?", chat.ID).Delete(&models.User{})
	if result.Error != nil {
		return result.Error
	}

	result = r.db.Delete(&chat)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func getUserIDs(users []*models.User) []int {
	ids := make([]int, len(users))
	for i, user := range users {
		ids[i] = user.ID
	}
	return ids
}

package models

import (
	"time"
)

type User struct {
	ID            int       `json:"id" gorm:"primaryKey" example:"1"`
	Username      string    `json:"username" gorm:"unique" example:"username"`
	Password      string    `json:"-"`
	Image         string    `json:"image" example:"image.png"`
	FollowedUsers []*User   `json:"followedUsers,omitempty" gorm:"many2many:followed_users"`
	BlockedUsers  []*User   `json:"blockedUsers,omitempty" gorm:"many2many:blocked_users"`
	Chats         []*Chat   `json:"chats,omitempty" gorm:"many2many:chat_users"`
	Devices       []*Device `json:"devices,omitempty"`
	IsOnline      bool      `json:"isOnline,omitempty" gorm:"default:false" example:"true"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"updatedAt,omitempty" example:"2024-01-01T12:00:00.000Z"`
}

type UserRepository interface {
	FindAll() (*[]User, error)
	FindByID(id int) (*User, error)
	FindByUsername(username string) (*User, error)
	FindArrayByPartUsername(username string, order string, limit int) (*[]User, error)
	Create(user *User) error
	Update(user *User) error
	UpdateWithAssociations(user *User) error
	Delete(user *User) error
	ReplaceFollowedUsers(user *User, followedUsers []*User) error
	ReplaceBlockedUsers(user *User, blockedUsers []*User) error
	ResetPassword(user *User, password string) error
	PasswordMatches(user *User, plainText string) (bool, error)
}

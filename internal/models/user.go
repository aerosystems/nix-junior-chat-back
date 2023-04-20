package models

import (
	"gorm.io/gorm"
	"os"
	"time"
)

type User struct {
	ID            int            `json:"id" gorm:"primaryKey" example:"1"`
	Username      string         `json:"username" gorm:"unique" example:"username"`
	Password      string         `json:"-"`
	Image         string         `json:"image" example:"image.png"`
	FollowedUsers []*User        `json:"followedUsers" gorm:"many2many:followed_users"`
	BlockedUsers  []*User        `json:"blockedUsers" gorm:"many2many:blocked_users"`
	CreatedAt     time.Time      `json:"-"`
	UpdatedAt     time.Time      `json:"-"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

type UserRepository interface {
	FindAll() (*[]User, error)
	FindByID(id int) (*User, error)
	FindByUsername(username string) (*User, error)
	FindArrayByPartUsername(username string, order string, limit int) (*[]User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(user *User) error
	ResetPassword(user *User, password string) error
	PasswordMatches(user *User, plainText string) (bool, error)
}

func (u *User) ModifyImage() {
	if u.Image == "" {
		u.Image = "default.png"
	}
	if os.Getenv("URL_PREFIX_IMAGES") != "" {
		u.Image = os.Getenv("URL_PREFIX_IMAGES") + u.Image
	}
}

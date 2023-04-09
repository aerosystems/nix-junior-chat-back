package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	ID        int            `json:"id" gorm:"primaryKey" example:"1"`
	Username  string         `json:"username" gorm:"unique" example:"username"`
	Password  string         `json:"-"`
	Friends   []*User        `json:"-" gorm:"many2many:user_friends"`
	Blacklist []*User        `json:"-" gorm:"many2many:user_blacklist"`
	CreatedAt time.Time      `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time      `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type UserRepository interface {
	FindAll() (*[]User, error)
	FindByID(id int) (*User, error)
	FindByUsername(username string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(user *User) error
	ResetPassword(user *User, password string) error
	PasswordMatches(user *User, plainText string) (bool, error)
}

package storage

import (
	"errors"

	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/go-redis/redis/v7"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepo struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewUserRepo(db *gorm.DB, cache *redis.Client) *UserRepo {
	return &UserRepo{
		db:    db,
		cache: cache,
	}
}

func (r *UserRepo) FindAll() (*[]models.User, error) {
	var users []models.User
	r.db.Find(&users)
	return &users, nil
}

func (r *UserRepo) FindByID(id int) (*models.User, error) {
	var user models.User
	result := r.db.Preload("FollowedUsers").Preload("BlockedUsers").Preload("Chats.Users").Preload("Devices").Find(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepo) FindByUsername(username string) (*models.User, error) {
	var user models.User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepo) FindArrayByPartUsername(username string, order string, limit int) (*[]models.User, error) {
	var users []models.User
	r.db.Where("username LIKE ?", username+"%").Order("username " + order).Limit(limit).Find(&users)
	return &users, nil
}

func (r *UserRepo) Create(user *models.User) error {
	result := r.db.Create(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *UserRepo) Update(user *models.User) error {
	result := r.db.Save(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *UserRepo) UpdateWithAssociations(user *models.User) error {
	result := r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *UserRepo) ReplaceFollowedUsers(user *models.User, followedUsers []*models.User) error {
	err := r.db.Model(&user).Association("FollowedUsers").Replace(followedUsers)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepo) ReplaceBlockedUsers(user *models.User, blockedUsers []*models.User) error {
	err := r.db.Model(&user).Association("BlockedUsers").Replace(blockedUsers)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepo) Delete(user *models.User) error {
	result := r.db.Delete(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// ResetPassword is the method we will use to change a user's password.
func (r *UserRepo) ResetPassword(user *models.User, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	result := r.db.Save(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// PasswordMatches uses Go's bcrypt package to compare a user supplied password
// with the hash we have stored for a given user in the database. If the password
// and hash match, we return true; otherwise, we return false.
func (r *UserRepo) PasswordMatches(user *models.User, plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(plainText))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			// invalid password
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

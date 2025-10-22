package repository

import (
	"errors"
	"fmt"
	"sample/internal/app/ds"
	"sample/internal/app/serializer"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("not found")
)

// создает нового пользователя
func (r *Repository) RegisterUser(userJSON serializer.UserJSON) (ds.User, error) {
	// Проверяем существование пользователя
	existingUser, err := r.GetUserByUsername(userJSON.Username)
	if err == nil && existingUser.Username != "" {
		return ds.User{}, errors.New("пользователь c таким именем уже существует")
	}

	// Создаем пользователя
	user := serializer.UserFromJSON(userJSON)
	user.User_ID = uuid.New()

	if err := r.db.Create(&user).Error; err != nil {
		return ds.User{}, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	return user, nil
}

// аутентифицирует пользователя и возвращает JWT токен
func (r *Repository) LoginUser(userJSON serializer.UserJSON) (ds.User, error) {
	var user ds.User

	user, err := r.GetUserByUsername(userJSON.Username)
	if err != nil {
		return ds.User{}, ErrNotFound
	}
	if userJSON.Password != user.PasswordHash {
		return ds.User{}, ErrNotFound
	}
	return user, nil
}

// GetUserByID возвращает пользователя по ID
func (r *Repository) GetUserByID(userID uuid.UUID) (ds.User, error) {
	var user ds.User
	if err := r.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return ds.User{}, ErrNotFound
	}
	return user, nil
}

func (r *Repository) GetUserByUsername(username string) (ds.User, error) {
	var user ds.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, errors.New("user not found")
		}
		return ds.User{}, fmt.Errorf("ошибка базы данных: %w", err)
	}
	return user, nil
}

// UpdateUserInfo обновляет информацию о пользователе
func (r *Repository) UpdateUserInfo(userID uuid.UUID, userJSON serializer.UserJSON) (ds.User, error) {
	user := serializer.UserFromJSON(userJSON)

	if err := r.db.Model(&ds.User{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"username": user.Username,
	}).Error; err != nil {
		return ds.User{}, err
	}

	return r.GetUserByID(userID)
}

package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"sample/internal/app/ds"
)

// POST регистрация
func (r *Repository) RegisterUser(user *ds.User) error {
	// Проверяем существование пользователя
	var count int64
	err := r.db.Model(&ds.User{}).Where("username = ?", user.Username).Count(&count).Error
	if err != nil {
		return fmt.Errorf("ошибка проверки пользователя: %w", err)
	}

	if count > 0 {
		return errors.New("пользователь с таким именем уже существует")
	}

	var maxID uint
	err = r.db.Model(&ds.User{}).Select("COALESCE(MAX(user_id), 0)").Scan(&maxID).Error
	if err != nil {
		return fmt.Errorf("ошибка получения максимального ID: %w", err)
	}
	user.User_ID = maxID + 1

	return r.db.Create(user).Error
}

// GET полей пользователя после аутентификации (для личного кабинета)
func (r *Repository) GetUserProfile(userID uint) (ds.User, error) {
	var user ds.User
	err := r.db.Select("user_id, username, is_moderator").Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return ds.User{}, err
	}
	return user, nil
}

// PUT пользователя (личный кабинет)
func (r *Repository) UpdateUserProfile(userID uint, updates map[string]interface{}) error {
	// Запрещаем изменение системных полей
	delete(updates, "user_id")
	delete(updates, "is_moderator")

	// Разрешаем менять только определенные поля для личного кабинета
	allowedFields := []string{"username", "password_hash"}
	cleanedUpdates := make(map[string]interface{})

	for key, value := range updates {
		for _, allowed := range allowedFields {
			if key == allowed {
				cleanedUpdates[key] = value
				break
			}
		}
	}

	if len(cleanedUpdates) == 0 {
		return errors.New("нет разрешенных полей для изменения")
	}

	return r.db.Model(&ds.User{}).Where("user_id = ?", userID).Updates(cleanedUpdates).Error
}

// POST аутентификация
func (r *Repository) AuthenticateUser(username, passwordHash string) (ds.User, error) {
	var user ds.User
	err := r.db.Where("username = ? AND password_hash = ?", username, passwordHash).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, errors.New("неверное имя пользователя или пароль")
		}
		return ds.User{}, err
	}
	return user, nil
}

// POST деавторизация
func (r *Repository) LogoutUser(userID uint) error {
	// В stateless-архитектуре обычно просто удаляем токен на клиенте
	// Здесь можно добавить логику для blacklist токенов если используете JWT
	// Или обновить поле last_logout в БД

	// Пример: обновляем время последнего выхода
	// return r.db.Model(&ds.User{}).Where("user_id = ?", userID).Update("last_logout", time.Now()).Error

	// Для простоты возвращаем nil - основная логика на клиенте
	return nil
}

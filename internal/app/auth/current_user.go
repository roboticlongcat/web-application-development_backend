// internal/app/auth/current_user.go
package auth

import "sync"

type CurrentUser struct {
	ID          uint
	Username    string
	IsModerator bool
}

var (
	instance *CurrentUser
	once     sync.Once
)

// Singleton функция для фиксированного пользователя
func GetCurrentUser() *CurrentUser {
	once.Do(func() {
		instance = &CurrentUser{
			ID:          1, // Фиксированный ID создателя
			Username:    "creator",
			IsModerator: false,
		}
	})
	return instance
}

// Функция для получения модератора (для завершения расчетов)
func GetModeratorUser() *CurrentUser {
	return &CurrentUser{
		ID:          2, // Фиксированный ID модератора
		Username:    "moderator",
		IsModerator: true,
	}
}

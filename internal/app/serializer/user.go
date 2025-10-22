package serializer

import (
	"sample/internal/app/ds"

	"github.com/google/uuid"
)

// UserJSON - структура для сериализации пользователя
type UserJSON struct {
	User_ID     uuid.UUID `json:"user_id"`
	Username    string    `json:"username" binding:"required"`
	Password    string    `json:"password" binding:"required"`
	IsModerator bool      `json:"is_moderator"`
}

type UserRespJSON struct {
	User_ID     uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	IsModerator bool      `json:"is_moderator"`
}

type UserRespWithTokenJSON struct {
	Token string       `json: "token"`
	User  UserRespJSON `json: "user"`
}

// UserToRespJSON преобразует модель в JSON для ответа (без пароля)
func UserToRespJSON(user ds.User) UserRespJSON {
	return UserRespJSON{
		User_ID:     user.User_ID,
		Username:    user.Username,
		IsModerator: user.IsModerator,
	}
}

// UserFromJSON преобразует JSON в модель
func UserFromJSON(userJSON UserJSON) ds.User {
	return ds.User{
		User_ID:      userJSON.User_ID,
		Username:     userJSON.Username,
		PasswordHash: userJSON.Password, // В реальном приложении нужно хэшировать!
		IsModerator:  userJSON.IsModerator,
	}
}

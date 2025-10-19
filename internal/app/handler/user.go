package handler

import (
	"net/http"
	"strconv"

	"sample/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// POST /api/users/register - регистрация пользователя
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var user ds.User
	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных пользователя",
		})
		return
	}

	// Проверяем обязательные поля
	if user.Username == "" || user.PasswordHash == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Имя пользователя и пароль обязательны",
		})
		return
	}

	if err := h.Repository.RegisterUser(&user); err != nil {
		if err.Error() == "пользователь с таким именем уже существует" {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			logrus.Error(err)
		}
		return
	}

	// Не возвращаем пароль в ответе
	responseUser := ds.User{
		User_ID:     user.User_ID,
		Username:    user.Username,
		IsModerator: user.IsModerator,
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Пользователь успешно зарегистрирован",
		"user":    responseUser,
	})
}

// GET /api/users/:id/profile - получение профиля пользователя
func (h *Handler) GetUserProfile(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID пользователя",
		})
		return
	}

	user, err := h.Repository.GetUserProfile(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Пользователь не найден",
		})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// PUT /api/users/:id/profile - обновление профиля пользователя
func (h *Handler) UpdateUserProfile(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID пользователя",
		})
		return
	}

	var updates map[string]interface{}
	if err := ctx.BindJSON(&updates); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных для обновления",
		})
		return
	}

	if err := h.Repository.UpdateUserProfile(uint(id), updates); err != nil {
		if err.Error() == "нет разрешенных полей для изменения" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			logrus.Error(err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Профиль пользователя успешно обновлен",
	})
}

// POST /api/users/login - аутентификация пользователя
func (h *Handler) LoginUser(ctx *gin.Context) {
	var credentials struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.BindJSON(&credentials); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных для аутентификации",
		})
		return
	}

	// В реальном приложении здесь нужно хэшировать пароль
	// passwordHash := hashPassword(credentials.Password)
	passwordHash := credentials.Password // Для демонстрации используем plain text

	user, err := h.Repository.AuthenticateUser(credentials.Username, passwordHash)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// В реальном приложении здесь генерируется JWT токен
	// token, err := generateJWTToken(user)

	// Не возвращаем пароль в ответе
	responseUser := ds.User{
		User_ID:     user.User_ID,
		Username:    user.Username,
		IsModerator: user.IsModerator,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Аутентификация успешна",
		"user":    responseUser,
		// "token":   token, // В реальном приложении
	})
}

// POST /api/users/logout - деавторизация пользователя
func (h *Handler) LogoutUser(ctx *gin.Context) {
	var request struct {
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных",
		})
		return
	}

	if err := h.Repository.LogoutUser(request.UserID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Выход выполнен успешно",
	})
}

// GET /api/users/me - получение текущего пользователя (для демонстрации)
func (h *Handler) GetCurrentUser(ctx *gin.Context) {
	// В реальном приложении userID извлекается из JWT токена
	// userID := ctx.GetUint("user_id")

	// Для демонстрации используем фиксированный ID
	userID := uint(1)

	user, err := h.Repository.GetUserProfile(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Пользователь не найден",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"User_ID": user.User_ID, "Username": user.Username, "IsModerator": user.IsModerator})
}

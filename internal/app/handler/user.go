package handler

import (
	"errors"
	"net/http"
	"sample/internal/app/ds"
	"sample/internal/app/repository"
	"sample/internal/app/serializer"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// RegisterUser godoc
// @Summary Регистрация пользователя
// @Description Регистрирует нового пользователя. Возвращает URL созданного ресурса в Location и тело созданного пользователя.
// @Tags users
// @Accept json
// @Produce json
// @Param user body serializer.UserJSON true "Параметры нового пользователя"
// @Success 201 {object} serializer.UserRespJSON "Пользователь создан"
// @Failure 400 {object} map[string]string "Ошибка валидации или входных данных"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/users/register [post]
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var userJSON serializer.UserJSON
	if err := ctx.BindJSON(&userJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.RegisterUser(userJSON)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, serializer.UserToRespJSON(user))
}

// LoginUser godoc
// @Summary Вход (получение токена)
// @Description Принимает логин/пароль, возвращает jwt-токен и данные пользователя.
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body serializer.UserJSON true "Логин и пароль"
// @Success 200 {object} serializer.UserRespWithTokenJSON "Ответ с токеном и пользователем"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/users/login [post]
func (h *Handler) LoginUser(ctx *gin.Context) {
	var userJSON serializer.UserJSON
	if err := ctx.BindJSON(&userJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.LoginUser(userJSON)
	if err == repository.ErrNotFound {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	claims := ds.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.JWTConfig.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:      user.User_ID,
		Username:    user.Username,
		IsModerator: user.IsModerator,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.JWTConfig.Secret))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.UserRespWithTokenJSON{
		Token: tokenString,
		User:  serializer.UserToRespJSON(user),
	})
}

// UpdateUserInfo godoc
// @Summary Изменить профиль пользователя
// @Description Обновляет профиль пользователя (может делать только сам пользователь).
// @Tags users
// @Accept json
// @Produce json
// @Param user body serializer.UserJSON true "Новые данные профиля"
// @Success 200 {object} serializer.UserRespJSON "Обновлённый профиль"
// @Failure 400 {object} map[string]string "Ошибка запроса или авторизации"
// @Failure 403 {object} map[string]string "Доступ запрещён"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/users/profile [put]
func (h *Handler) UpdateUserInfo(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var userJSON serializer.UserJSON
	if err := ctx.BindJSON(&userJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByID(userID)
	if err == repository.ErrNotFound {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if user.User_ID != userID {
		h.errorHandler(ctx, http.StatusForbidden, errors.New("access denied"))
		return
	}

	_, err = h.Repository.UpdateUserInfo(userID, userJSON)
	if err == repository.ErrNotFound {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Изменения внесены",
	})
}

// LogoutUser godoc
// @Summary Выход (удаление токена)
// @Description Удаляет токен текущего пользователя из хранилища. Возвращает {"status":"signed_out"}.
// @Tags users
// @Produce json
// @Success 200 {object} map[string]string "status"
// @Failure 400 {object} map[string]string "Проблема с получением user_id"
// @Failure 500 {object} map[string]string "Внутренняя ошибка при удалении токена"
// @Security ApiKeyAuth
// @Router /api/users/logout [post]
func (h *Handler) LogoutUser(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("invalid header"))
		return
	}
	tokenStr := authHeader[len("Bearer "):]

	err := h.Redis.WriteJWTToBlacklist(ctx.Request.Context(), tokenStr, h.JWTConfig.ExpiresIn)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Деавторизация прошла успешно",
	})
}

// / GetCurrentUser godoc
// @Summary Получить текущего пользователя
// @Description Возвращает данные текущего аутентифицированного пользователя
// @Tags users
// @Produce json
// @Success 200 {object} serializer.UserRespJSON "Данные пользователя"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Security ApiKeyAuth
// @Router /api/users/profile [get]
func (h *Handler) GetCurrentUser(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.UserToRespJSON(user))
}

// getUserID извлекает user_id из контекста (устанавливается в middleware)
func getUserID(ctx *gin.Context) (uuid.UUID, error) {
	userID, exists := ctx.Get(userCtx)
	if !exists {
		return uuid.Nil, errors.New("user_id not found")
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid user_id type")
	}

	return userIDUUID, nil
}

package handler

import (
	"sample/internal/app/config"
	"sample/internal/app/redis"
	"sample/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
	JWTConfig  *config.JWTConfig
}

func NewHandler(r *repository.Repository, redis *redis.Client, jwtConfig *config.JWTConfig) *Handler {
	return &Handler{
		Repository: r,
		Redis:      redis,
		JWTConfig:  jwtConfig,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	// Public routes - доступны без авторизации
	public := router.Group("/api")
	{
		public.GET("/patients", h.GetPatients)
		public.GET("/patients/:patient_id", h.GetPatient)
		public.POST("/users/register", h.RegisterUser)
		public.POST("/users/login", h.LoginUser)
	}

	// User routes - требуют JWT авторизации
	user := router.Group("/api")
	user.Use(h.AuthMiddleware) // только обычные пользователи
	{
		user.POST("/patients", h.CreatePatient)
		user.PUT("/patients/:patient_id", h.UpdatePatient)
		user.DELETE("/patients/:patient_id", h.DeletePatient)
		user.POST("/patients/:patient_id/photo", h.UploadPatientPhoto)
		user.POST("/patients/:patient_id/add", h.AddPatientToInsulinCalculation)

		// Расчеты инсулина (пользовательские операции)
		user.GET("/insulin-calculations/info", h.GetInsulinCalculationCartInfo)
		user.GET("/insulin-calculations", h.GetInsulinCalculations)
		user.GET("/insulin-calculations/:insulin_calculation_id", h.GetInsulinCalculationWithPatients)
		user.PUT("/insulin-calculations/:insulin_calculation_id", h.UpdateInsulinCalculation)
		user.PUT("/insulin-calculations/:insulin_calculation_id/form", h.FormInsulinCalculation)
		user.DELETE("/insulin-calculations/:insulin_calculation_id", h.DeleteInsulinCalculation)

		// Расчеты-пациенты (м-м)
		user.DELETE("/insulin-calculations/:insulin_calculation_id/patients/:patient_id", h.RemovePatientFromInsulinCalculation)
		user.PUT("/insulin-calculations/:insulin_calculation_id/patients/:patient_id", h.UpdatePatientInInsulinCalculation)

		// Пользователи
		user.POST("/users/logout", h.LogoutUser)
		user.GET("/users/profile", h.GetCurrentUser)
		user.PUT("/users/profile", h.UpdateUserInfo)
	}

	// Moderator routes - требуют JWT и прав модератора
	moderator := router.Group("/api")
	moderator.Use(h.AuthMiddleware, h.ModeratorMiddleware)
	{
		// Расчеты инсулина (модераторские операции)
		moderator.PUT("/insulin-calculations/:insulin_calculation_id/complete", h.CompleteInsulinCalculation)
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("static/styles", "./resources/styles")
	router.Static("/img", "./resources/img")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"error": err.Error(),
	})
}

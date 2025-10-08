package handler

import (
	"sample/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api")
	{
		// Пациенты
		api.GET("/patients", h.GetPatients)
		api.GET("/patients/:patient_id", h.GetPatient)
		api.POST("/patients", h.CreatePatient)
		api.PUT("/patients/:patient_id", h.UpdatePatient)
		api.DELETE("/patients/:patient_id", h.DeletePatient)
		api.POST("/patients/:patient_id/photo", h.UploadPatientPhoto)
		api.POST("/patients/:patient_id/add", h.AddPatientToInsulinCalculation)

		// Расчеты инсулина
		api.GET("/insulin-calculations/info", h.GetInsulinCalculationCartInfo)
		api.GET("/insulin-calculations", h.GetInsulinCalculations)
		api.GET("/insulin-calculations/:insulin_calculation_id", h.GetInsulinCalculationWithPatients)
		api.PUT("/insulin-calculations/:insulin_calculation_id", h.UpdateInsulinCalculation)
		api.PUT("/insulin-calculations/:insulin_calculation_id/form", h.FormInsulinCalculation)
		api.PUT("/insulin-calculations/:insulin_calculation_id/complete", h.CompleteInsulinCalculation)
		api.DELETE("/insulin-calculations/:insulin_calculation_id", h.DeleteInsulinCalculation)

		// Расчеты-пациенты (м-м)
		api.DELETE("/insulin-calculations/:insulin_calculation_id/patients/:patient_id", h.RemovePatientFromInsulinCalculation)
		api.PUT("/insulin-calculations/:insulin_calculation_id/patients/:patient_id", h.UpdatePatientInInsulinCalculation)
		api.PUT("/insulin-calculations/:insulin_calculation_id/patients/:patient_id/bread-units", h.UpdateBreadUnits)
		api.PUT("/insulin-calculations/:insulin_calculation_id/patients/:patient_id/glucose", h.UpdateCurrentGlucose)
		api.GET("/insulin-calculations/:insulin_calculation_id/patients/:patient_id", h.GetPatientInCalculation)

		// Пользователи
		api.POST("/users/register", h.RegisterUser)
		api.POST("/users/login", h.LoginUser)
		api.POST("/users/logout", h.LogoutUser)
		api.GET("/users/me", h.GetCurrentUser)
		api.GET("/users/:id/profile", h.GetUserProfile)
		api.PUT("/users/profile", h.UpdateUserProfile)
	}

	// Старые роуты для обратной совместимости (можно удалить после перехода на API)
	//router.GET("/", h.GetPatients)
	//router.GET("/patient/:id", h.GetPatient)
	//router.GET("/insulin_calculation/:id", h.GetInsulinCalculation)
	//router.POST("/insulin_calculation/add", h.AddPatientToInsulinCalculation)
	//router.POST("/insulin_calculation/delete", h.DeleteInsulinCalculation)
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

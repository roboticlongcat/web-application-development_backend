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
	router.GET("/", h.GetPatients)
	router.GET("/patient/:id", h.GetPatient)
	router.GET("/insulin_calculation/:id", h.GetInsulinCalculation)
	router.POST("/insulin_calculation/add", h.AddPatientToInsulinCalculation)
	router.POST("/insulin_calculation/delete", h.DeleteInsulinCalculation)
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

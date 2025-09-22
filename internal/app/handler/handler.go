package handler

import (
	"net/http"
	"sample/internal/app/repository"
	"strconv"

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

func (h *Handler) GetPatients(ctx *gin.Context) {
	var patients []repository.Patient
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		patients, err = h.Repository.GetPatients()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		patients, err = h.Repository.GetPatientsByName(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	count, _ := h.Repository.GetCalculationItemsCount(1)

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"patients":      patients,
		"query":         searchQuery,
		"count":         count,
		"calculationID": 1,
	})
}

func (h *Handler) GetPatient(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	patient, err := h.Repository.GetPatient(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "patient_detail.html", gin.H{
		"patient": patient,
	})
}

func (h *Handler) GetCalculation(ctx *gin.Context) {
	IDStr := ctx.Param("id")
	calculationID, err := strconv.Atoi(IDStr)

	if err != nil {
		logrus.Error("Неверный ID расчета:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID расчета"})
		return
	}

	var patients []repository.Patient

	patients, err = h.Repository.GetCalculation(calculationID)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "calculation.html", gin.H{
		"patients": patients,
	})
}

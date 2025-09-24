package handler

import (
	"net/http"
	"strconv"
	"strings"

	"sample/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetPatients(ctx *gin.Context) {
	var patients []ds.Patient
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		patients, err = h.Repository.GetPatients()
	} else {
		patients, err = h.Repository.GetPatientsByName(searchQuery)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	hasActiveCalculation := h.Repository.HasActiveCalculation()
	activeCalculationID := h.Repository.GetActiveCalculationID()

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"patients":             patients,
		"query":                searchQuery,
		"count":                h.Repository.GetCalculationCount(),
		"hasActiveCalculation": hasActiveCalculation,
		"calculationID":        activeCalculationID,
	})
}

func (h *Handler) GetPatient(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	patient, err := h.Repository.GetPatient(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "patient_detail.html", gin.H{
		"patient": patient,
	})
}

func (h *Handler) GetCalculation(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	isDraft, err := h.Repository.IsDraftCalculation(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if !isDraft {
		ctx.Redirect(http.StatusSeeOther, ctx.Request.Referer())
		return
	}

	calculationPatients, err := h.Repository.GetCalculation(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "calculation.html", gin.H{
		"patients":      calculationPatients,
		"calculationID": id,
	})
}

func (h *Handler) AddPatientToCalculation(ctx *gin.Context) {
	patientIDStr := ctx.PostForm("patient_id")
	patientID, err := strconv.Atoi(patientIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID пациента"})
		return
	}

	creatorID := 1

	currentGlucose := float32(8.0) // пример значения
	breadUnits := float32(2.0)     // пример значения

	err = h.Repository.AddPatientToCalculation(uint(patientID), uint(creatorID), float32(currentGlucose), float32(breadUnits))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/")
}

func (h *Handler) DeleteCalculation(ctx *gin.Context) {
	calculationIDStr := ctx.PostForm("calculation_id")
	calculationID, err := strconv.Atoi(calculationIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	//calculationID := 1

	err = h.Repository.DeleteCalculation(uint(calculationID))
	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return
	}

	// после вызова сразу произойдет обновление страницы
	ctx.Redirect(http.StatusFound, "/")
}

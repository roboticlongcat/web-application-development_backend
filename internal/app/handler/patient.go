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

	patientSearchQuery := ctx.Query("query")
	if patientSearchQuery == "" {
		patients, err = h.Repository.GetPatients()
	} else {
		patients, err = h.Repository.GetPatientsByName(patientSearchQuery)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}
	activeInsulinCalculationID := h.Repository.GetActiveInsulinCalculationID()
	var hasActiveInsulinCalculation bool
	var patientCount int

	if activeInsulinCalculationID != 0 {
		patientCount, err = h.Repository.GetInsulinCalculationItemsCount(activeInsulinCalculationID)
		if err != nil {
			logrus.Error("Error getting insulin calculation items count:", err)
		}
		hasActiveInsulinCalculation = patientCount > 0
	} else {
		hasActiveInsulinCalculation = false
		patientCount = 0
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"patients":                    patients,
		"query":                       patientSearchQuery,
		"count":                       patientCount,
		"hasActiveInsulinCalculation": hasActiveInsulinCalculation,
		"insulin_calculation_ID":      activeInsulinCalculationID,
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

	patient, err := h.Repository.GetPatient(uint(id))
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

func (h *Handler) GetInsulinCalculation(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	isDraft, err := h.Repository.IsDraftInsulinCalculation(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if !isDraft {
		ctx.Redirect(http.StatusSeeOther, ctx.Request.Referer())
		return
	}

	insulincalculationPatients, err := h.Repository.GetInsulinCalculation(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "insulin_calculation.html", gin.H{
		"patients":               insulincalculationPatients,
		"insulin_calculation_ID": id,
	})
}

func (h *Handler) AddPatientToInsulinCalculation(ctx *gin.Context) {
	patientIDStr := ctx.PostForm("patient_id")
	patientID, err := strconv.Atoi(patientIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID пациента"})
		return
	}

	creatorID := 1

	currentGlucose := float32(8.0)
	breadUnits := float32(2.0)

	err = h.Repository.AddPatientToInsulinCalculation(uint(patientID), uint(creatorID), float32(currentGlucose), float32(breadUnits))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/")
}

func (h *Handler) DeleteInsulinCalculation(ctx *gin.Context) {
	insulincalculationIDStr := ctx.PostForm("insulin_calculation_id")
	insulincalculationID, err := strconv.Atoi(insulincalculationIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.DeleteInsulinCalculation(uint(insulincalculationID))
	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}

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

	patientSearchQuery := ctx.Query("query")
	if patientSearchQuery == "" {
		patients, err = h.Repository.GetPatients()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		patients, err = h.Repository.GetPatientsByName(patientSearchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	patients_count, _ := h.Repository.GetInsulinCalculationItemsCount(1)

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"patients":             patients,
		"query":                patientSearchQuery,
		"patients_count":       patients_count,
		"insulinCalculationID": 1,
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

func (h *Handler) GetInsulinCalculation(ctx *gin.Context) {
	IDStr := ctx.Param("id")
	calculationID, err := strconv.Atoi(IDStr)

	if err != nil {
		logrus.Error("Неверный ID расчета инсулина:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID расчета инсулина"})
		return
	}

	insulinCalculation, err := h.Repository.GetInsulinCalculation(calculationID)
	if err != nil {
		logrus.Error(err)
	}

	var patients []repository.Patient
	var IDPatients [5]int = [5]int{1, 5}

	for _, IDpatient := range insulinCalculation.Items {
		for _, number := range IDPatients {
			if number == IDpatient.PatientID {
				patient, _ := h.Repository.GetPatient(number)
				patients = append(patients, patient)
			}
		}
	}

	ctx.HTML(http.StatusOK, "insulin_calculation.html", gin.H{
		"patients":           patients,
		"insulinCalculation": insulinCalculation,
	})
}

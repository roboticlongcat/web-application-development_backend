package handler

import (
	"net/http"
	"strconv"

	"sample/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// DELETE /api/insulin-calculations/:insulin_calculation_id/patients/:patient_id - удаление пациента из расчета
func (h *Handler) RemovePatientFromInsulinCalculation(ctx *gin.Context) {
	insulinCalculationIDStr := ctx.Param("insulin_calculation_id")
	insulinCalculationID, err := strconv.Atoi(insulinCalculationIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID расчета",
		})
		return
	}

	patientIDStr := ctx.Param("patient_id")
	patientID, err := strconv.Atoi(patientIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID пациента",
		})
		return
	}

	if err := h.Repository.RemovePatientFromInsulinCalculation(uint(insulinCalculationID), uint(patientID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Пациент успешно удален из расчета",
	})
}

// PUT /api/insulin-calculations/:insulin_calculation_id/patients/:patient_id - обновление данных пациента в расчете
func (h *Handler) UpdatePatientInInsulinCalculation(ctx *gin.Context) {
	insulinCalculationIDStr := ctx.Param("insulin_calculation_id")
	insulinCalculationID, err := strconv.Atoi(insulinCalculationIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID расчета",
		})
		return
	}

	patientIDStr := ctx.Param("patient_id")
	patientID, err := strconv.Atoi(patientIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID пациента",
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

	if err := h.Repository.UpdatePatientInInsulinCalculation(uint(insulinCalculationID), uint(patientID), updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Данные пациента в расчете успешно обновлены",
	})
}

// PUT /api/insulin-calculations/:insulin_calculation_id/patients/:patient_id/bread-units - изменение хлебных единиц
func (h *Handler) UpdateBreadUnits(ctx *gin.Context) {
	insulinCalculationIDStr := ctx.Param("insulin_calculation_id")
	insulinCalculationID, err := strconv.Atoi(insulinCalculationIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID расчета",
		})
		return
	}

	patientIDStr := ctx.Param("patient_id")
	patientID, err := strconv.Atoi(patientIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID пациента",
		})
		return
	}

	var request struct {
		BreadUnits float32 `json:"bread_units" binding:"required"`
	}

	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных",
		})
		return
	}

	if err := h.Repository.UpdateBreadUnits(uint(insulinCalculationID), uint(patientID), request.BreadUnits); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Хлебные единицы успешно обновлены",
	})
}

// PUT /api/insulin-calculations/:insulin_calculation_id/patients/:patient_id/glucose - изменение текущей глюкозы
func (h *Handler) UpdateCurrentGlucose(ctx *gin.Context) {
	insulinCalculationIDStr := ctx.Param("insulin_calculation_id")
	insulinCalculationID, err := strconv.Atoi(insulinCalculationIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID расчета",
		})
		return
	}

	patientIDStr := ctx.Param("patient_id")
	patientID, err := strconv.Atoi(patientIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID пациента",
		})
		return
	}

	var request struct {
		CurrentGlucose float32 `json:"current_glucose" binding:"required"`
	}

	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных",
		})
		return
	}

	if err := h.Repository.UpdateCurrentGlucose(uint(insulinCalculationID), uint(patientID), request.CurrentGlucose); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Текущая глюкоза успешно обновлена",
	})
}

// GET /api/insulin-calculations/:insulin_calculation_id/patients/:patient_id - получение данных пациента в расчете
func (h *Handler) GetPatientInCalculation(ctx *gin.Context) {
	insulinCalculationIDStr := ctx.Param("insulin_calculation_id")
	insulinCalculationID, err := strconv.Atoi(insulinCalculationIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID расчета",
		})
		return
	}

	patientIDStr := ctx.Param("patient_id")
	patientID, err := strconv.Atoi(patientIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID пациента",
		})
		return
	}

	// Получаем все пациенты расчета и находим нужного
	patients, err := h.Repository.GetInsulinCalculation(uint(insulinCalculationID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	var patientInCalculation *ds.InsulinCalculationPatients
	for _, p := range patients {
		if p.Patient_ID == uint(patientID) {
			patientInCalculation = &p
			break
		}
	}

	if patientInCalculation == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Пациент не найден в расчете",
		})
		return
	}

	ctx.JSON(http.StatusOK, patientInCalculation)
}

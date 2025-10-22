package handler

import (
	"net/http"
	"sample/internal/app/serializer"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RemovePatientFromInsulinCalculation godoc
// @Summary Удалить пациента из расчета инсулина
// @Description Удаляет пациента из расчета инсулина
// @Tags insulin-calculations
// @Accept json
// @Produce json
// @Param insulin_calculation_id path int true "ID расчета инсулина"
// @Param patient_id path int true "ID пациента"
// @Success 200 {object} map[string]string "Пациент успешно удален из расчета"
// @Failure 400 {object} map[string]string "Неверные ID"
// @Failure 404 {object} map[string]string "Расчет или пациент не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/insulin-calculations/{insulin_calculation_id}/patients/{patient_id} [delete]
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

// UpdatePatientInInsulinCalculation godoc
// @Summary Обновить данные пациента в расчете инсулина
// @Description Обновляет данные пациента в расчете инсулина и пересчитывает инсулин
// @Tags insulin-calculations
// @Accept json
// @Produce json
// @Param insulin_calculation_id path int true "ID расчета инсулина"
// @Param patient_id path int true "ID пациента"
// @Param data body serializer.UpdatePatientInCalculationRequest true "Данные для обновления"
// @Success 200 {object} map[string]string "Данные пациента в расчете успешно обновлены"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 404 {object} map[string]string "Расчет или пациент не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/insulin-calculations/{insulin_calculation_id}/patients/{patient_id} [put]
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

	var request serializer.UpdatePatientInCalculationRequest
	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных для обновления",
		})
		return
	}
	updates := map[string]interface{}{
		"current_glucose": request.CurrentGlucose,
		"bread_units":     request.BreadUnits,
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

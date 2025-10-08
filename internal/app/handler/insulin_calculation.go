package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GET /api/insulin-calculations/info - иконка корзины
func (h *Handler) GetInsulinCalculationCartInfo(ctx *gin.Context) {
	insulinCalculationID, count, err := h.Repository.GetInsulinCalculationInfo()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	if insulinCalculationID == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "нет черновика",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"insulin_calculation_id": insulinCalculationID,
		"patient_count":          count,
	})
}

// GET /api/insulin-calculations - список расчетов с фильтрацией
func (h *Handler) GetInsulinCalculations(ctx *gin.Context) {
	// Получаем параметры фильтрации
	status := ctx.Query("status")
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")

	filters := map[string]interface{}{}
	if status != "" {
		filters["status"] = status
	}
	if startDate != "" {
		filters["start_date"] = startDate
	}
	if endDate != "" {
		filters["end_date"] = endDate
	}

	insulinCalculations, err := h.Repository.GetInsulinCalculations(filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"insulin_calculations": insulinCalculations,
		"count":                len(insulinCalculations),
	})
}

// GET /api/insulin-calculations/:id - получение расчета с пациентами
func (h *Handler) GetInsulinCalculationWithPatients(ctx *gin.Context) {
	idStr := ctx.Param("insulin_calculation_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID расчета",
		})
		return
	}

	insulinCalculation, patients, err := h.Repository.GetInsulinCalculationWithPatients(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Расчет не найден",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"insulin_calculation": insulinCalculation,
		"patients":            patients,
	})
}

// PUT /api/insulin-calculations/:id - обновление полей расчета
func (h *Handler) UpdateInsulinCalculation(ctx *gin.Context) {
	idStr := ctx.Param("insulin_calculation_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID расчета",
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

	if err := h.Repository.UpdateInsulinCalculation(uint(id), updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Расчет успешно обновлен",
	})
}

// PUT /api/insulin-calculations/:id/form - формирование расчета создателем
func (h *Handler) FormInsulinCalculation(ctx *gin.Context) {
	idStr := ctx.Param("insulin_calculation_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID расчета",
		})
		return
	}

	if err := h.Repository.FormInsulinCalculation(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Расчет успешно сформирован",
	})
}

// PUT /api/insulin-calculations/:id/complete - завершение/отклонение модератором
func (h *Handler) CompleteInsulinCalculation(ctx *gin.Context) {
	idStr := ctx.Param("insulin_calculation_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID расчета",
		})
		return
	}

	var request struct {
		Status string `json:"status" binding:"required"`
	}

	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных",
		})
		return
	}

	if err := h.Repository.CompleteInsulinCalculation(uint(id), request.Status); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Расчет успешно завершен",
	})
}

// DELETE /api/insulin-calculations/:id - удаление расчета
func (h *Handler) DeleteInsulinCalculation(ctx *gin.Context) {
	idStr := ctx.Param("insulin_calculation_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID расчета",
		})
		return
	}

	if err := h.Repository.DeleteInsulinCalculation(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Расчет успешно удален",
	})
}

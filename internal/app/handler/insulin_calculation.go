package handler

import (
	"net/http"
	"sample/internal/app/serializer"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetInsulinCalculationCartInfo godoc
// @Summary Получить информацию о корзине расчета
// @Description Возвращает информацию о текущем черновике расчета (ID и количество пациентов)
// @Tags insulin-calculations
// @Accept json
// @Produce json
// @Success 200 {object} serializer.InsulinCalculationCartInfoRespJSON "Информация о корзине"
// @Failure 404 {object} map[string]string "Черновик не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /api/insulin-calculations/info [get]
func (h *Handler) GetInsulinCalculationCartInfo(ctx *gin.Context) {
	creatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	insulinCalculationID, count, err := h.Repository.GetInsulinCalculationInfo(creatorID)
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

// GetInsulinCalculations godoc
// @Summary Получить список расчетов инсулина
// @Description Возвращает список расчетов с фильтрацией по статусу и дате
// @Tags insulin-calculations
// @Accept json
// @Produce json
// @Param status query string false "Фильтр по статусу"
// @Param start_date query string false "Начальная дата фильтрации"
// @Param end_date query string false "Конечная дата фильтрации"
// @Success 200 {array} serializer.InsulinCalculationRespJSON "Список расчетов"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /api/insulin-calculations [get]
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

	userID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	insulinCalculations, err := h.Repository.GetInsulinCalculations(filters, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	var response []serializer.InsulinCalculationRespJSON
	for _, calc := range insulinCalculations {
		response = append(response, serializer.InsulinCalculationToRespJSON(calc))
	}

	ctx.JSON(http.StatusOK, response)
}

// GetInsulinCalculationWithPatients godoc
// @Summary Получить расчет с пациентами
// @Description Возвращает полную информацию о расчете включая список пациентов
// @Tags insulin-calculations
// @Accept json
// @Produce json
// @Param insulin_calculation_id path int true "ID расчета инсулина"
// @Success 200 {object} serializer.InsulinCalculationWithPatientsRespJSON "Данные расчета с пациентами"
// @Failure 400 {object} map[string]string "Неверный ID расчета"
// @Failure 404 {object} map[string]string "Расчет не найден"
// @Security ApiKeyAuth
// @Router /api/insulin-calculations/{insulin_calculation_id} [get]
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

	// Преобразуем пациентов в структуру ответа
	var patientsResp []serializer.PatientInCalculationRespJSON
	for _, patient := range patients {
		patientsResp = append(patientsResp, serializer.PatientInCalculationRespJSON{
			Name:              patient["name"].(string),
			Sensitivity:       patient["sensitivity"].(float32),
			CurrentGlucose:    patient["current_glucose"].(float32),
			BreadUnits:        patient["bread_units"].(float32),
			CalculatedInsulin: patient["calculated_insulin"].(float32),
		})
	}

	response := serializer.InsulinCalculationWithPatientsRespJSON{
		InsulinCalculation: serializer.InsulinCalculationToRespJSON(insulinCalculation),
		Patients:           patientsResp,
		PatientCount:       len(patientsResp),
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateInsulinCalculation godoc
// @Summary Обновить данные расчета
// @Description Обновляет данные расчета инсулина
// @Tags insulin-calculations
// @Accept json
// @Produce json
// @Param insulin_calculation_id path int true "ID расчета инсулина"
// @Param data body serializer.InsulinCalculationJSON true "Данные для обновления"
// @Success 200 {object} map[string]string "Расчет успешно обновлен"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 404 {object} map[string]string "Расчет не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/insulin-calculations/{insulin_calculation_id} [put]
func (h *Handler) UpdateInsulinCalculation(ctx *gin.Context) {
	idStr := ctx.Param("insulin_calculation_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID расчета",
		})
		return
	}

	var calculationJSON serializer.InsulinCalculationJSON
	if err := ctx.BindJSON(&calculationJSON); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат для обновления данных",
		})
		return
	}

	updates := map[string]interface{}{
		"comment": calculationJSON.Comment,
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

// FormInsulinCalculation godoc
// @Summary Сформировать расчет
// @Description Переводит расчет из статуса черновика в сформированный
// @Tags insulin-calculations
// @Accept json
// @Produce json
// @Param insulin_calculation_id path int true "ID расчета инсулина"
// @Success 200 {object} map[string]string "Расчет успешно сформирован"
// @Failure 400 {object} map[string]string "Ошибка формирования"
// @Failure 404 {object} map[string]string "Расчет не найден"
// @Security ApiKeyAuth
// @Router /api/insulin-calculations/{insulin_calculation_id}/form [put]
func (h *Handler) FormInsulinCalculation(ctx *gin.Context) {
	idStr := ctx.Param("insulin_calculation_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID расчета",
		})
		return
	}

	creatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.FormInsulinCalculation(uint(id), creatorID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Расчет успешно сформирован",
	})
}

// CompleteInsulinCalculation godoc
// @Summary Завершить или отклонить расчет
// @Description Завершает или отклоняет расчет модератором
// @Tags insulin-calculations
// @Accept json
// @Produce json
// @Param insulin_calculation_id path int true "ID расчета инсулина"
// @Param data body serializer.CompleteCalculationRequest true "Статус завершения"
// @Success 200 {object} map[string]string "Расчет успешно завершен"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Расчет не найден"
// @Security ApiKeyAuth
// @Router /api/insulin-calculations/{insulin_calculation_id}/complete [put]
func (h *Handler) CompleteInsulinCalculation(ctx *gin.Context) {
	idStr := ctx.Param("insulin_calculation_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID расчета",
		})
		return
	}

	var request serializer.CompleteCalculationRequest

	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных",
		})
		return
	}

	moderatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.CompleteInsulinCalculation(uint(id), request.Status, moderatorID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Расчет успешно завершен",
	})
}

// DeleteInsulinCalculation godoc
// @Summary Удалить расчет
// @Description Удаляет расчет инсулина (только черновики)
// @Tags insulin-calculations
// @Accept json
// @Produce json
// @Param insulin_calculation_id path int true "ID расчета инсулина"
// @Success 200 {object} map[string]string "Расчет успешно удален"
// @Failure 400 {object} map[string]string "Неверный ID расчета"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Расчет не найден"
// @Security ApiKeyAuth
// @Router /api/insulin-calculations/{insulin_calculation_id} [delete]
func (h *Handler) DeleteInsulinCalculation(ctx *gin.Context) {
	idStr := ctx.Param("insulin_calculation_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID расчета",
		})
		return
	}

	creatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteInsulinCalculation(uint(id), creatorID); err != nil {
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

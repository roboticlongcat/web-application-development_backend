package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"sample/internal/app/serializer"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetPatients godoc
// @Summary Получить список пациентов
// @Description Возвращает список пациентов с возможностью фильтрации
// @Tags patients
// @Accept json
// @Produce json
// @Param name query string false "Фильтр по имени"
// @Param type query string false "Фильтр по типу пациента (1,2,3)"
// @Param status query string false "Фильтр по статусу"
// @Success 200 {array} serializer.PatientRespJSON "Список пациентов"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/patients [get]
func (h *Handler) GetPatients(ctx *gin.Context) {
	name := ctx.Query("name")
	patientType := ctx.Query("type")
	status := ctx.Query("status")

	filters := map[string]interface{}{}
	if name != "" {
		filters["name"] = name
	}
	if patientType != "" {
		filters["type"] = patientType
	}
	if status != "" {
		filters["status"] = status
	}

	patients, err := h.Repository.GetPatients(filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	var response []serializer.PatientRespJSON
	for _, patient := range patients {
		response = append(response, serializer.PatientToRespJSON(patient))
	}

	ctx.JSON(http.StatusOK, response)
}

// GetPatient godoc
// @Summary Получить пациента по ID
// @Description Возвращает данные конкретного пациента
// @Tags patients
// @Accept json
// @Produce json
// @Param patient_id path int true "ID пациента"
// @Success 200 {object} serializer.PatientRespJSON "Данные пациента"
// @Failure 400 {object} map[string]string "Неверный ID пациента"
// @Failure 404 {object} map[string]string "Пациент не найден"
// @Security ApiKeyAuth
// @Router /api/patients/{patient_id} [get]
func (h *Handler) GetPatient(ctx *gin.Context) {
	idStr := ctx.Param("patient_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID пациента",
		})
		return
	}

	patient, err := h.Repository.GetPatient(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Пациент не найден",
		})
		return
	}

	ctx.JSON(http.StatusOK, serializer.PatientToRespJSON(patient))
}

// CreatePatient godoc
// @Summary Создать нового пациента
// @Description Создает нового пациента в системе
// @Tags patients
// @Accept json
// @Produce json
// @Param patient body serializer.PatientJSON true "Данные пациента"
// @Success 201 {object} map[string]string "Пациент успешно создан"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/patients [post]
func (h *Handler) CreatePatient(ctx *gin.Context) {
	var patientJSON serializer.PatientJSON
	if err := ctx.BindJSON(&patientJSON); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных пациента",
		})
		return
	}
	patient := serializer.PatientFromJSON(patientJSON)
	if err := h.Repository.CreatePatient(patient); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Пациент успешно создан",
	})
}

// UpdatePatient godoc
// @Summary Обновить данные пациента
// @Description Обновляет данные существующего пациента
// @Tags patients
// @Accept json
// @Produce json
// @Param patient_id path int true "ID пациента"
// @Param patient body serializer.PatientJSON true "Новые данные пациента"
// @Success 200 {object} map[string]string "Пациент успешно обновлен"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 404 {object} map[string]string "Пациент не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/patients/{patient_id} [put]
func (h *Handler) UpdatePatient(ctx *gin.Context) {
	idStr := ctx.Param("patient_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID пациента",
		})
		return
	}
	var patientJSON serializer.PatientJSON
	if err := ctx.BindJSON(&patientJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный формат данных для обновления"))
		return
	}

	updates := map[string]interface{}{
		"name":        patientJSON.Name,
		"sensitivity": patientJSON.Sensitivity,
		"type":        patientJSON.Type,
		"glucose":     patientJSON.Glucose,
		"description": patientJSON.Description,
		"photo_url":   patientJSON.PhotoURL,
	}

	if err := h.Repository.UpdatePatient(uint(id), updates); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		logrus.Error(err)
		return
	}

	_, err = h.Repository.GetPatient(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("пациент не найден"))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Пациент успешно изменен",
	})
}

// DeletePatient godoc
// @Summary Удалить пациента
// @Description Помечает пациента как удаленного (мягкое удаление)
// @Tags patients
// @Accept json
// @Produce json
// @Param patient_id path int true "ID пациента"
// @Success 200 {object} map[string]string "Пациент успешно удален"
// @Failure 400 {object} map[string]string "Неверный ID пациента"
// @Failure 404 {object} map[string]string "Пациент не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/patients/{patient_id} [delete]
func (h *Handler) DeletePatient(ctx *gin.Context) {
	idStr := ctx.Param("patient_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID пациента",
		})
		return
	}

	if err := h.Repository.DeletePatient(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Пациент успешно удален",
	})
}

// UploadPatientPhoto godoc
// @Summary Загрузить фото пациента
// @Description Загружает фотографию для пациента
// @Tags patients
// @Accept multipart/form-data
// @Produce json
// @Param patient_id path int true "ID пациента"
// @Param photo formData file true "Фото пациента"
// @Success 200 {object} map[string]string "Фото успешно загружено"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 404 {object} map[string]string "Пациент не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/patients/{patient_id}/photo [post]
func (h *Handler) UploadPatientPhoto(ctx *gin.Context) {
	idStr := ctx.Param("patient_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID пациента",
		})
		return
	}

	file, err := ctx.FormFile("photo")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Файл не найден",
		})
		return
	}

	// Создаем временную папку если не существует
	tempDir := "./temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка создания временной папки",
		})
		return
	}

	// Сохраняем файл во временную папку
	filePath := fmt.Sprintf("%s/%d_%s", tempDir, id, file.Filename)
	if err := ctx.SaveUploadedFile(file, filePath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка сохранения файла: " + err.Error(),
		})
		return
	}

	// Загружаем в MinIO
	if err := h.Repository.UploadPatientPhoto(uint(id), filePath); err != nil {
		// Удаляем временный файл в случае ошибки
		os.Remove(filePath)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	// Удаляем временный файл после успешной загрузки
	os.Remove(filePath)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Фото успешно загружено",
	})
}

// AddPatientToInsulinCalculation godoc
// @Summary Добавить пациента в расчет инсулина
// @Description Добавляет пациента в заявку-черновик для расчета инсулина
// @Tags patients
// @Accept json
// @Produce json
// @Param patient_id path int true "ID пациента"
// @Param calculation body serializer.AddToCalculationRequest true "Данные для расчета"
// @Success 200 {object} map[string]string "Пациент успешно добавлен в расчет"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 404 {object} map[string]string "Пациент не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/patients/{patient_id}/add [post]
func (h *Handler) AddPatientToInsulinCalculation(ctx *gin.Context) {
	idStr := ctx.Param("patient_id")
	patientID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID пациента",
		})
		return
	}

	var request serializer.AddToCalculationRequest

	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных",
		})
		return
	}

	creatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.AddPatientToInsulinCalculation(
		uint(patientID),
		request.CurrentGlucose,
		request.BreadUnits,
		creatorID,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Пациент успешно добавлен в расчет",
	})
}

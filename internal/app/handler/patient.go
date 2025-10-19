package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"sample/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GET /api/patients - список пациентов с фильтрацией
func (h *Handler) GetPatients(ctx *gin.Context) {
	// Получаем параметры фильтрации
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

	ctx.JSON(http.StatusOK, patients)
}

// GET /api/patients/:id - получение одного пациента
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

	ctx.JSON(http.StatusOK, patient)
}

// POST /api/patients - создание пациента
func (h *Handler) CreatePatient(ctx *gin.Context) {
	var patient ds.Patient
	if err := ctx.BindJSON(&patient); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных пациента",
		})
		return
	}

	if err := h.Repository.CreatePatient(&patient); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Пациент успешно создан",
		"patient": patient,
	})
}

// PUT /api/patients/:id - обновление пациента
func (h *Handler) UpdatePatient(ctx *gin.Context) {
	idStr := ctx.Param("patient_id")
	id, err := strconv.Atoi(idStr)
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

	if err := h.Repository.UpdatePatient(uint(id), updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Пациент успешно обновлен",
	})
}

// DELETE /api/patients/:id - удаление пациента
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

// POST /api/patients/:id/photo - загрузка фото пациента
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

// POST /api/patients/:id/add - добавление пациента в расчет
func (h *Handler) AddPatientToInsulinCalculation(ctx *gin.Context) {
	idStr := ctx.Param("patient_id")
	patientID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID пациента",
		})
		return
	}

	var request struct {
		CurrentGlucose float32 `json:"current_glucose" binding:"required"`
		BreadUnits     float32 `json:"bread_units" binding:"required"`
	}

	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных",
		})
		return
	}

	err = h.Repository.AddPatientToInsulinCalculation(
		uint(patientID),
		request.CurrentGlucose,
		request.BreadUnits,
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

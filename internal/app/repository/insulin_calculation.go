package repository

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"sample/internal/app/auth"
	"sample/internal/app/ds"
)

// GET иконки корзины
func (r *Repository) GetInsulinCalculationInfo() (uint, int64, error) {
	insulinCalculationID, err := r.GetActiveInsulinCalculationID()
	if err != nil {
		logrus.Errorf("Error checking active calculation: %v", err)
		return 0, 0, err
	}
	if insulinCalculationID == 0 {
		return 0, 0, nil
	}

	count := r.GetInsulinCalculationCount()

	return insulinCalculationID, count, nil
}

// GET список (кроме удаленных и черновика, поля модератора и создателя через логины)
func (r *Repository) GetInsulinCalculations(filters map[string]interface{}) ([]ds.InsulinCalculation, error) {
	var insulinCalculations []ds.InsulinCalculation

	query := r.db.Preload("Creator", func(db *gorm.DB) *gorm.DB {
		return db.Select("user_id, username")
	}).Preload("Moderator", func(db *gorm.DB) *gorm.DB {
		return db.Select("user_id, username")
	}).Where("status != 'удален' AND status != 'черновик'")

	// Фильтрация по диапазону даты формирования и статусу
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		query = query.Where("calculated_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		query = query.Where("calculated_at <= ?", endDate)
	}

	err := query.Find(&insulinCalculations).Error
	if err != nil {
		return nil, err
	}

	return insulinCalculations, nil
}

// GET одна запись (поля расчета + его пациенты с данными из м-м связи)
func (r *Repository) GetInsulinCalculationWithPatients(id uint) (ds.InsulinCalculation, []map[string]interface{}, error) {
	var insulinCalculation ds.InsulinCalculation
	err := r.db.Preload("Creator", func(db *gorm.DB) *gorm.DB {
		return db.Select("user_id, username")
	}).
		Preload("Moderator", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id, username")
		}).
		Where("insulin_calculation_id = ?", id).
		First(&insulinCalculation).Error
	if err != nil {
		return ds.InsulinCalculation{}, nil, err
	}

	// Получаем пациентов с данными из связи many-to-many
	var patients []map[string]interface{}
	err = r.db.Table("insulin_calculation_patients").
		Select(`
            patients.name as name,
            patients.sensitivity as sensitivity,
            insulin_calculation_patients.current_glucose as current_glucose,
            insulin_calculation_patients.bread_units as bread_units,
            insulin_calculation_patients.calculated_insulin as calculated_insulin
        `).
		Joins("JOIN patients ON patients.patient_id = insulin_calculation_patients.patient_id").
		Where("insulin_calculation_patients.insulin_calculation_id = ?", id).
		Find(&patients).Error

	if err != nil {
		return ds.InsulinCalculation{}, nil, err
	}

	return insulinCalculation, patients, nil
}

// PUT изменения полей расчета по теме
func (r *Repository) UpdateInsulinCalculation(id uint, updates map[string]interface{}) error {
	currentUser := auth.GetCurrentUser()

	var insulinCalculation ds.InsulinCalculation
	if err := r.db.First(&insulinCalculation, id).Error; err != nil {
		return err
	}

	// Проверяем, что текущий пользователь - создатель расчета
	if insulinCalculation.CreatorID != currentUser.ID {
		return errors.New("только создатель может формировать расчет")
	}

	// Запрещаем изменение системных полей
	delete(updates, "insulin_calculation_id")
	delete(updates, "status")
	delete(updates, "created_at")
	delete(updates, "creator_id")
	delete(updates, "calculated_at")
	delete(updates, "completed_at")
	delete(updates, "moderator_id")

	if len(updates) == 0 {
		return errors.New("нет разрешенных полей для изменения")
	}

	return r.db.Model(&ds.InsulinCalculation{}).Where("insulin_calculation_id = ?", id).Updates(updates).Error
}

// PUT сформировать создателем (дата формирования). Проверка на обязательные поля
func (r *Repository) FormInsulinCalculation(id uint) error {
	currentUser := auth.GetCurrentUser()

	var insulinCalculation ds.InsulinCalculation
	if err := r.db.First(&insulinCalculation, id).Error; err != nil {
		return err
	}
	// Проверяем, что текущий пользователь - создатель расчета
	if insulinCalculation.CreatorID != currentUser.ID {
		return errors.New("только создатель может формировать расчет")
	}
	// Проверяем, что расчет в статусе черновика
	isDraft, err := r.IsDraftInsulinCalculation(id)
	if err != nil {
		return err
	}
	if !isDraft {
		return errors.New("можно формировать только черновики расчетов")
	}

	// Проверка, что в расчете есть пациенты
	count := r.GetInsulinCalculationCount()
	if count == 0 {
		return errors.New("нельзя сформировать расчет без пациентов")
	}

	now := time.Now()
	return r.db.Model(&ds.InsulinCalculation{}).Where("insulin_calculation_id = ?", id).Updates(map[string]interface{}{
		"status":        "сформирован",
		"calculated_at": &now,
	}).Error
}

// PUT завершить/отклонить модератором
func (r *Repository) CompleteInsulinCalculation(id uint, status string) error {
	moderator := auth.GetCurrentUser()

	if status != "завершён" && status != "отклонен" {
		return errors.New("неверный статус для завершения")
	}

	// Проверяем, что расчет в статусе сформирован
	var insulinCalculation ds.InsulinCalculation
	if err := r.db.Where("insulin_calculation_id = ? AND status = ?", id, "сформирован").First(&insulinCalculation).Error; err != nil {
		return errors.New("можно завершать только сформированные расчеты")
	}

	// Проверяем, что текущий пользователь - модератор
	if !moderator.IsModerator {
		return errors.New("только модератор может завершать расчет")
	}

	if err := r.recalculateAllInsulinInCalculation(id); err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       status,
		"completed_at": &now,
		"moderator_id": moderator.ID,
	}

	return r.db.Model(&ds.InsulinCalculation{}).Where("insulin_calculation_id = ?", id).Updates(updates).Error
}

// DELETE удаление (дата формирования)
func (r *Repository) DeleteInsulinCalculation(insulinCalculationID uint) error {
	currentUser := auth.GetCurrentUser()
	var insulinCalculation ds.InsulinCalculation
	if err := r.db.First(&insulinCalculation, insulinCalculationID).Error; err != nil {
		return err
	}

	// Проверяем, что текущий пользователь - создатель расчета
	if insulinCalculation.CreatorID != currentUser.ID {
		return errors.New("только создатель может удалять расчет")
	}

	// Проверяем, что расчет в статусе черновика
	if insulinCalculation.Status != "черновик" {
		return errors.New("можно удалять только черновики расчетов")
	}

	err := r.db.Model(&ds.InsulinCalculation{}).Where("insulin_calculation_id = ?", insulinCalculationID).UpdateColumn("status", "удален").Error
	if err != nil {
		return fmt.Errorf("ошибка при удалении расчета с id %d: %w", insulinCalculationID, err)
	}
	return nil
}

// вспомогательные бро
func (r *Repository) GetInsulinCalculation(id uint) ([]ds.InsulinCalculationPatients, error) {
	var insulincalculationPatients []ds.InsulinCalculationPatients
	err := r.db.Where("insulin_calculation_id = ?", id).Preload("Patient").Find(&insulincalculationPatients).Error
	if err != nil {
		return nil, err
	}

	return insulincalculationPatients, nil
}

func (r *Repository) GetInsulinCalculationCount() int64 {
	var insulincalculationID uint
	var count int64
	creatorID := auth.GetCurrentUser().ID

	err := r.db.Model(&ds.InsulinCalculation{}).Where("creator_id = ? AND status = ?", creatorID, "черновик").Select("insulin_calculation_id").First(&insulincalculationID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.InsulinCalculationPatients{}).Where("insulin_calculation_id = ?", insulincalculationID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in lists_chats:", err)
	}

	return count
}

func (r *Repository) GetActiveInsulinCalculationID() (uint, error) {
	var calculation ds.InsulinCalculation
	creatorID := auth.GetCurrentUser().ID

	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "черновик").
		Find(&calculation).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil // Черновика нет - это нормально
		}
		return 0, err // Другие ошибки прокидываем выше
	}
	return calculation.Insulin_Calculation_ID, nil
}

func (r *Repository) IsDraftInsulinCalculation(insulincalculationID uint) (bool, error) {
	var insulincalculation ds.InsulinCalculation
	err := r.db.Select("status").Where("insulin_calculation_id = ?", insulincalculationID).First(&insulincalculation).Error
	if err != nil {
		return false, err
	}
	return insulincalculation.Status == "черновик", nil
}

func (r *Repository) recalculateAllInsulinInCalculation(insulinCalculationID uint) error {
	patients, err := r.GetInsulinCalculation(insulinCalculationID)
	if err != nil {
		return err
	}

	for _, patient := range patients {
		// Пересчитываем инсулин для каждого пациента
		newInsulin := r.CalculateInsulin(
			patient.CurrentGlucose,
			patient.BreadUnits,
			patient.Patient_ID,
		)

		// Обновляем значение в базе
		err = r.db.Model(&ds.InsulinCalculationPatients{}).
			Where("insulin_calculation_id = ? AND patient_id = ?", insulinCalculationID, patient.Patient_ID).
			Update("calculated_insulin", newInsulin).Error

		if err != nil {
			return err
		}
	}

	return nil
}

// формула расчета инсулина с ограничениями
func (r *Repository) CalculateInsulin(currentGlucose, breadUnits float32, patientID uint) float32 {
	var patient ds.Patient
	if err := r.db.First(&patient, patientID).Error; err != nil {
		return 0
	}

	// Упрощенная формула на основе ваших заметок:
	// ДОЗА ИНСУЛИНА = ИНСУЛИН НА КОРРЕКЦИЮ + ИНСУЛИН НА ЕДУ

	targetGlucose := patient.Glucose // Целевое значение глюкозы

	// Ограничиваем чувствительность чтобы избежать деления на 0
	if patient.Sensitivity <= 0 {
		patient.Sensitivity = 0.1 // Минимальное значение
	}

	sensitivityCoefficient := float32(100.0) / patient.Sensitivity // Коэффициент чувствительности

	// Инсулин на коррекцию уровня глюкозы
	correctionInsulin := (currentGlucose - targetGlucose) / sensitivityCoefficient
	if correctionInsulin < 0 {
		correctionInsulin = 0 // Не даем отрицательный инсулин
	}

	// Инсулин на еду (на ХЕ)
	insulinCarbRatio := float32(500.0) / patient.Sensitivity // Соотношение инсулин/углеводы
	foodInsulin := insulinCarbRatio * breadUnits

	// Общая доза инсулина с ограничениями
	totalInsulin := correctionInsulin + foodInsulin

	// Ограничиваем минимальное значение
	if totalInsulin < 0 {
		totalInsulin = 0
	}

	// Округляем до 2 знаков после запятой
	totalInsulin = float32(math.Round(float64(totalInsulin)*100) / 100)

	return totalInsulin
}

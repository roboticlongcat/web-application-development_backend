package repository

import (
	"errors"
	"sample/internal/app/auth"
	"sample/internal/app/ds"
)

// DELETE удаление из расчета (без PK м-м)
func (r *Repository) RemovePatientFromInsulinCalculation(insulinCalculationID uint, patientID uint) error {
	currentUser := auth.GetCurrentUser()

	var insulinCalculation ds.InsulinCalculation
	if err := r.db.First(&insulinCalculation, insulinCalculationID).Error; err != nil {
		return err
	}

	// Проверяем, что текущий пользователь - создатель расчета
	if insulinCalculation.CreatorID != currentUser.ID {
		return errors.New("только создатель может формировать расчет")
	}
	return r.db.Where("insulin_calculation_id = ? AND patient_id = ?", insulinCalculationID, patientID).
		Delete(&ds.InsulinCalculationPatients{}).Error
}

// PUT изменение количества/порядка/значения в м-м (без PK м-м)
// PUT изменение количества/порядка/значения в м-м (без PK м-м)
func (r *Repository) UpdatePatientInInsulinCalculation(insulinCalculationID uint, patientID uint, updates map[string]interface{}) error {
	// Запрещаем изменение PK м-м
	delete(updates, "insulin_calculation_patient_id")
	delete(updates, "insulin_calculation_id")
	delete(updates, "patient_id")

	if len(updates) == 0 {
		return nil // Нет полей для обновления
	}

	currentUser := auth.GetCurrentUser()

	var insulinCalculation ds.InsulinCalculation
	if err := r.db.First(&insulinCalculation, insulinCalculationID).Error; err != nil {
		return err
	}

	// Проверяем, что текущий пользователь - создатель расчета
	if insulinCalculation.CreatorID != currentUser.ID {
		return errors.New("только создатель может изменять расчет")
	}

	// Если меняется глюкоза или хлебные единицы - пересчитываем инсулин
	if _, glucoseExists := updates["current_glucose"]; glucoseExists || updates["bread_units"] != nil {
		var calculationPatient ds.InsulinCalculationPatients
		if err := r.db.Where("insulin_calculation_id = ? AND patient_id = ?", insulinCalculationID, patientID).
			First(&calculationPatient).Error; err != nil {
			return err
		}

		// Обновляем значения для расчета из updates
		if glucose, ok := updates["current_glucose"].(float64); ok {
			calculationPatient.CurrentGlucose = float32(glucose)
		} else if glucose, ok := updates["current_glucose"].(float32); ok {
			calculationPatient.CurrentGlucose = glucose
		}

		if breadUnits, ok := updates["bread_units"].(float64); ok {
			calculationPatient.BreadUnits = float32(breadUnits)
		} else if breadUnits, ok := updates["bread_units"].(float32); ok {
			calculationPatient.BreadUnits = breadUnits
		}

		// Пересчитываем инсулин
		calculationPatient.CalculatedInsulin = r.CalculateInsulin(
			calculationPatient.CurrentGlucose,
			calculationPatient.BreadUnits,
			patientID,
		)
		updates["calculated_insulin"] = calculationPatient.CalculatedInsulin

		// Удаляем поля которые мы уже обработали, чтобы не было дублирования
		delete(updates, "current_glucose")
		delete(updates, "bread_units")
	}

	return r.db.Model(&ds.InsulinCalculationPatients{}).
		Where("insulin_calculation_id = ? AND patient_id = ?", insulinCalculationID, patientID).
		Updates(updates).Error
}

// PUT изменение количества хлебных единиц
func (r *Repository) UpdateBreadUnits(insulinCalculationID uint, patientID uint, breadUnits float32) error {
	currentUser := auth.GetCurrentUser()

	var insulinCalculation ds.InsulinCalculation
	if err := r.db.First(&insulinCalculation, insulinCalculationID).Error; err != nil {
		return err
	}

	// Проверяем, что текущий пользователь - создатель расчета
	if insulinCalculation.CreatorID != currentUser.ID {
		return errors.New("только создатель может формировать расчет")
	}

	return r.UpdatePatientInInsulinCalculation(insulinCalculationID, patientID, map[string]interface{}{
		"bread_units": breadUnits,
	})
}

// PUT изменение текущей глюкозы
func (r *Repository) UpdateCurrentGlucose(insulinCalculationID uint, patientID uint, glucose float32) error {
	currentUser := auth.GetCurrentUser()

	var insulinCalculation ds.InsulinCalculation
	if err := r.db.First(&insulinCalculation, insulinCalculationID).Error; err != nil {
		return err
	}

	// Проверяем, что текущий пользователь - создатель расчета
	if insulinCalculation.CreatorID != currentUser.ID {
		return errors.New("только создатель может формировать расчет")
	}

	return r.UpdatePatientInInsulinCalculation(insulinCalculationID, patientID, map[string]interface{}{
		"current_glucose": glucose,
	})
}

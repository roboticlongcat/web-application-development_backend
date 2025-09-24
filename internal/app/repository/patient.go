package repository

import (
	"fmt"

	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/sirupsen/logrus"

	"sample/internal/app/ds"
)

func (r *Repository) GetPatients() ([]ds.Patient, error) {
	var patients []ds.Patient
	err := r.db.Where("status != 'удален'").Find(&patients).Error
	if err != nil {
		return nil, err
	}
	if len(patients) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return patients, nil
}

func (r *Repository) GetPatient(id int) (ds.Patient, error) {
	patient := ds.Patient{}
	err := r.db.Where("id = ?", id).Find(&patient).Error

	if err != nil {
		return ds.Patient{}, err
	}

	return patient, nil
}

func (r *Repository) GetPatientsByName(name string) ([]ds.Patient, error) {
	var patients []ds.Patient
	err := r.db.Where("name ILIKE ? AND status != 'удален'", "%"+name+"%").Find(&patients).Error
	if err != nil {
		return nil, err
	}
	return patients, nil
}

func (r *Repository) GetCalculation(id int) ([]ds.CalculationPatient, error) {
	var calculationPatients []ds.CalculationPatient

	// Сначала получаем связи
	err := r.db.Where("calculation_id = ?", id).Preload("Patient").Find(&calculationPatients).Error
	if err != nil {
		return nil, err
	}

	return calculationPatients, nil
}

func (r *Repository) GetCalculationItemsCount(calculationID int) (int, error) {
	calculation, err := r.GetCalculation(calculationID)
	if err != nil {
		return 0, err
	}

	return len(calculation), nil
}

func (r *Repository) GetCalculationCount() int64 {
	var calculationID uint
	var count int64
	creatorID := 1

	err := r.db.Model(&ds.Calculation{}).Where("creator_id = ? AND status = ?", creatorID, "черновик").Select("id").First(&calculationID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.CalculationPatient{}).Where("calculation_id = ?", calculationID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in lists_chats:", err)
	}

	return count
}

func (r *Repository) DeleteCalculation(calculationID uint) error {
	err := r.db.Model(&ds.Calculation{}).Where("id = ?", calculationID).UpdateColumn("status", "удалён").Error
	fmt.Println(calculationID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении расчета с id %d: %w", calculationID, err)
	}

	return nil
}

func (r *Repository) GetActiveCalculationID() uint {
	var calculationID uint
	creatorID := 1

	err := r.db.Model(&ds.Calculation{}).
		Where("creator_id = ? AND status = ?", creatorID, "черновик").
		Select("id").First(&calculationID).Error

	if err != nil {
		return 0
	}
	return calculationID
}

func (r *Repository) AddPatientToCalculation(patientID uint, creatorID uint, currentGlucose, breadUnits float32) error {
	var calculation ds.Calculation

	// Ищем активную заявку пользователя
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "черновик").
		First(&calculation).Error

	// Если заявки нет - создаем новую
	if errors.Is(err, gorm.ErrRecordNotFound) {
		calculation = ds.Calculation{
			Status:      "черновик",
			CreatedAt:   time.Now(),
			CreatorID:   int(creatorID),
			ModeratorID: nil, // Пока нет модератора
		}
		if err := r.db.Create(&calculation).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Проверяем, нет ли уже этого пациента в заявке
	var count int64
	r.db.Model(&ds.CalculationPatient{}).
		Where("calculation_id = ? AND patient_id = ?", calculation.ID, patientID).
		Count(&count)

	if count == 0 {
		// Получаем данные пациента для расчета
		var patient ds.Patient
		if err := r.db.First(&patient, patientID).Error; err != nil {
			return err
		}

		// Рассчитываем инсулин по формуле

		calculatedInsulin := (currentGlucose - patient.Glucose) / patient.Sensitivity

		calculationPatient := ds.CalculationPatient{
			CalculationID:     calculation.ID,
			PatientID:         int(patientID),
			CurrentGlucose:    currentGlucose,
			BreadUnits:        breadUnits,
			CalculatedInsulin: calculatedInsulin,
		}

		if err := r.db.Create(&calculationPatient).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) IsDraftCalculation(calculationID int) (bool, error) {
	var calculation ds.Calculation
	err := r.db.Select("status").Where("id = ?", calculationID).First(&calculation).Error
	if err != nil {
		return false, err
	}
	return calculation.Status == "черновик", nil
}

func (r *Repository) HasActiveCalculation() bool {
	calculationID := r.GetActiveCalculationID()
	return calculationID != 0
}

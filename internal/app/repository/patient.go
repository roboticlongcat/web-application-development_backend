package repository

import (
	"fmt"
	"math"

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

func (r *Repository) GetPatient(id uint) (ds.Patient, error) {
	patient := ds.Patient{}
	err := r.db.Where("patient_id = ?", id).Find(&patient).Error

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

func (r *Repository) GetInsulinCalculation(id uint) ([]ds.InsulinCalculationPatients, error) {
	var insulincalculationPatients []ds.InsulinCalculationPatients
	err := r.db.Where("insulin_calculation_id = ?", id).Preload("Patient").Preload("InsulinCalculation").Find(&insulincalculationPatients).Error
	if err != nil {
		return nil, err
	}

	return insulincalculationPatients, nil
}

func (r *Repository) GetInsulinCalculationItemsCount(insulincalculationID uint) (int, error) {
	calculation, err := r.GetInsulinCalculation(insulincalculationID)
	if err != nil {
		return 0, err
	}

	return len(calculation), nil
}

func (r *Repository) GetInsulinCalculationCount() int64 {
	var insulincalculationID uint
	var count int64
	creatorID := 1

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

func (r *Repository) DeleteInsulinCalculation(insulincalculationID uint) error {
	err := r.db.Model(&ds.InsulinCalculation{}).Where("insulin_calculation_id = ?", insulincalculationID).UpdateColumn("status", "удален").Error
	fmt.Println(insulincalculationID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении расчета с id %d: %w", insulincalculationID, err)
	}

	return nil
}

func (r *Repository) GetActiveInsulinCalculationID() uint {
	var insulincalculationID uint
	creatorID := 1

	err := r.db.Model(&ds.InsulinCalculation{}).
		Where("creator_id = ? AND status = ?", creatorID, "черновик").
		Select("insulin_calculation_id").First(&insulincalculationID).Error

	if err != nil {
		return 0
	}
	return insulincalculationID
}

func (r *Repository) AddPatientToInsulinCalculation(patientID uint, creatorID uint, currentGlucose, breadUnits float32) error {
	var insulincalculation ds.InsulinCalculation

	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "черновик").
		First(&insulincalculation).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		var maxID uint
		r.db.Model(&ds.InsulinCalculation{}).Select("COALESCE(MAX(insulin_calculation_id))").Scan(&maxID)

		insulincalculation = ds.InsulinCalculation{
			Insulin_Calculation_ID: maxID + 1, // Явно указываем следующий ID
			Status:                 "черновик",
			CreatedAt:              time.Now(),
			CreatorID:              creatorID,
			ModeratorID:            2,
		}
		if err := r.db.Create(&insulincalculation).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	var count int64
	r.db.Model(&ds.InsulinCalculationPatients{}).
		Where("insulin_calculation_id = ? AND patient_id = ?", insulincalculation.Insulin_Calculation_ID, patientID).
		Count(&count)

	if count > 0 {
		err = r.db.Model(&ds.InsulinCalculationPatients{}).
			Where("insulin_calculation_id = ? AND patient_id = ?", insulincalculation.Insulin_Calculation_ID, patientID).
			Updates(map[string]interface{}{
				"current_glucose":    currentGlucose,
				"bread_units":        breadUnits,
				"calculated_insulin": 0,
			}).Error
	} else {
		var maxID uint
		r.db.Model(&ds.InsulinCalculationPatients{}).Select("COALESCE(MAX(insulin_calculation_patient_id))").Scan(&maxID)
		calculationPatient := ds.InsulinCalculationPatients{
			Insulin_Calculation_Patient_ID: maxID + 1,
			Insulin_Calculation_ID:         insulincalculation.Insulin_Calculation_ID,
			Patient_ID:                     patientID,
			CurrentGlucose:                 currentGlucose,
			BreadUnits:                     breadUnits,
			CalculatedInsulin:              0,
		}
		err = r.db.Create(&calculationPatient).Error
	}

	return err
}

func (r *Repository) IsDraftInsulinCalculation(insulincalculationID uint) (bool, error) {
	var insulincalculation ds.InsulinCalculation
	err := r.db.Select("status").Where("insulin_calculation_id = ?", insulincalculationID).First(&insulincalculation).Error
	if err != nil {
		return false, err
	}
	return insulincalculation.Status == "черновик", nil
}

func (r *Repository) HasActiveInsulinCalculation() bool {
	insulincalculationID := r.GetActiveInsulinCalculationID()
	return insulincalculationID != 0
}

// формула расчета инсулина
func (r *Repository) CalculateInsulin(currentGlucose, breadUnits float32, patientID uint) float32 {
	var patient ds.Patient
	if err := r.db.First(&patient, patientID).Error; err != nil {
		return 0
	}

	SDI := float32(100) / patient.Sensitivity
	correctionInsulin := (currentGlucose - patient.Glucose) / patient.Sensitivity
	insulinCarbRatio := float32(500.0) / SDI
	foodInsulin := insulinCarbRatio * breadUnits
	totalInsulin := correctionInsulin + foodInsulin

	if totalInsulin < 0 {
		totalInsulin = 0
	}

	totalInsulin = float32(math.Round(float64(totalInsulin)*100) / 100)

	// Логирование для отладки
	logrus.Printf(
		"Расчет инсулина: patientID=%d, СДИ=%.2f, КЧ=%.2f, коррекция=%.2f, инс/угл=%.2f, еда=%.2f, итого=%.2f",
		patientID, SDI, patient.Sensitivity, correctionInsulin, insulinCarbRatio, foodInsulin, totalInsulin,
	)

	return totalInsulin
}

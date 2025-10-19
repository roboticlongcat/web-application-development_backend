package repository

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"sample/internal/app/auth"
	"sample/internal/app/ds"
)

// GET список с фильтрацией
func (r *Repository) GetPatients(filters map[string]interface{}) ([]ds.Patient, error) {
	var patients []ds.Patient
	query := r.db.Where("status != 'удален'")
	// Применяем фильтры
	if name, ok := filters["name"]; ok && name != "" {
		query = query.Where("name ILIKE ?", "%"+name.(string)+"%")
	}
	if patientType, ok := filters["type"]; ok && patientType != "" {
		query = query.Where("type = ?", patientType)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&patients).Error
	if len(patients) == 0 {
		return nil, err
	}
	return patients, nil
}

// GET 1 запись
func (r *Repository) GetPatient(id uint) (ds.Patient, error) {
	patient := ds.Patient{}
	err := r.db.Where("patient_id = ?", id).Find(&patient).Error
	if err != nil {
		return ds.Patient{}, err
	}
	return patient, nil
}

// POST добавление без фото
func (r *Repository) CreatePatient(patient *ds.Patient) error {
	var maxID uint
	r.db.Model(&ds.Patient{}).Select("COALESCE(MAX(patient_id), 0)").Scan(&maxID)
	patient.Patient_ID = maxID + 1
	patient.Status = "действует"
	return r.db.Create(patient).Error
}

// PUT изменение
func (r *Repository) UpdatePatient(id uint, updates map[string]interface{}) error {
	delete(updates, "patient_id")
	delete(updates, "status")
	return r.db.Model(&ds.Patient{}).Where("patient_id = ? AND status != 'удален'", id).Updates(updates).Error
}

// DELETE удаление с фото
func (r *Repository) DeletePatient(id uint) error {
	var patient ds.Patient
	if err := r.db.Where("patient_id = ?", id).First(&patient).Error; err != nil {
		return err
	}

	if patient.PhotoURL != "" {
		if err := r.minioClient.DeletePatientPhoto(patient.PhotoURL); err != nil {
			// Логируем ошибку, но не прерываем удаление пациента
			//log.Printf("Ошибка удаления фото: %v", err)
		}
	}
	return r.db.Model(&ds.Patient{}).Where("patient_id = ?", id).Update("status", "удален").Error
}

// POST добавление изображения
func (r *Repository) UploadPatientPhoto(patientID uint, imagePath string) error {
	var patient ds.Patient
	if err := r.db.Where("patient_id = ? AND status != 'удален'", patientID).First(&patient).Error; err != nil {
		return err
	}

	if patient.PhotoURL != "" {
		if err := r.minioClient.DeletePatientPhoto(patient.PhotoURL); err != nil {
			// log.Printf("Ошибка удаления старого фото: %v", err)
		}
	}

	objectName, err := r.minioClient.UploadPatientPhoto(patientID, imagePath)
	if err != nil {
		return err
	}
	url := "http://localhost:9000/test/" + objectName

	return r.db.Model(&ds.Patient{}).Where("patient_id = ?", patientID).Update("photo_url", url).Error
}

// POST добавления в заявку-черновик
func (r *Repository) AddPatientToInsulinCalculation(patientID uint, currentGlucose, breadUnits float32) error {
	creatorID := auth.GetCurrentUser().ID

	var insulincalculation ds.InsulinCalculation

	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "черновик").First(&insulincalculation).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		var maxID uint
		r.db.Model(&ds.InsulinCalculation{}).Select("COALESCE(MAX(insulin_calculation_id))").Scan(&maxID)
		fmt.Print(maxID)

		newinsulincalculation := ds.InsulinCalculation{
			Insulin_Calculation_ID: maxID + 1,
			Status:                 "черновик",
			CreatedAt:              time.Now(),
			CreatorID:              creatorID,
			ModeratorID:            2,
		}
		if err := r.db.Create(&newinsulincalculation).Error; err != nil {
			return err
		}

		// Получаем созданную заявку обратно чтобы получить ID
		if err := r.db.Where("creator_id = ? AND status = ?", creatorID, "черновик").
			First(&insulincalculation).Error; err != nil {
			return err
		}
		// Проверяем что ID заявки не 0
		if insulincalculation.Insulin_Calculation_ID == 0 {
			return fmt.Errorf("не удалось получить ID заявки %d", maxID)
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
		var maxLinkID uint
		r.db.Model(&ds.InsulinCalculationPatients{}).Select("COALESCE(MAX(insulin_calculation_patient_id))").Scan(&maxLinkID)
		insulinCalculationPatient := ds.InsulinCalculationPatients{
			Insulin_Calculation_Patient_ID: maxLinkID + 1,
			Insulin_Calculation_ID:         insulincalculation.Insulin_Calculation_ID,
			Patient_ID:                     patientID,
			CurrentGlucose:                 currentGlucose,
			BreadUnits:                     breadUnits,
			CalculatedInsulin:              0,
		}
		err = r.db.Create(&insulinCalculationPatient).Error
	}

	return err
}

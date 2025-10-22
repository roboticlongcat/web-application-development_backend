package serializer

import "sample/internal/app/ds"

// PatientJSON - структура для сериализации пациента
type PatientJSON struct {
	Patient_ID  uint    `json:"patient_id"`
	Name        string  `json:"name" binding:"required"`
	Sensitivity float32 `json:"sensitivity" binding:"required"`
	Type        int     `json:"type" binding:"required"`
	Glucose     float32 `json:"glucose" binding:"required"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	PhotoURL    string  `json:"photo_url"`
}

// PatientRespJSON - структура для ответа с пациентом
type PatientRespJSON struct {
	Patient_ID  uint    `json:"patient_id"`
	Name        string  `json:"name"`
	Sensitivity float32 `json:"sensitivity"`
	Type        int     `json:"type"`
	Glucose     float32 `json:"glucose"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	PhotoURL    string  `json:"photo_url"`
}

// PatientToJSON преобразует модель в JSON
func PatientToRespJSON(patient ds.Patient) PatientRespJSON {
	return PatientRespJSON{
		Patient_ID:  patient.Patient_ID,
		Name:        patient.Name,
		Sensitivity: patient.Sensitivity,
		Type:        patient.Type,
		Glucose:     patient.Glucose,
		Description: patient.Description,
		Status:      patient.Status,
		PhotoURL:    patient.PhotoURL,
	}
}

// PatientFromJSON преобразует JSON в модель
func PatientFromJSON(patientJSON PatientJSON) ds.Patient {
	return ds.Patient{
		Patient_ID:  patientJSON.Patient_ID,
		Name:        patientJSON.Name,
		Sensitivity: patientJSON.Sensitivity,
		Type:        patientJSON.Type,
		Glucose:     patientJSON.Glucose,
		Description: patientJSON.Description,
		Status:      patientJSON.Status,
		PhotoURL:    patientJSON.PhotoURL,
	}
}

// PatientListToJSON преобразует список пациентов
func PatientListToJSON(patients []ds.Patient) []PatientRespJSON {
	result := make([]PatientRespJSON, len(patients))
	for i, patient := range patients {
		result[i] = PatientToRespJSON(patient)
	}
	return result
}

// AddToCalculationRequest - запрос на добавление пациента в расчет
type AddToCalculationRequest struct {
	CurrentGlucose float32 `json:"current_glucose" binding:"required"`
	BreadUnits     float32 `json:"bread_units" binding:"required"`
}

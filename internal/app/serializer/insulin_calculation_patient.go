package serializer

import (
	"sample/internal/app/ds"
)

// InsulinCalculationPatientJSON - структура для сериализации связи пациента с расчетом
type InsulinCalculationPatientJSON struct {
	InsulinCalculationPatientID uint    `json:"insulin_calculation_patient_id"`
	InsulinCalculationID        uint    `json:"insulin_calculation_id" binding:"required"`
	PatientID                   uint    `json:"patient_id" binding:"required"`
	CurrentGlucose              float32 `json:"current_glucose" binding:"required"`
	BreadUnits                  float32 `json:"bread_units" binding:"required"`
	CalculatedInsulin           float32 `json:"calculated_insulin"`
}

// InsulinCalculationPatientRespJSON - структура для ответа со связью пациента с расчетом
type InsulinCalculationPatientRespJSON struct {
	InsulinCalculationPatientID uint            `json:"insulin_calculation_patient_id"`
	InsulinCalculationID        uint            `json:"insulin_calculation_id"`
	PatientID                   uint            `json:"patient_id"`
	CurrentGlucose              float32         `json:"current_glucose"`
	BreadUnits                  float32         `json:"bread_units"`
	CalculatedInsulin           float32         `json:"calculated_insulin"`
	Patient                     PatientRespJSON `json:"patient,omitempty"`
}

// InsulinCalculationPatientToRespJSON преобразует модель в JSON для ответа
func InsulinCalculationPatientToRespJSON(icp ds.InsulinCalculationPatients) InsulinCalculationPatientRespJSON {
	resp := InsulinCalculationPatientRespJSON{
		InsulinCalculationPatientID: icp.Insulin_Calculation_Patient_ID,
		InsulinCalculationID:        icp.Insulin_Calculation_ID,
		PatientID:                   icp.Patient_ID,
		CurrentGlucose:              icp.CurrentGlucose,
		BreadUnits:                  icp.BreadUnits,
		CalculatedInsulin:           icp.CalculatedInsulin,
	}

	// Если загружены связанные данные пациента
	if icp.Patient.Patient_ID != 0 {
		resp.Patient = PatientToRespJSON(icp.Patient)
	}

	return resp
}

// InsulinCalculationPatientFromJSON преобразует JSON в модель
func InsulinCalculationPatientFromJSON(icpJSON InsulinCalculationPatientJSON) ds.InsulinCalculationPatients {
	return ds.InsulinCalculationPatients{
		Insulin_Calculation_Patient_ID: icpJSON.InsulinCalculationPatientID,
		Insulin_Calculation_ID:         icpJSON.InsulinCalculationID,
		Patient_ID:                     icpJSON.PatientID,
		CurrentGlucose:                 icpJSON.CurrentGlucose,
		BreadUnits:                     icpJSON.BreadUnits,
		CalculatedInsulin:              icpJSON.CalculatedInsulin,
	}
}

// UpdatePatientInCalculationRequest - запрос на обновление данных пациента в расчете
type UpdatePatientInCalculationRequest struct {
	CurrentGlucose float32 `json:"current_glucose" binding:"required"`
	BreadUnits     float32 `json:"bread_units" binding:"required"`
}

package serializer

import (
	"sample/internal/app/ds"
	"time"

	"github.com/google/uuid"
)

// InsulinCalculationJSON - структура для сериализации расчета инсулина
type InsulinCalculationJSON struct {
	InsulinCalculationID uint       `json:"insulin_calculation_id"`
	Status               string     `json:"status" binding:"required"`
	CreatedAt            time.Time  `json:"created_at"`
	CreatorID            uuid.UUID  `json:"creator_id" binding:"required"`
	CalculatedAt         *time.Time `json:"calculated_at"`
	CompletedAt          *time.Time `json:"completed_at"`
	ModeratorID          uuid.UUID  `json:"moderator_id"`
	Comment              string     `json:"comment"`
}

// InsulinCalculationRespJSON - структура для ответа с расчетом инсулина
type InsulinCalculationRespJSON struct {
	InsulinCalculationID uint       `json:"insulin_calculation_id"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"created_at"`
	CalculatedAt         *time.Time `json:"calculated_at"`
	CompletedAt          *time.Time `json:"completed_at"`
	Comment              string     `json:"comment"`
	CreatorUsername      string     `json:"creator_username"`
	ModeratorUsername    string     `json:"moderator_username,omitempty"`
}

// InsulinCalculationWithPatientsRespJSON - структура для ответа с расчетом и пациентами
type InsulinCalculationWithPatientsRespJSON struct {
	InsulinCalculation InsulinCalculationRespJSON     `json:"insulin_calculation"`
	Patients           []PatientInCalculationRespJSON `json:"patients"`
	PatientCount       int                            `json:"patient_count"`
}

// PatientInCalculationRespJSON - структура для пациента в расчете
type PatientInCalculationRespJSON struct {
	Name              string  `json:"name"`
	Sensitivity       float32 `json:"sensitivity"`
	CurrentGlucose    float32 `json:"current_glucose"`
	BreadUnits        float32 `json:"bread_units"`
	CalculatedInsulin float32 `json:"calculated_insulin"`
}

// InsulinCalculationCartInfoRespJSON - структура для информации о корзине
type InsulinCalculationCartInfoRespJSON struct {
	InsulinCalculationID uint  `json:"insulin_calculation_id"`
	PatientCount         int64 `json:"patient_count"`
}

// CompleteCalculationRequest - запрос на завершение расчета
type CompleteCalculationRequest struct {
	Status string `json:"status" binding:"required"`
}

// InsulinCalculationToRespJSON преобразует модель в JSON для ответа
func InsulinCalculationToRespJSON(calculation ds.InsulinCalculation) InsulinCalculationRespJSON {
	resp := InsulinCalculationRespJSON{
		InsulinCalculationID: calculation.Insulin_Calculation_ID,
		Status:               calculation.Status,
		CreatedAt:            calculation.CreatedAt,
		CalculatedAt:         calculation.CalculatedAt,
		CompletedAt:          calculation.CompletedAt,
		Comment:              calculation.Comment,
		CreatorUsername:      calculation.Creator.Username,
	}

	if calculation.Moderator.User_ID != uuid.Nil {
		resp.ModeratorUsername = calculation.Moderator.Username
	}

	return resp
}

// InsulinCalculationFromJSON преобразует JSON в модель
func InsulinCalculationFromJSON(calcJSON InsulinCalculationJSON) ds.InsulinCalculation {
	return ds.InsulinCalculation{
		Insulin_Calculation_ID: calcJSON.InsulinCalculationID,
		Status:                 calcJSON.Status,
		CreatedAt:              calcJSON.CreatedAt,
		CreatorID:              calcJSON.CreatorID,
		CalculatedAt:           calcJSON.CalculatedAt,
		CompletedAt:            calcJSON.CompletedAt,
		ModeratorID:            calcJSON.ModeratorID,
		Comment:                calcJSON.Comment,
	}
}

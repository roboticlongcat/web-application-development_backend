package ds

type InsulinCalculationPatients struct {
	Insulin_Calculation_Patient_ID uint    `gorm:"primaryKey"`
	Insulin_Calculation_ID         uint    `gorm:"primaryKey"`
	Patient_ID                     uint    `gorm:"primaryKey"`
	CurrentGlucose                 float32 `gorm:"type:decimal(4,2);not null;check:current_glucose > 0"`
	BreadUnits                     float32 `gorm:"type:decimal(4,2);not null;check:bread_units >= 0"`
	CalculatedInsulin              float32 `gorm:"type:decimal(4,2)"`

	Patient            Patient            `gorm:"foreignKey:Patient_ID;references:Patient_ID"`
	InsulinCalculation InsulinCalculation `gorm:"foreignKey:Patient_ID;references:Insulin_Calculation_ID"`
}

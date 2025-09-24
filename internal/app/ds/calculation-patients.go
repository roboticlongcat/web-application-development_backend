package ds

type CalculationPatient struct {
	CalculationID     int     `gorm:"primaryKey"`
	PatientID         int     `gorm:"primaryKey"`
	CurrentGlucose    float32 `gorm:"type:decimal(4,2);not null;check:current_glucose > 0"`
	BreadUnits        float32 `gorm:"type:decimal(4,2);not null;check:bread_units >= 0"`
	CalculatedInsulin float32 `gorm:"type:decimal(4,2)"`

	Patient     Patient     `gorm:"foreignKey:PatientID;references:ID"`
	Calculation Calculation `gorm:"foreignKey:PatientID;references:ID"`
}

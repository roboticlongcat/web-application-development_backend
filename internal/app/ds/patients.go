package ds

type Patient struct {
	Patient_ID  uint    `gorm:"primaryKey"`
	Name        string  `gorm:"type:varchar(100);not null"`
	Sensitivity float32 `gorm:"type:decimal(3,2);not null;check:sensitivity > 0"`
	Type        int     `gorm:"type:integer;not null;check:type IN (1,2,3)"`
	Glucose     float32 `gorm:"type:decimal(4,2);not null;check:glucose > 0"`
	Description string  `gorm:"type:text;not null"`
	Status      string  `gorm:"type:varchar(20);not null;default:'действует';check:status IN ('удален', 'действует')"`
	PhotoURL    string  `gorm:"type:varchar(500)"`
}


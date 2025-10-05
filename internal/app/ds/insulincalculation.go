package ds

import (
	"time"
)

type InsulinCalculation struct {
	Insulin_Calculation_ID uint       `gorm:"primaryKey"`
	Status                 string     `gorm:"type:varchar(20);not null;default:'удален';check:status IN ('черновик','удален','сформирован','завершён', 'отклонен')"`
	CreatedAt              time.Time  `gorm:"type:timestamp;not null;default:current_timestamp"`
	CreatorID              uint       `gorm:"type:integer;not null"`
	CalculatedAt           *time.Time `gorm:"type:timestamp"`
	CompletedAt            *time.Time `gorm:"type:timestamp"`
	ModeratorID            uint       `gorm:"type:integer"`

	Creator   User `gorm:"foreignKey:CreatorID"`
	Moderator User `gorm:"foreignKey:ModeratorID"`
}

package ds

import (
	"time"
)

type Calculation struct {
	ID           int        `gorm:"primaryKey"`
	Status       string     `gorm:"type:varchar(20);not null;default:'удален';check:status IN ('черновик','удален','сформирован','завершён', 'отклонен')"`
	CreatedAt    time.Time  `gorm:"type:timestamp;not null;default:current_timestamp"`
	CreatorID    int        `gorm:"type:integer;not null"`
	CalculatedAt *time.Time `gorm:"type:timestamp"`
	CompletedAt  *time.Time `gorm:"type:timestamp"`
	ModeratorID  *int       `gorm:"type:integer"`
}

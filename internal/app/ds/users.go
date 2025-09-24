package ds

type User struct {
	ID           int    `gorm:"primaryKey"`
	Username     string `gorm:"type:varchar(150);not null;unique"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	IsModerator  bool   `gorm:"type:boolean;not null;default:false"`
}

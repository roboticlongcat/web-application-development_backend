package ds

type User struct {
	User_ID      uint   `gorm:"primaryKey"`
	Username     string `gorm:"type:varchar(150);not null;unique"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	IsModerator  bool   `gorm:"type:boolean;not null;default:false"`
}

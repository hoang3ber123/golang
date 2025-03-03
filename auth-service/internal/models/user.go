package models

import (
	"auth-service/internal/services"
	"time"

	"gorm.io/gorm"
)

// Define user model
type User struct {
	BaseModel
	Username      string    `gorm:"unique;not null"`
	Password      string    `gorm:"not null"`
	Name          string    `gorm:"type:varchar(255)"`
	Email         string    `gorm:"type:varchar(60);not null"`
	IsEmailVerify bool      `gorm:"default:false;not null;"`
	Dob           time.Time `gorm:"not null"`
	Position      *string
	IsActive      bool    `gorm:"default:true;not null"`
	PhoneNumber   *string `gorm:"type:varchar(11)"`
	Contact       *string
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	// Gọi hàm tạo uuid của basemodel vì đã tạo hook cho BeforeCreate nên nó sẽ không nhận diện hook basemodel phải gọi thủ công
	u.BaseModel.BeforeCreate(tx)
	// Hash mật khẩu
	hashedPassword, err := services.HashPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword
	return
}

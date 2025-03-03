package models

import (
	"auth-service/internal/services"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Define user model
type Employee struct {
	BaseModel
	Username    string    `gorm:"unique;not null"`
	Password    string    `gorm:"not null"`
	Identity    string    `gorm:"not null;unique;index"`
	Name        string    `gorm:"type:varchar(255)"`
	Email       string    `gorm:"type:varchar(60);not null;unique"`
	Dob         time.Time `gorm:"not null"`
	Position    string    `gorm:"not null"`
	PhoneNumber string    `gorm:"type:varchar(11);not null"`
	Contact     string    `gorm:"not null"`
	IsActive    bool      `gorm:"default:true"`
	// relationship with Role
	RoleID *uuid.UUID `gorm:"type:char(36);index"` // Foreign key to Role
	Role   *Role      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// Hook
func (e *Employee) BeforeCreate(tx *gorm.DB) (err error) {
	// Gọi hàm tạo uuid của basemodel vì đã tạo hook cho BeforeCreate nên nó sẽ không nhận diện hook basemodel phải gọi thủ công
	e.BaseModel.BeforeCreate(tx)
	// Hash mật khẩu
	hashedPassword, err := services.HashPassword(e.Password)
	if err != nil {
		return err
	}
	e.Password = hashedPassword
	return
}

// struct query
type EmployeeQuery struct {
	Identity    string `query:"identity"`
	Username    string `query:"username"`
	Email       string `query:"email"`
	Name        string `query:"name"`
	Position    string `query:"position"`
	PhoneNumber string `query:"phone_number"`
	IsActive    bool   `query:"is_active"`
	RoleTitle   string `query:"role_title"`
}

package models

import (
	"errors"

	"gorm.io/gorm"
)

// Define role models
type Role struct {
	BaseSlugUnique
	Employees []Employee `gorm:"foreignKey:RoleID"`
}

func (r *Role) BeforeDelete(tx *gorm.DB) (err error) {
	if r.Title == "admin" || r.Title == "manager" || r.Title == "employee" {
		return errors.New("can't not delete admin or manager role")
	}
	return
}

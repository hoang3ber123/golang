package models

import "github.com/google/uuid"

type Category struct {
	BaseSlugUnique
	RelatedType string     `gorm:"type:varchar(50);not null"`
	Products    []Product  `gorm:"many2many:product_categories;constraint:OnDelete:CASCADE"`
	ParentID    *uuid.UUID `gorm:"type:char(36)"`
	Children    []Category `gorm:"foreignKey:ParentID;constraint:OnDelete:SET NULL"` // mối quan hệ một nhiều: category có nhiều category con
}

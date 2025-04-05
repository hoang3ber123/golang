package models

import "github.com/google/uuid"

type Category struct {
	BaseSlugUnique            // ID, Title, Slug, CreatedAt, CreatedDate
	Description    string     `gorm:"type:text" json:"description"`
	Products       []Product  `gorm:"many2many:product_categories;constraint:OnDelete:RESTRICT" json:"products"`
	ParentID       *uuid.UUID `gorm:"type:char(36)" json:"parent_id"`
	Children       []Category `gorm:"foreignKey:ParentID;constraint:OnDelete:RESTRICT" json:"children"`
}

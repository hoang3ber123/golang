package serializers

import (
	"auth-service/internal/db"
	"auth-service/internal/models"
	"errors"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type RoleCreateSerializer struct {
	Title string `json:"title" validate:"required"`
}

func (s *RoleCreateSerializer) IsValid(c *fiber.Ctx) error {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return errors.New("Invalid input: " + err.Error())
	}
	// Basic validation with go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return errors.New("Validation failed: " + err.Error())
	}
	// Custom validation: Check title is admin manager
	invalidTitles := map[string]bool{"admin": true, "manager": true, "employee": true} // init sets
	if invalidTitles[s.Title] {
		return errors.New("Title is not valid: " + s.Title)
	}

	// Custom validation: Check for duplicate Title in the database
	var titleExists bool
	if err := db.DB.Raw("SELECT EXISTS (SELECT 1 FROM roles WHERE title = ?)", s.Title).Scan(&titleExists).Error; err == nil && titleExists {
		// If a product with this name exists, return an error
		return errors.New("title already exists")
	}
	//
	return nil
}

// ToModel converts the serializer to a model
func (s *RoleCreateSerializer) ToModel() models.Role {
	return models.Role{
		BaseSlugUnique: models.BaseSlugUnique{ // Embeded struct
			Title: s.Title,
		},
	}
}

type RoleUpdateSerializer struct {
	Title string `json:"title"`
}

func (s *RoleUpdateSerializer) IsValid(c *fiber.Ctx) error {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return errors.New("Invalid input: " + err.Error())
	}
	// Basic validation with go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return errors.New("Validation failed: " + err.Error())

	}

	// Custom validation: Check title is admin manager
	invalidTitles := map[string]bool{"admin": true, "manager": true, "employee": true} // init sets
	if invalidTitles[s.Title] {
		return errors.New("Title is not valid: " + s.Title)
	}

	// Custom validation: Check for duplicate Title in the database
	var titleExists bool
	if err := db.DB.Raw("SELECT EXISTS (SELECT 1 FROM roles WHERE title = ?)", s.Title).Scan(&titleExists).Error; err == nil && titleExists {
		// If a product with this name exists, return an error
		return errors.New("title already exists")
	}
	//
	return nil
}

// Change validate data to instance
func (s *RoleUpdateSerializer) Update(instance *models.Role) error {
	if err := copier.Copy(&instance, &s); err != nil {
		return err
	}
	if err := db.DB.Save(&instance).Error; err != nil {
		return err
	}
	return nil
}

type RoleDetailResponseSerializer struct {
	BaseResponseSerializer
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

func RoleDetailResponse(instance *models.Role) *RoleDetailResponseSerializer {
	return &RoleDetailResponseSerializer{
		BaseResponseSerializer: BaseResponseSerializer{
			ID:        instance.ID,
			CreatedAt: instance.CreatedAt,
			UpdatedAt: instance.UpdatedAt,
		},
		Slug:  instance.Slug,
		Title: instance.Title,
	}
}

// RoleListResponseSerializer struct để serialize danh sách Role
type RoleListResponseSerializer struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
	Slug  string    `json:"slug"`
}

// RoleListResponse serialize danh sách Role thành slice RoleListResponseSerializer
func RoleListResponse(instance *[]models.Role) []RoleListResponseSerializer {
	results := make([]RoleListResponseSerializer, len(*instance)) // Preallocate slice

	for i, val := range *instance {
		// Copy từng phần tử từ models.Role vào serializer
		results[i] = RoleListResponseSerializer{
			ID:    val.ID,
			Title: val.Title,
			Slug:  val.Slug,
		}
	}

	return results
}

type RoleDeleteSerializer struct {
	IDs []string `json:"ids" validate:"required,dive,uuid_rfc4122"`
}

func (s *RoleDeleteSerializer) IsValid(c *fiber.Ctx) error {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return errors.New("Invalid input: " + err.Error())
	}
	// Basic validation with go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return errors.New("Validation failed: " + err.Error())
	}

	//
	return nil
}

func (s *RoleDeleteSerializer) Delete() error {
	slugTitles := []string{"admin", "manager", "employee"}

	result := db.DB.Where("id IN (?) and slug NOT IN (?)", s.IDs, slugTitles).Delete(&models.Role{})
	// Kiểm tra nếu không có bản ghi nào bị xóa
	if result.RowsAffected == 0 {
		return errors.New("no matching roles found")
	}
	//
	return nil
}

package serializers

import (
	"auth-service/internal/db"
	"auth-service/internal/models"
	"auth-service/internal/responses"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type RoleCreateSerializer struct {
	Title string `json:"title" validate:"required"`
}

func (s *RoleCreateSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}

	// Basic validation với go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	// Custom validation: Check title không nằm trong danh sách không hợp lệ
	invalidTitles := map[string]bool{
		"admin":    true,
		"manager":  true,
		"employee": true,
	}
	if invalidTitles[s.Title] {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Title is not valid: "+s.Title)
	}

	// Custom validation: Check trùng lặp Title trong database
	var titleExists bool
	if err := db.DB.Raw("SELECT EXISTS (SELECT 1 FROM roles WHERE title = ?)", s.Title).Scan(&titleExists).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if titleExists {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Title already exists: "+s.Title)
	}

	// Nếu không có lỗi, trả về nil
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

func (s *RoleUpdateSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}
	// Basic validation with go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	// Custom validation: Check title is admin manager
	invalidTitles := map[string]bool{"admin": true, "manager": true, "employee": true} // init sets
	if invalidTitles[s.Title] {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Title is not valid: "+s.Title)
	}

	// Custom validation: Check trùng lặp Title trong database
	var titleExists bool
	if err := db.DB.Raw("SELECT EXISTS (SELECT 1 FROM roles WHERE title = ?)", s.Title).Scan(&titleExists).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+err.Error())
	}
	if titleExists {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Title already exists: "+s.Title)
	}
	//
	return nil
}

// Change validate data to instance
func (s *RoleUpdateSerializer) Update(instance *models.Role) *responses.ErrorResponse {
	// Sao chép dữ liệu từ serializer sang instance
	if err := copier.Copy(instance, s); err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to copy data: "+err.Error())
	}

	// Lưu thay đổi vào database
	if err := db.DB.Save(instance).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to update role: "+err.Error())
	}

	// Trả về nil nếu thành công
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

func (s *RoleDeleteSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}
	// Basic validation with go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	//
	return nil
}

func (s *RoleDeleteSerializer) Delete() *responses.ErrorResponse {
	slugTitles := []string{"admin", "manager", "employee"}

	// Thực hiện xóa các role có ID trong s.IDs và slug không nằm trong slugTitles
	result := db.DB.Where("id IN (?) AND slug NOT IN (?)", s.IDs, slugTitles).Delete(&models.Role{})
	if result.Error != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to delete roles: "+result.Error.Error())
	}

	// Kiểm tra nếu không có bản ghi nào bị xóa
	if result.RowsAffected == 0 {
		return responses.NewErrorResponse(fiber.StatusNotFound, "No matching roles found to delete")
	}

	// Trả về nil nếu thành công
	return nil
}

package serializers

import (
	"fmt"
	"product-service/internal/db"
	"product-service/internal/models"
	"product-service/internal/responses"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type CategoryCreateSerializer struct {
	Title    string     `json:"title" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id"`
}

func (s *CategoryCreateSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}

	// Basic validation với go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	// Custom validation: Kiểm tra ParentID nếu có
	if s.ParentID != nil {
		var parentIDExists bool
		if err := db.DB.Raw("SELECT EXISTS (SELECT 1 FROM categories WHERE id = ?)", s.ParentID).Scan(&parentIDExists).Error; err != nil {
			return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error checking parent ID: "+err.Error())
		}
		if !parentIDExists {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "Category parent ID does not exist: "+fmt.Sprint(*s.ParentID))
		}
	}

	// Custom validation: Kiểm tra trùng lặp Title trong database
	var titleExists bool
	if err := db.DB.Raw("SELECT EXISTS (SELECT 1 FROM categories WHERE title = ?)", s.Title).Scan(&titleExists).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error checking title: "+err.Error())
	}
	if titleExists {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Title already exists: "+s.Title)
	}

	// Nếu không có lỗi, trả về nil
	return nil
}

// ToModel converts the serializer to a model
func (s *CategoryCreateSerializer) ToModel() *models.Category {
	return &models.Category{
		BaseSlugUnique: models.BaseSlugUnique{ // Embeded struct
			Title: s.Title,
		},
		ParentID: s.ParentID,
	}
}

type CategoryUpdateSerializer struct {
	Title    string     `json:"title"`
	ParentID *uuid.UUID `json:"parent_id"`
}

func (s *CategoryUpdateSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}

	// Basic validation với go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	// Custom validation: Kiểm tra ParentID nếu có
	if s.ParentID != nil {
		var parentIDExists bool
		if err := db.DB.Raw("SELECT EXISTS (SELECT 1 FROM categories WHERE id = ?)", s.ParentID).Scan(&parentIDExists).Error; err != nil {
			return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error checking parent ID: "+err.Error())
		}
		if !parentIDExists {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "Category parent ID does not exist: "+fmt.Sprint(*s.ParentID))
		}
	}

	// Custom validation: Kiểm tra trùng lặp Title trong database
	var titleExists bool
	if err := db.DB.Raw("SELECT EXISTS (SELECT 1 FROM categories WHERE title = ?)", s.Title).Scan(&titleExists).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error checking title: "+err.Error())
	}
	if titleExists {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Title already exists: "+s.Title)
	}

	// Nếu không có lỗi, trả về nil
	return nil
}

// Change validate data to instance
func (s *CategoryUpdateSerializer) Update(instance *models.Category) *responses.ErrorResponse {
	// Sao chép dữ liệu từ serializer sang instance
	if err := copier.Copy(instance, s); err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to copy data: "+err.Error())
	}

	// Lưu thay đổi vào database
	if err := db.DB.Save(instance).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to update: "+err.Error())
	}

	// Trả về nil nếu thành công
	return nil
}

type CategoryDetailResponseSerializer struct {
	BaseResponseSerializer
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

func CategoryDetailResponse(instance *models.Category) *CategoryDetailResponseSerializer {
	return &CategoryDetailResponseSerializer{
		BaseResponseSerializer: BaseResponseSerializer{
			ID:        instance.ID,
			CreatedAt: instance.CreatedAt,
			UpdatedAt: instance.UpdatedAt,
		},
		Slug:  instance.Slug,
		Title: instance.Title,
	}
}

// CategoryListResponseSerializer struct để serialize danh sách Category
type CategoryListResponseSerializer struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
	Slug  string    `json:"slug"`
}

// CategoryListResponse serialize danh sách Category thành slice CategoryListResponseSerializer
func CategoryListResponse(instance *[]models.Category) []CategoryListResponseSerializer {
	results := make([]CategoryListResponseSerializer, len(*instance)) // Preallocate slice

	for i, val := range *instance {
		// Copy từng phần tử từ models.Category vào serializer
		results[i] = CategoryListResponseSerializer{
			ID:    val.ID,
			Title: val.Title,
			Slug:  val.Slug,
		}
	}

	return results
}

type CategoryDeleteSerializer struct {
	IDs []string `json:"ids" validate:"required,dive,uuid_rfc4122"`
}

func (s *CategoryDeleteSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
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

func (s *CategoryDeleteSerializer) Delete() *responses.ErrorResponse {
	// Thực hiện xóa các category có ID trong s.IDs
	result := db.DB.Where("id IN (?)", s.IDs).Delete(&models.Category{})
	if result.Error != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to delete categories: "+result.Error.Error())
	}

	// Kiểm tra nếu không có bản ghi nào bị xóa
	if result.RowsAffected == 0 {
		return responses.NewErrorResponse(fiber.StatusNotFound, "No matching categories found to delete")
	}

	// Trả về nil nếu thành công
	return nil
}

type CategoryRecommendSerializer struct {
	Query string `json:"query" validate:"required"`
}

func (s *CategoryRecommendSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}

	// Basic validation với go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	// Nếu không có lỗi, trả về nil
	return nil
}

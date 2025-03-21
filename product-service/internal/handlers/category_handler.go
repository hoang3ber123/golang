package handlers

import (
	"errors"
	"product-service/internal/db"
	"product-service/internal/models"
	"product-service/internal/responses"
	"product-service/internal/serializers"
	"product-service/pagination"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CategoryCreate(c *fiber.Ctx) error {
	// Xử lý tạo category nếu xác thực thành công
	serializer := new(serializers.CategoryCreateSerializer)
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}
	// Serializer to model
	Category := serializer.ToModel()
	if err := db.DB.Create(&Category).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to create category: "+err.Error()).Send(c)
	}

	// Response
	return responses.NewSuccessResponse(fiber.StatusCreated, serializers.CategoryDetailResponse(Category)).Send(c)
}

func CategoryUpdate(c *fiber.Ctx) error {
	id := c.Params("id")
	var instance models.Category

	// Kiểm tra xem category có tồn tại không
	if err := db.DB.First(&instance, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return responses.NewErrorResponse(fiber.StatusNotFound, "Category not found").Send(c)
		}
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+err.Error()).Send(c)
	}

	serializer := new(serializers.CategoryUpdateSerializer)

	// Nếu có lỗi validation, return ngay lập tức
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}

	// Nếu validation OK, thực hiện update
	if err := serializer.Update(&instance); err != nil {
		return err.Send(c)
	}

	// Trả về response thành công
	return responses.NewSuccessResponse(fiber.StatusOK, serializers.CategoryDetailResponse(&instance)).Send(c)
}

func CategoryList(c *fiber.Ctx) error {
	// Prepare search scope (optional)
	titleQueries := c.Query("title")
	parentIDQueries := c.Query("parent_id")
	// Initializer query
	query := db.DB.Model(&models.Category{})
	if titleQueries != "" {
		query.Where("title LIKE ?", "%"+titleQueries+"%")
	}
	if parentIDQueries != "" {
		query.Where("parent_id = ?", parentIDQueries)
	} else {
		query.Where("parent_id IS NULL", parentIDQueries)
	}
	// Sử dụng hàm phân trang
	var Categories []models.Category
	paginator, err := pagination.PaginateWithGORM(c, query, &Categories)
	if err != nil {
		return err.Send(c)
	}
	var result interface{}
	if Categories != nil {
		result = serializers.CategoryListResponse(&Categories)
	}
	return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
		"pagination": paginator,
		"result":     result,
	}).Send(c)
}

func CategoryDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var instance models.Category
	if err := db.DB.First(&instance, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return responses.NewErrorResponse(fiber.StatusNotFound, "Category not found").Send(c)
		}
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+err.Error()).Send(c)
	}
	return responses.NewSuccessResponse(fiber.StatusOK, serializers.CategoryDetailResponse(&instance)).Send(c)
}

func CategoryDelete(c *fiber.Ctx) error {
	serializer := new(serializers.CategoryDeleteSerializer)
	//  Nếu có lỗi validation, return ngay lập tức
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}

	//  Nếu validation OK, thực hiện delete
	if err := serializer.Delete(); err != nil {
		return err.Send(c)
	}

	//  Trả về response thành công
	return responses.NewSuccessResponse(fiber.StatusOK, "Delete successfully").Send(c)
}

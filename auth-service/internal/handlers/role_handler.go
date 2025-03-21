package handlers

import (
	"auth-service/internal/db"
	"auth-service/internal/models"
	"auth-service/internal/responses"
	"auth-service/internal/serializers"
	"auth-service/pagination"
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RoleCreate(c *fiber.Ctx) error {
	serializer := new(serializers.RoleCreateSerializer)
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}

	// Chuyển serializer thành model
	role := serializer.ToModel()
	if err := db.DB.Create(&role).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to create role: "+err.Error()).Send(c)
	}

	// Response
	return responses.NewSuccessResponse(fiber.StatusCreated, serializers.RoleDetailResponse(&role)).Send(c)
}

func RoleUpdate(c *fiber.Ctx) error {
	id := c.Params("id")
	var instance models.Role

	// Kiểm tra xem role có tồn tại không
	if err := db.DB.First(&instance, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return responses.NewErrorResponse(fiber.StatusNotFound, "Role not found").Send(c)
		}
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+err.Error()).Send(c)
	}

	serializer := new(serializers.RoleUpdateSerializer)

	// Nếu có lỗi validation, return ngay lập tức
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}

	// Nếu validation OK, thực hiện update
	if err := serializer.Update(&instance); err != nil {
		return err.Send(c)
	}

	// Trả về response thành công
	return responses.NewSuccessResponse(fiber.StatusOK, serializers.RoleDetailResponse(&instance)).Send(c)
}

func RoleList(c *fiber.Ctx) error {
	// Prepare search scope (optional)
	titleQueries := c.Query("title")
	// Initializer query
	query := db.DB.Model(&models.Role{})
	if titleQueries != "" {
		query.Where("title LIKE ?", "%"+titleQueries+"%")
	}
	// Sử dụng hàm phân trang
	var roles []models.Role
	paginator, err := pagination.PaginateWithGORM(c, query, &roles)
	if err != nil {
		return err.Send(c)
	}
	var result interface{}
	if roles != nil {
		result = serializers.RoleListResponse(&roles)
	}
	return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
		"pagination": paginator,
		"result":     result,
	}).Send(c)
}

func RoleDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var instance models.Role
	if err := db.DB.First(&instance, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return responses.NewErrorResponse(fiber.StatusNotFound, "Role not found").Send(c)
		}
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+err.Error()).Send(c)
	}
	return responses.NewSuccessResponse(fiber.StatusOK, serializers.RoleDetailResponse(&instance)).Send(c)
}

func RoleDelete(c *fiber.Ctx) error {

	serializer := new(serializers.RoleDeleteSerializer)

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

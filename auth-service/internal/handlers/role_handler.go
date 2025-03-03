package handlers

import (
	"auth-service/internal/db"
	"auth-service/internal/models"
	"auth-service/internal/responses"
	"auth-service/internal/serializers"
	"auth-service/pagination"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RoleCreate(c *fiber.Ctx) error {
	serializer := new(serializers.RoleCreateSerializer)
	if err := serializer.IsValid(c); err != nil {
		return responses.SendErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	// Serializer to model
	role := serializer.ToModel()
	if err := db.DB.Create(&role).Error; err != nil {
		return responses.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	// Response
	return responses.SendSuccessResponse(c, fiber.StatusCreated, serializers.RoleDetailResponse(&role))
}

func RoleUpdate(c *fiber.Ctx) error {
	id := c.Params("id")
	var instance models.Role

	// Kiểm tra xem role có tồn tại không
	if err := db.DB.First(&instance, "id = ?", id).Error; err == gorm.ErrRecordNotFound {
		return responses.SendErrorResponse(c, fiber.StatusNotFound, "Role not found")
	}

	serializer := new(serializers.RoleUpdateSerializer)

	// Nếu có lỗi validation, return ngay lập tức
	if err := serializer.IsValid(c); err != nil {
		return responses.SendErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	// Nếu validation OK, thực hiện update
	if err := serializer.Update(&instance); err != nil {
		return responses.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	// Trả về response thành công
	return responses.SendSuccessResponse(c, fiber.StatusOK, serializers.RoleDetailResponse(&instance))
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
		return responses.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch: "+err.Error())
	}

	return responses.SendSuccessResponse(c, fiber.StatusOK, fiber.Map{
		"pagination": paginator,
		"result":     serializers.RoleListResponse(&roles),
	})
}

func RoleDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var instance models.Role
	if err := db.DB.First(&instance, "slug = ?", slug).Error; err == gorm.ErrRecordNotFound {
		return responses.SendErrorResponse(c, fiber.StatusNotFound, "Role not found")
	}
	return responses.SendSuccessResponse(c, fiber.StatusOK, serializers.RoleDetailResponse(&instance))
}

func RoleDelete(c *fiber.Ctx) error {

	serializer := new(serializers.RoleDeleteSerializer)

	//  Nếu có lỗi validation, return ngay lập tức
	if err := serializer.IsValid(c); err != nil {
		return responses.SendErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	//  Nếu validation OK, thực hiện delete
	if err := serializer.Delete(); err != nil {
		return responses.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	//  Trả về response thành công
	return responses.SendSuccessResponse(c, fiber.StatusOK, "delete successfully")
}

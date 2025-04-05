package handlers

import (
	"auth-service/internal/db"
	"auth-service/internal/models"
	"auth-service/internal/responses"
	"auth-service/internal/serializers"
	"auth-service/pagination"

	"github.com/gofiber/fiber/v2"
)

func UserDetail(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return responses.NewSuccessResponse(fiber.StatusOK, serializers.UserDetailResponse(user)).Send(c)
}

func UserList(c *fiber.Ctx) error {
	// Prepare search scope (optional)
	usernameQueries := c.Query("username")
	emailQueries := c.Query("email")
	isEmailVerify := c.Query("is_email_verify")
	isActive := c.Query("is_active")

	// Initializer query
	query := db.DB.Model(&models.User{})

	if usernameQueries != "" {
		query.Where("username LIKE ?", "%"+usernameQueries+"%")
	}

	if emailQueries != "" {
		query.Where("email LIKE ?", "%"+emailQueries+"%")
	}

	// Kiểm tra is_email_verify nếu có
	if isEmailVerify != "" {
		if isEmailVerify == "true" {
			query.Where("is_email_verify = ?", true)
		} else if isEmailVerify == "false" {
			query.Where("is_email_verify = ?", false)
		}
	}

	// Kiểm tra is_active nếu có
	if isActive != "" {
		if isActive == "true" {
			query.Where("is_active = ?", true)
		} else if isActive == "false" {
			query.Where("is_active = ?", false)
		}
	}

	// Sử dụng hàm phân trang
	var instance []models.User
	paginator, err := pagination.PaginateWithGORM(c, query, &instance)
	if err != nil {
		return err.Send(c)
	}

	var result interface{}
	if instance != nil {
		result = serializers.UserListResponse(&instance)
	}
	return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
		"pagination": paginator,
		"result":     result,
	}).Send(c)
}

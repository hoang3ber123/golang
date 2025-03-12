package handlers

import (
	"auth-service/internal/models"
	"auth-service/internal/responses"
	"auth-service/internal/serializers"

	"github.com/gofiber/fiber/v2"
)

func UserDetail(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return responses.NewSuccessResponse(fiber.StatusOK, serializers.UserDetailResponse(user)).Send(c)
}

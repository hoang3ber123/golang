package handlers

import (
	"auth-service/internal/models"
	"auth-service/internal/responses"
	"auth-service/internal/serializers"

	"github.com/gofiber/fiber/v2"
)

func UserDetail(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return responses.SendSuccessResponse(c, fiber.StatusOK, serializers.UserDetailResponse(user))
}

func Decentralize(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": serializers.UserDetailResponse(user),
	})
}

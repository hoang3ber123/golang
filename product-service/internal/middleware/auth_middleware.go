package middleware

import (
	grpcclient "product-service/internal/grpc_client"
	"product-service/internal/models"
	"product-service/internal/responses"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func AuthEmployeeMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// sample token string taken from the New example
		tokenString := c.Cookies("Authorization-employee")
		if tokenString == "" {
			return responses.ErrForbiden.Send(c)
		}
		// Check employee authenticated
		userInfo, err := grpcclient.AuthRequest(tokenString, allowedRoles)
		if err != nil {
			return err.Send(c)
		}

		id, _ := uuid.Parse(userInfo.ID)

		employee := &models.Employee{
			ID:        id,
			Identity:  userInfo.Identity,
			Email:     userInfo.Email,
			Name:      userInfo.Name,
			IsActive:  userInfo.IsActive,
			RoleTitle: userInfo.RoleTitle,
		}

		// Save userInfo in context fiber
		c.Locals("employee", employee)

		// Proceed to next handler
		return c.Next()
	}
}

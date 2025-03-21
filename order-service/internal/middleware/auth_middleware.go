package middleware

import (
	grpcclient "order-service/internal/grpc_client"
	"order-service/internal/models"
	"order-service/internal/responses"

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
		userInfo, err := grpcclient.AuthEmployeeRequest(tokenString, allowedRoles)
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
func AuthUserMiddleware(c *fiber.Ctx) error {
	// sample token string taken from the New example
	tokenString := c.Cookies("Authorization")
	if tokenString == "" {
		return responses.ErrForbiden.Send(c)
	}
	// Check User authenticated
	userInfo, err := grpcclient.AuthUserRequest(tokenString)
	if err != nil {
		return err.Send(c)
	}
	id, _ := uuid.Parse(userInfo.ID)

	user := &models.User{
		ID:    id,
		Email: userInfo.Email,
		Name:  userInfo.Name,
	}
	// Save userInfo in context fiber
	c.Locals("user", user)

	// Proceed to next handler
	return c.Next()
}

package middleware

import (
	"auth-service/config"
	"auth-service/internal/db"
	"auth-service/internal/models"
	"auth-service/internal/responses"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func JWTAuthMiddleware(c *fiber.Ctx) error {
	// Get secret key
	secretKey := config.Config.JWTSecret
	if secretKey == "" {
		log.Println("Warning: JWT_SECRET is empty")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server misconfiguration",
		})
	}
	// sample token string taken from the New example
	tokenString := c.Cookies("Authorization")

	// if don't have token in cookie
	if tokenString == "" {
		log.Println("Warning: tokenString is empty")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid token",
		})
	}

	// Check token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Println("Warning: claims is not valid")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// Check token expiration
	exp, ok := claims["exp"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Token does not contain expiration",
		})
	}
	if time.Now().Unix() > int64(exp) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Token is expired",
		})
	}

	// Find user
	var user models.User
	db.DB.First(&user, "id = ?", claims["sub"])
	if user.ID == uuid.Nil {
		log.Printf("Warning: Can not find user %s", claims["sub"])
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}
	if !user.IsEmailVerify {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Please verify your email",
		})
	}
	if !user.IsActive {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Your account is blocked or unactived",
		})
	}

	// Store user info in context for later use
	c.Locals("user", &user)

	// Proceed to next handler
	return c.Next()
}

func JWTAuthEmployeeMiddleware(c *fiber.Ctx) error {
	// Get secret key
	secretKey := config.Config.JWTEmployeeSecret
	if secretKey == "" {
		log.Println("Warning: JWTEmployeeSecret is empty")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server misconfiguration",
		})
	}
	// sample token string taken from the New example
	tokenString := c.Cookies("Authorization-employee")

	// if don't have token in cookie
	if tokenString == "" {
		log.Println("Warning: tokenString is empty")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid token",
		})
	}

	// Check token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Println("Warning: claims is not valid")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// Check token expiration
	exp, ok := claims["exp"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Token does not contain expiration",
		})
	}
	if time.Now().Unix() > int64(exp) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Token is expired",
		})
	}

	// Find user
	var user models.Employee
	db.DB.Joins("Role").First(&user, "employees.id = ?", claims["sub"])
	if user.ID == uuid.Nil {
		log.Printf("Warning: Can not find user %s", claims["sub"])
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}
	if !user.IsActive {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Your account is blocked or unactived",
		})
	}
	// Store user info in context for later use
	c.Locals("employee", &user)

	// Proceed to next handler
	return c.Next()
}

// Middleware kiểm tra vai trò
func RestrictRoleMiddlware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Giả sử thông tin user được lấy từ context (thực tế có thể từ JWT, session, v.v.)
		user := c.Locals("employee").(*models.Employee)
		// Kiểm tra xem vai trò của user có nằm trong danh sách được phép không
		if user.Role != nil {
			for _, role := range allowedRoles {
				if user.Role.Title == role {
					return c.Next() // Cho phép đi tiếp nếu hợp lệ
				}
			}
		}

		// Trả về lỗi nếu không có quyền
		return responses.SendErrorResponse(c, fiber.StatusForbidden, "You don't have permission")
	}
}

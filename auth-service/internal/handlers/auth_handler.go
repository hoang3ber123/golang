package handlers

import (
	"auth-service/config"
	"auth-service/internal/db"
	"auth-service/internal/email"
	"auth-service/internal/models"
	"auth-service/internal/responses"
	"auth-service/internal/serializers"
	"auth-service/internal/services"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func SignUp(c *fiber.Ctx) error {

	serializer := new(serializers.UserSignUpSerializer)
	if err := serializer.IsValid(c); err != nil {
		return err // Error is already formatted by Deserialize
	}

	// Serializer to model
	user := serializer.ToModel()
	if err := db.DB.Create(&user).Error; err != nil {
		return responses.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create product")
	}

	// Send verify email
	go func() {
		email.SendVerifyMail(user.ID, user.Email)
	}()

	// Response
	return responses.SendSuccessResponse(c, fiber.StatusCreated, "User register successfully, please verify email to active account")
}

func VerifyEmail(c *fiber.Ctx) error {
	// get token
	tokenString := c.Params("token")
	// if don't have token in cookie
	if tokenString == "" {
		return responses.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid token")
	}
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Config.JWTSecretMail), nil
	})
	if err != nil {
		return responses.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid token")
	}
	// Check token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return responses.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid token")
	}
	// Check token expiration
	exp, ok := claims["exp"].(float64)
	if !ok {
		return responses.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid token claims")
	}
	if time.Now().Unix() > int64(exp) {
		return responses.SendErrorResponse(c, fiber.StatusBadRequest, "Token is expired")
	}
	// Lấy user ID từ token
	sub, ok := claims["sub"].(string)
	if !ok {
		return responses.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid token payload")
	}
	userID, err := uuid.Parse(sub)
	if err != nil {
		return responses.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid user ID")
	}
	// Cập nhật trạng thái xác thực email trong database
	result := db.DB.Model(&models.User{}).Where("id = ? AND is_email_verify = false", userID).Update("is_email_verify", true)
	if result.RowsAffected == 0 {
		return responses.SendErrorResponse(c, fiber.StatusNotFound, "User not found or already verified") // Error is already formatted by Deserialize
	}
	return responses.SendSuccessResponse(c, fiber.StatusOK, "Email verified successfully")
}

func Login(c *fiber.Ctx) error {
	// Using UserloginSerializer
	serializer := new(serializers.UserLoginSerializer)
	// Validate request
	user, err := serializer.Login(c)
	if err != nil {
		return responses.SendErrorResponse(c, fiber.StatusBadRequest, err.Error()) // Error is already formatted by Deserialize
	}
	// Generate jwt token
	tokenString, _ := services.GenerateJWT(user.ID)
	// Setting tokenString to cookies
	c.Cookie(&fiber.Cookie{
		Name:     "Authorization",
		Value:    tokenString,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HTTPOnly: true,   // không cho phép javascript đọc cookie
		Secure:   false,  // Không cần HTTPS trong môi trường test
		SameSite: "None", // Cho phép gửi với mọi request cross-site
	})
	// Response
	return responses.SendSuccessResponse(c, fiber.StatusOK, "Login succesfully")
}

func Logout(c *fiber.Ctx) error {
	cookie := new(fiber.Cookie)
	cookie.Name = "Authorization"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-time.Hour) // Hết hạn ngay lập tức

	c.Cookie(cookie)

	return responses.SendSuccessResponse(c, fiber.StatusOK, "Sign out successfully")
}

func EmployeeLogin(c *fiber.Ctx) error {
	// Using UserloginSerializer
	serializer := new(serializers.UserLoginSerializer)
	// Validate request
	user, err := serializer.EmployeeLogin(c)
	if err != nil {
		return responses.SendErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	// Generate jwt token
	tokenString, _ := services.GenerateEmployeeJWT(user.ID)
	// Setting tokenString to cookies
	c.Cookie(&fiber.Cookie{
		Name:     "Authorization-employee",
		Value:    tokenString,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HTTPOnly: true,   // không cho phép javascript đọc cookie
		Secure:   false,  // Không cần HTTPS trong môi trường test
		SameSite: "None", // Cho phép gửi với mọi request cross-site
	})
	// Response
	return responses.SendSuccessResponse(c, fiber.StatusOK, "Login succesfully")
}

func EmployeeLogout(c *fiber.Ctx) error {
	cookie := new(fiber.Cookie)
	cookie.Name = "Authorization-employee"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-time.Hour) // Hết hạn ngay lập tức

	c.Cookie(cookie)

	return responses.SendSuccessResponse(c, fiber.StatusOK, "Sign out successfully")
}

func EmployeeSignUp(c *fiber.Ctx) error {

	serializer := new(serializers.UserSignUpSerializer)
	if err := serializer.IsValid(c); err != nil {
		return err // Error is already formatted by Deserialize
	}

	// Serializer to model
	user := serializer.ToModel()
	if err := db.DB.Create(&user).Error; err != nil {
		return responses.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create product")
	}

	// Response
	return responses.SendSuccessResponse(c, fiber.StatusCreated, "Create successfully")
}

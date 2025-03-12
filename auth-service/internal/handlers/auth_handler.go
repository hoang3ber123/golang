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
		return err.Send(c) // Error is already formatted by Deserialize
	}

	// Serializer to model
	user := serializer.ToModel()
	if err := db.DB.Create(&user).Error; err != nil {
		// Create error
		return responses.NewErrorResponse(fiber.StatusInternalServerError, err.Error()).Send(c)
	}

	// Send verify email
	go func() {
		email.SendVerifyMail(user.ID, user.Email)
	}()

	// Response
	return responses.NewSuccessResponse(fiber.StatusCreated,
		"User register successfully, please verify email to active account").Send(c)
}

func VerifyEmail(c *fiber.Ctx) error {
	// get token
	tokenString := c.Params("token")
	// if don't have token in cookie
	if tokenString == "" {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid token").Send(c)
	}
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Config.JWTSecretMail), nil
	})
	if err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid token").Send(c)
	}
	// Check token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid token").Send(c)
	}
	// Check token expiration
	exp, ok := claims["exp"].(float64)
	if !ok {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid token claims").Send(c)
	}
	if time.Now().Unix() > int64(exp) {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Token is expired").Send(c)
	}
	// Lấy user ID từ token
	sub, ok := claims["sub"].(string)
	if !ok {
		return responses.NewErrorResponse(fiber.StatusUnauthorized, "Invalid token payload").Send(c)
	}
	userID, err := uuid.Parse(sub)
	if err != nil {
		return responses.NewErrorResponse(fiber.StatusUnauthorized, "Invalid user ID").Send(c)
	}
	// Cập nhật trạng thái xác thực email trong database
	result := db.DB.Model(&models.User{}).Where("id = ? AND is_email_verify = false", userID).Update("is_email_verify", true)
	if result.RowsAffected == 0 {
		return responses.NewErrorResponse(fiber.StatusNotFound, "User not found or already verified").Send(c) // Error is already formatted by Deserialize
	}
	return responses.NewSuccessResponse(fiber.StatusOK, "Email verified successfully").Send(c)
}

func Login(c *fiber.Ctx) error {
	// Using UserloginSerializer
	serializer := new(serializers.UserLoginSerializer)
	// Validate request
	user, err := serializer.Login(c)
	if err != nil {
		return err.Send(c)
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
	return responses.NewSuccessResponse(fiber.StatusOK, "Login succesfully").Send(c)
}

func Logout(c *fiber.Ctx) error {
	cookie := new(fiber.Cookie)
	cookie.Name = "Authorization"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-time.Hour) // Hết hạn ngay lập tức

	c.Cookie(cookie)

	return responses.NewSuccessResponse(fiber.StatusOK, "Sign out successfully").Send(c)
}

func EmployeeLogin(c *fiber.Ctx) error {
	// Using UserloginSerializer
	serializer := new(serializers.UserLoginSerializer)
	// Validate request
	user, err := serializer.EmployeeLogin(c)
	if err != nil {
		return err.Send(c)
	}
	// Generate jwt token
	tokenString, _ := services.GenerateEmployeeJWT(user.ID)
	fmt.Printf("token string: %s", tokenString)
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
	return responses.NewSuccessResponse(fiber.StatusOK, "Login succesfully").Send(c)
}

func EmployeeLogout(c *fiber.Ctx) error {
	cookie := new(fiber.Cookie)
	cookie.Name = "Authorization-employee"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-time.Hour) // Hết hạn ngay lập tức

	c.Cookie(cookie)

	return responses.NewSuccessResponse(fiber.StatusOK, "Sign out successfully").Send(c)
}

func EmployeeSignUp(c *fiber.Ctx) error {

	serializer := new(serializers.UserSignUpSerializer)
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c) // Error is already formatted by Deserialize
	}

	// Serializer to model
	user := serializer.ToModel()
	if err := db.DB.Create(&user).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, err.Error()).Send(c)
	}

	// Response
	return responses.NewSuccessResponse(fiber.StatusCreated, "Create successfully").Send(c)
}

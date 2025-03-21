package services

import (
	"auth-service/config"
	"auth-service/internal/responses"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	HashPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", errors.New("failed to hash password")
	}
	return string(HashPassword), nil
}

func CheckPasswordHash(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil // Nếu không có lỗi thì mật khẩu đúng
}

// Hàm tạo JWT token
func GenerateJWT(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Config.JWTSecret))
}

// Hàm tạo JWT token cho employee
func GenerateEmployeeJWT(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Config.JWTEmployeeSecret))
}

// Hàm tạo token để verify email
func GenerateTokenVerifyEmailJWT(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Minute * 10).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Config.JWTSecretMail))
}

// IsEmployeeAuthenticated xác thực token của employee và trả về sub (user ID) nếu hợp lệ
func IsEmployeeAuthenticated(tokenString string) (string, *responses.ErrorResponse) {
	// Lấy secret key từ config
	secretKey := config.Config.JWTEmployeeSecret
	if secretKey == "" {
		log.Println("Warning: JWTEmployeeSecret is empty")
		return "", responses.NewErrorResponse(fiber.StatusInternalServerError, "Server misconfiguration: missing JWT secret")
	}

	// Kiểm tra token rỗng
	if tokenString == "" {
		log.Println("Warning: tokenString is empty")
		return "", responses.NewErrorResponse(fiber.StatusUnauthorized, "No token provided")
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil || !token.Valid {
		log.Println("Warning: invalid token -", err)
		return "", responses.NewErrorResponse(fiber.StatusUnauthorized, "Invalid token")
	}

	// Lấy claims từ token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("Warning: claims is not valid")
		return "", responses.NewErrorResponse(fiber.StatusUnauthorized, "Invalid token claims")
	}

	// Kiểm tra expiration
	if exp, ok := claims["exp"].(float64); !ok || time.Now().Unix() > int64(exp) {
		log.Println("Warning: token is expired or missing expiration")
		return "", responses.NewErrorResponse(fiber.StatusUnauthorized, "Token is expired")
	}

	// Lấy sub (user ID) từ claims
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		log.Println("Warning: invalid or missing sub in token")
		return "", responses.NewErrorResponse(fiber.StatusUnauthorized, "Invalid token payload")
	}

	// Trả về sub nếu thành công
	return sub, nil
}

// IsUserAuthenticated xác thực token của user và trả về sub (user ID) nếu hợp lệ
func IsUserAuthenticated(tokenString string) (string, *responses.ErrorResponse) {
	// Lấy secret key từ config
	secretKey := config.Config.JWTSecret
	if secretKey == "" {
		log.Println("Warning: JWTSecret is empty")
		return "", responses.NewErrorResponse(fiber.StatusInternalServerError, "Server misconfiguration: missing JWT secret")
	}

	// Kiểm tra token rỗng
	if tokenString == "" {
		log.Println("Warning: tokenString is empty")
		return "", responses.NewErrorResponse(fiber.StatusUnauthorized, "No token provided")
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil || !token.Valid {
		log.Println("Warning: invalid token -", err)
		return "", responses.NewErrorResponse(fiber.StatusUnauthorized, "Invalid token")
	}

	// Lấy claims từ token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("Warning: claims is not valid")
		return "", responses.NewErrorResponse(fiber.StatusUnauthorized, "Invalid token claims")
	}

	// Kiểm tra expiration
	if exp, ok := claims["exp"].(float64); !ok || time.Now().Unix() > int64(exp) {
		log.Println("Warning: token is expired or missing expiration")
		return "", responses.NewErrorResponse(fiber.StatusUnauthorized, "Token is expired")
	}

	// Lấy sub (user ID) từ claims
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		log.Println("Warning: invalid or missing sub in token")
		return "", responses.NewErrorResponse(fiber.StatusUnauthorized, "Invalid token payload")
	}

	// Trả về sub nếu thành công
	return sub, nil
}

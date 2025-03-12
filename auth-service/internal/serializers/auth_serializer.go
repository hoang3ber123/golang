package serializers

import (
	"auth-service/internal/db"
	"auth-service/internal/models"
	"auth-service/internal/responses"
	"auth-service/internal/services"
	"fmt"
	"time"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Define UserSignUpSerializer
type UserSignUpSerializer struct {
	Username    string  `json:"username" validate:"required"`
	Password    string  `json:"password" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Email       string  `json:"email" validate:"required,email"`
	Dob         any     `json:"dob" validate:"required"`
	Position    *string `json:"position,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty" validate:"max=11"`
	Contact     *string `json:"contact,omitempty"`
}

// Deserialize parses and validates input, including duplicate field check
func (s *UserSignUpSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse JSON vào struct
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}

	// Xác thực cơ bản với go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	// Kiểm tra định dạng ngày sinh (YYYY-MM-DD)
	dobStr, ok := s.Dob.(string)
	if !ok {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid dob type, must be a string")
	}
	parsedDob, err := time.Parse("2006-01-02", dobStr)
	if err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid dob format. Use YYYY-MM-DD (e.g., 2025-02-21)")
	}
	s.Dob = parsedDob

	// Kiểm tra trùng lặp Username
	var usernameExists bool
	if err := db.DB.Raw("SELECT EXISTS (SELECT 1 FROM users WHERE username = ?)", s.Username).Scan(&usernameExists).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error checking username: "+err.Error())
	}
	if usernameExists {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Username already exists")
	}

	// Kiểm tra trùng lặp Email đã xác minh
	var emailExists bool
	if err := db.DB.Raw("SELECT EXISTS (SELECT 1 FROM users WHERE email = ? AND is_email_verify = true)", s.Email).Scan(&emailExists).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error checking email: "+err.Error())
	}
	if emailExists {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Email already exists and verified")
	}

	return nil
}

// ToModel converts the serializer to a model
func (s *UserSignUpSerializer) ToModel() models.User {
	// send mail for user
	return models.User{
		Username:    s.Username,
		Password:    s.Password,
		Name:        s.Name,
		Email:       s.Email,
		Dob:         s.Dob.(time.Time),
		Position:    s.Position,
		PhoneNumber: s.PhoneNumber,
		Contact:     s.Contact,
	}
}

// Define Userlogin serializer
type UserLoginSerializer struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Deserialize parses and validates input, including duplicate field check
func (s *UserLoginSerializer) Login(c *fiber.Ctx) (*models.User, *responses.ErrorResponse) {
	var user models.User
	// Parse the incoming JSON into the serializer struct
	if err := c.BodyParser(s); err != nil {
		return nil, responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}
	// Basic validation with go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return nil, responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}
	// Custom validation: Check password compare
	result := db.DB.First(&user, "username = ?", s.Username)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, responses.NewErrorResponse(fiber.StatusBadRequest, "User not found")
		}
		return nil, responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+result.Error.Error())
	}

	// Check password correctness
	if !services.CheckPasswordHash(user.Password, s.Password) {
		return nil, responses.NewErrorResponse(fiber.StatusBadRequest, "Password or Username is incorrect")
	}

	// No error
	return &user, nil
}

// Deserialize parses and validates input, including duplicate field check
func (s *UserLoginSerializer) EmployeeLogin(c *fiber.Ctx) (*models.Employee, *responses.ErrorResponse) {
	var user models.Employee
	// Parse the incoming JSON into the serializer struct
	if err := c.BodyParser(s); err != nil {
		return nil, responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}
	// Basic validation with go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return nil, responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}
	// Custom validation: Check password compare
	result := db.DB.First(&user, "username = ?", s.Username)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, responses.NewErrorResponse(fiber.StatusBadRequest, "User not found")
		}
		return nil, responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+result.Error.Error())
	}

	// Check password correctness
	if !services.CheckPasswordHash(user.Password, s.Password) {
		return nil, responses.NewErrorResponse(fiber.StatusBadRequest, "Password or Username is incorrect")
	}

	// No error
	return &user, nil
}

// Define EmployeeSignUpSerializer
type EmployeeSignUpSerializer struct {
	Identity    string     `json:"identity" validate:"required"`
	Username    string     `json:"username" validate:"required"`
	Password    string     `json:"password" validate:"required"`
	Name        string     `json:"name" validate:"required"`
	Email       string     `json:"email" validate:"required,email"`
	Dob         string     `json:"dob" validate:"required"`
	Position    string     `json:"position" validate:"required"`
	PhoneNumber string     `json:"phone_number" validate:"max=11,required"`
	Contact     string     `json:"contact" validate:"required"`
	IsActive    bool       `json:"is_active,omitempty"`
	RoleID      *uuid.UUID `json:"role_id,omitempty" validate:"required"`
}

// Define EmployeeListSignUpSerializer
type EmployeeListSignUpSerializer struct {
	Employees []EmployeeSignUpSerializer `json:"employees" validate:"required,dive"`
}

func (s *EmployeeListSignUpSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	if err := c.BodyParser(&s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}

	// Basic validation with go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	return nil
}

func (s *EmployeeListSignUpSerializer) ToModel() ([]*models.Employee, error) {
	fmt.Println("To model")
	employees := make([]*models.Employee, len(s.Employees))
	for i, employee := range s.Employees {
		fmt.Println(employee)
		Dob, _ := time.Parse("2006-01-02", employee.Dob)
		employees[i] = &models.Employee{
			Identity:    employee.Identity,
			Username:    employee.Username,
			Password:    employee.Password,
			Name:        employee.Name,
			Email:       employee.Email,
			Dob:         Dob,
			Position:    employee.Position,
			PhoneNumber: employee.PhoneNumber,
			Contact:     employee.Contact,
			IsActive:    employee.IsActive,
			RoleID:      employee.RoleID,
		}
	}
	return employees, nil
}

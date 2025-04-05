package serializers

import (
	"fmt"
	"product-service/internal/responses"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CartProductAddSerializer struct {
	ID uuid.UUID `json:"id" validate:"required"`
}

func (s *CartProductAddSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}

	// Basic validation với go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	return nil
}

type CartProductRemoveSerializer struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive"`
}

func (s *CartProductRemoveSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}
	fmt.Println("ids:", s)
	// Basic validation với go-playground/validator
	fmt.Println("bắt đầu validate")
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}
	fmt.Println("validate thành công")

	return nil
}

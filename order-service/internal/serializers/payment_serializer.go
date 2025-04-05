package serializers

import (
	"order-service/internal/responses"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// type RefundPaymentSerializer struct {

type RefundPaymentSerializer struct {
	Reason string `json:"refund_reason" form:"refund_reason" validate:"required,oneof=duplicate fraudulent requested_by_customer"`
}

func (s *RefundPaymentSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}
	return nil
}

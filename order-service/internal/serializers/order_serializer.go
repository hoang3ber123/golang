package serializers

import (
	"order-service/internal/models"
	"order-service/internal/responses"

	"github.com/gofiber/fiber/v2"
)

// CreateOrderSerializer
type CreateOrderSerializer struct {
	Cart          map[string][]string `json:"cart" validate:"required,dive"`
	PaymentMethod string              `json:"payment_method" validate:"required"`
}

func (s *CreateOrderSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}
	if s.PaymentMethod == "" {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Payment method is required")
	}
	if s.PaymentMethod == "" {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Payment method is required")
	}
	if len(s.Cart) == 0 {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Cart is required and cannot be empty")
	}
	return nil
}

// OrderListResponseSerializer struct để serialize danh sách Order
type OrderListResponseSerializer struct {
	BaseResponseSerializer
	UserID        string  `json:"user_id"`
	PaymentMethod string  `json:"payment_method"`
	TransactionID string  `json:"transaction_id"`
	AmountPaid    float64 `json:"amount_paid"`
	PaymentStatus string  `json:"payment_status"`
}

// OrderListResponse serialize danh sách Order thành slice OrderListResponseSerializer
func OrderListResponse(instance *[]models.Order) []OrderListResponseSerializer {
	results := make([]OrderListResponseSerializer, len(*instance)) // Preallocate slice

	for i, val := range *instance {
		// Copy từng phần tử từ models.Order vào serializer
		results[i] = OrderListResponseSerializer{
			BaseResponseSerializer: BaseResponseSerializer{
				ID:        val.ID,
				CreatedAt: val.CreatedAt,
				UpdatedAt: val.UpdatedAt,
			},
			UserID:        val.UserID,
			PaymentMethod: val.PaymentMethod,
			TransactionID: val.TransactionID,
			AmountPaid:    val.AmountPaid,
			PaymentStatus: val.PaymentStatus,
		}
	}

	return results
}

// CreateOrderSerializer
type OrderDetailSerializer struct {
	Cart []models.Product `json:"cart" validate:"required,dive"`
}

type OrderDetailResponseSerializer struct {
	BaseResponseSerializer
	UserID        models.User      `json:"user"`
	PaymentMethod string           `json:"payment_method"`
	TransactionID string           `json:"transaction_id"`
	AmountPaid    float64          `json:"amount_paid"`
	PaymentStatus string           `json:"payment_status"`
	Products      []models.Product `json:"products" validate:"dive"`
}

func OrderDetailResponse(instance *models.Order) *OrderDetailResponseSerializer {
	// Gọi grpc lấy thông tin user
	// gọi grpc lấy thông tin của product
	// gọi gprc
	return &OrderDetailResponseSerializer{
		BaseResponseSerializer: BaseResponseSerializer{
			ID:        instance.ID,
			CreatedAt: instance.CreatedAt,
			UpdatedAt: instance.UpdatedAt,
		},
	}
}

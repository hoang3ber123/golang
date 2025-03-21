package services

import (
	"order-service/internal/db"
	"order-service/internal/models"
	"order-service/internal/responses"
	"sync"

	"github.com/gofiber/fiber/v2"
)

func CreateOrder(userID string, products []*models.Product, paymentMethod, transactionID string, amountPaid float64) *responses.ErrorResponse {
	order := models.Order{
		UserID:        userID,
		PaymentMethod: paymentMethod,
		TransactionID: transactionID,
		AmountPaid:    amountPaid,
		PaymentStatus: "processing",
	}

	tx := db.DB.Begin()
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return responses.NewErrorResponse(fiber.StatusInternalServerError, err.Error())
	}

	var orderDetailPool = sync.Pool{
		New: func() interface{} {
			return &models.OrderDetail{}
		},
	}

	orderDetails := make([]models.OrderDetail, len(products))
	for i, product := range products {
		orderDetail := orderDetailPool.Get().(*models.OrderDetail)
		orderDetail.OrderID = order.ID
		orderDetail.RelatedID = product.ID
		orderDetail.RelatedType = product.RelatedType
		orderDetail.TotalPrice = product.Price
		orderDetails[i] = *orderDetail
		orderDetailPool.Put(orderDetail)
	}

	if err := tx.Create(&orderDetails).Error; err != nil {
		tx.Rollback()
		return responses.NewErrorResponse(fiber.StatusInternalServerError, err.Error())
	}

	tx.Commit()
	return nil
}

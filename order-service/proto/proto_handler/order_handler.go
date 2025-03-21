package protohandler

import (
	"context"
	"fmt"
	"order-service/internal/db"
	"order-service/internal/models"
	"order-service/pagination"
	"time"

	"github.com/gofiber/fiber/v2"
	order_proto "github.com/hoang3ber123/proto-golang/order"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

// OrderServiceServer triển khai interface từ proto
type OrderServiceServer struct {
	order_proto.UnimplementedOrderServiceServer
}

// Gửi request tạo order Service
func (s *OrderServiceServer) CreateOrder(ctx context.Context, req *order_proto.CreateOrderRequest) (*order_proto.CreateOrderResponse, error) {
	transactionID := req.TransactionId
	paymentMethod := req.PaymentMethod

	switch paymentMethod {
	case "stripe":
		{
			if transactionID == "" {
				return &order_proto.CreateOrderResponse{
					IsCreated:  true,
					StatusCode: fiber.StatusCreated,
					Error:      "transaction ID is empty",
				}, nil
			}

			// Thêm tham số Expand để lấy đầy đủ thông tin PaymentIntent
			expand := "payment_intent"
			params := &stripe.CheckoutSessionParams{
				Params: stripe.Params{
					Expand: []*string{&expand}, // Expand PaymentIntent
				},
			}
			s, err := session.Get(transactionID, params)
			if err != nil {
				return &order_proto.CreateOrderResponse{
					IsCreated:  true,
					StatusCode: fiber.StatusCreated,
					Error:      err.Error(),
				}, nil
			}
			response := map[string]interface{}{
				"payment_intent_id": s.PaymentIntent.ID,
				"amount":            s.PaymentIntent.Amount,
				"currency":          s.PaymentIntent.Currency,
				"status":            s.PaymentIntent.Status,
				"metadata":          s.Metadata,
			}
			fmt.Println("body:", response)
			// duyệt mảng tạo
			// for i, product := range s.Metadata["product"] {

			// }
			// tạo product
			// services.CreateOrder()
		}
	}

	return &order_proto.CreateOrderResponse{
		IsCreated:  true,
		StatusCode: fiber.StatusCreated,
		Error:      "",
	}, nil
}

// Gửi request check xem sản phẩm đã mua chưa Service
func (s *OrderServiceServer) CheckBoughtProduct(ctx context.Context, req *order_proto.CheckBoughtProductRequest) (*order_proto.CheckBoughtProductResponse, error) {
	var isBought bool

	// Truy vấn kiểm tra sản phẩm đã mua chưa
	err := db.DB.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM order_details
			JOIN orders ON order_details.order_id = orders.id 
			WHERE orders.user_id = ? 
			AND orders.payment_status = 'success' 
			AND related_id = ? 
			AND related_type = ?
		)
	`, req.UserId, req.RelatedId, req.RelatedType).Scan(&isBought).Error

	if err != nil {
		return &order_proto.CheckBoughtProductResponse{
			IsBought:   false,
			StatusCode: fiber.StatusInternalServerError, // 500 - Lỗi server
			Error:      "Database query failed",
		}, nil
	}

	if !isBought {
		return &order_proto.CheckBoughtProductResponse{
			IsBought:   false,
			StatusCode: fiber.StatusBadRequest, // 400 - Người dùng chưa mua
			Error:      "Product not purchased",
		}, nil
	}

	return &order_proto.CheckBoughtProductResponse{
		IsBought:   true,
		StatusCode: fiber.StatusOK, // 200 - Thành công
		Error:      "",
	}, nil
}

// GetProductIDs xử lý yêu cầu lấy danh sách product IDs
func (s *OrderServiceServer) GetProductIDs(ctx context.Context, req *order_proto.GetProductIDsRequest) (*order_proto.GetProductIDsResponse, error) {
	// Khởi tạo query cơ bản
	query := db.DB.Model(&models.OrderDetail{}).
		Joins("JOIN orders ON orders.id = order_details.order_id").
		Select("order_details.related_id")

	// Xử lý các điều kiện lọc
	if req.UserId != "" {
		query = query.Where("orders.user_id = ?", req.UserId)
	}
	if req.OrderId != "" {
		query = query.Where("orders.id = ?", req.OrderId)
	}
	if req.RelatedType != "" {
		query = query.Where("order_details.related_type = ?", req.RelatedType)
	}

	// Kiểm tra và lọc payment_method
	if req.PaymentMethod != "" {
		if !models.AllowPaymentMethod[req.PaymentMethod] {
			return &order_proto.GetProductIDsResponse{
				ProductIds: nil,
				Pagination: nil,
				StatusCode: fiber.StatusBadRequest,
				Error:      "payment method invalid",
			}, nil
		}
		query = query.Where("orders.payment_method = ?", req.PaymentMethod)
	}

	// Kiểm tra và lọc payment_status
	if req.PaymentStatus != "" {
		if !models.AllowPaymentMethod[req.PaymentStatus] {
			return &order_proto.GetProductIDsResponse{
				ProductIds: nil,
				Pagination: nil,
				StatusCode: fiber.StatusBadRequest,
				Error:      "payment status invalid",
			}, nil
		}
		query = query.Where("orders.payment_status = ?", req.PaymentStatus)
	}

	// Lọc theo start_payment_day
	if req.StartPaymentDay != "" {
		startTime, err := time.Parse("2006-01-02", req.StartPaymentDay)
		if err != nil {
			return &order_proto.GetProductIDsResponse{
				ProductIds: nil,
				Pagination: nil,
				StatusCode: fiber.StatusBadRequest,
				Error:      "invalid start_payment_day format (use YYYY-MM-DD)",
			}, nil
		}
		query = query.Where("orders.updated_date >= ?", startTime)
	}

	// Lọc theo end_payment_day
	if req.EndPaymentDay != "" {
		endTime, err := time.Parse("2006-01-02", req.EndPaymentDay)
		if err != nil {
			return &order_proto.GetProductIDsResponse{
				ProductIds: nil,
				Pagination: nil,
				StatusCode: fiber.StatusBadRequest,
				Error:      "invalid end_payment_day format (use YYYY-MM-DD)",
			}, nil
		}
		query = query.Where("orders.updated_date <= ?", endTime)
	}

	// Lọc theo min_price
	if req.MinPrice > 0 {
		query = query.Where("order_details.total_price >= ?", req.MinPrice)
	}

	// Lọc theo max_price
	if req.MaxPrice > 0 {
		if req.MaxPrice < req.MinPrice {
			return &order_proto.GetProductIDsResponse{
				ProductIds: nil,
				Pagination: nil,
				StatusCode: fiber.StatusBadRequest,
				Error:      "max price must be higher than min price",
			}, nil
		}
		query = query.Where("order_details.total_price <= ?", req.MaxPrice)
	}

	// Sắp xếp: ưu tiên payment_day_order trước price_order
	if req.PriceOrder != "" {
		if req.PriceOrder == "asc" || req.PriceOrder == "desc" {
			query = query.Order("order_details.total_price " + req.PriceOrder)
		}
	} else if req.PaymentDayOrder != "" {
		if req.PaymentDayOrder == "asc" || req.PaymentDayOrder == "desc" {
			query = query.Order("orders.updated_at " + req.PaymentDayOrder)
		}
	}

	// Giả lập fiber.Ctx để dùng Paginate (trong gRPC cần mock hoặc bỏ)
	// Ở đây tôi giả định bạn có cách tích hợp fiber.Ctx, nếu không thì cần điều chỉnh
	var productIDs []string
	pagination, errResp := pagination.Paginate(int(req.Page), int(req.PageSize), query, &productIDs)
	if errResp != nil {
		return &order_proto.GetProductIDsResponse{
			ProductIds: nil,
			Pagination: nil,
			StatusCode: int32(errResp.StatusCode),
			Error:      errResp.Message,
		}, nil
	}

	// Trả về response
	return &order_proto.GetProductIDsResponse{
		ProductIds: productIDs,
		Pagination: &order_proto.Pagination{
			Page:      int32(pagination.Page),
			PageSize:  int32(pagination.PageSize),
			Total:     int32(pagination.Total),
			TotalPage: int32(pagination.TotalPage),
		},
		StatusCode: fiber.StatusOK,
		Error:      "",
	}, nil
}

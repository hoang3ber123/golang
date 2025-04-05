package grpcclient

import (
	"context"
	"log"
	"product-service/config"
	"product-service/internal/models"
	"product-service/internal/responses"
	"product-service/internal/serializers"
	"product-service/pagination"
	"time"

	"github.com/gofiber/fiber/v2"
	pb "github.com/hoang3ber123/proto-golang/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

var connOrderClient *grpc.ClientConn
var orderClient pb.OrderServiceClient

// Hàm khởi tạo gRPC Client với cơ chế tái kết nối
func InitOrderGRPCClient() {
	var err error
	connOrderClient, err = grpc.Dial(
		config.Config.OrderServiceHost+":"+config.Config.GRPCOrderPort,
		grpc.WithInsecure(),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  1.0 * time.Second, // Lần đầu retry sau 1s
				Multiplier: 1.6,               // Mỗi lần sau đó delay gấp 1.6 lần
				MaxDelay:   10 * time.Second,  // Tối đa 10s
			},
			MinConnectTimeout: 5 * time.Second, // Thời gian tối thiểu để kết nối lại
		}),
	)
	if err != nil {
		log.Fatalf("Can't not connect to gRPC server: %v", err)
	}

	orderClient = pb.NewOrderServiceClient(connOrderClient)
	log.Println("Connected gRPC Server!")
}

func CheckBoughtRequest(UserID, RelatedID, RelatedType string) *responses.ErrorResponse {
	// Tạo context với timeout 3 giây
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Gửi request đến order service
	res, err := orderClient.CheckBoughtProduct(ctx, &pb.CheckBoughtProductRequest{
		UserId:      UserID,
		RelatedId:   RelatedID,
		RelatedType: RelatedType,
	})

	if err != nil {
		log.Printf("Error calling order service: %s", err.Error())
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Order service error: "+err.Error())
	}

	// Kiểm tra lỗi trả về từ gRPC response
	if res.Error != "" {
		return responses.NewErrorResponse(int(res.StatusCode), res.Error)
	}

	return nil
}

func GetProductIDs(query *serializers.ProductQuerySerializer) ([]string, *pagination.Pagination, *responses.ErrorResponse) {
	// Tạo context với timeout 3 giây
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Gửi request đến order service
	res, err := orderClient.GetProductIDs(ctx, &pb.GetProductIDsRequest{
		OrderId:         query.OrderID,
		UserId:          query.UserID,
		RelatedType:     query.RelatedType,
		PaymentMethod:   query.PaymentMethod,
		PaymentStatus:   query.PaymentStatus,
		Page:            query.Page,
		PageSize:        query.PageSize,
		MaxPrice:        query.MaxPrice,
		MinPrice:        query.MinPrice,
		EndPaymentDay:   query.EndPaymentDay,
		StartPaymentDay: query.StartPaymentDay,
		PaymentDayOrder: query.PaymentDayOrder,
		PriceOrder:      query.PriceOrder,
	})
	if err != nil {
		log.Printf("Error calling Order Service: %s", err.Error())
		return nil, nil, responses.NewErrorResponse(fiber.StatusInternalServerError, "Order service error: "+err.Error())
	}

	if res.Error != "" {
		return nil, nil, responses.NewErrorResponse(int(res.StatusCode), res.Error)
	}
	return res.ProductIds, &pagination.Pagination{
		Page:      int(res.Pagination.Page),
		PageSize:  int(res.Pagination.PageSize),
		Total:     int(res.Pagination.Total),
		TotalPage: int(res.Pagination.TotalPage),
	}, nil
}

// lấy danh sách product id đã mua
func GetAllProductIDs(c *fiber.Ctx, relatedType string) ([]string, *responses.ErrorResponse) {
	// Tạo context với timeout 3 giây
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	userInterface := c.Locals("user")
	if userInterface == nil {
		return nil, responses.NewErrorResponse(fiber.StatusInternalServerError, "Error when get user info")
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return nil, responses.NewErrorResponse(fiber.StatusInternalServerError, "Error when get user info")
	}
	// Gửi request đến order service
	res, err := orderClient.GetAllProductIDs(ctx, &pb.GetAllProductIDsRequest{
		UserId:      user.ID.String(),
		RelatedType: relatedType,
	})
	if err != nil {
		log.Printf("Error calling Order Service: %s", err.Error())
		return nil, responses.NewErrorResponse(fiber.StatusInternalServerError, "Order service error: "+err.Error())
	}

	if res.Error != "" {
		return nil, responses.NewErrorResponse(int(res.StatusCode), res.Error)
	}
	return res.ProductIds, nil
}

// Hàm đóng kết nối khi không cần nữa (nếu cần)
func CloseOrderGRPCClient() {
	if connOrderClient != nil {
		connOrderClient.Close()
	}
}

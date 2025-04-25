package grpcclient

import (
	"context"
	"log"
	"order-service/config"
	"order-service/internal/models"
	"order-service/internal/responses"
	"time"

	"github.com/gofiber/fiber/v2"
	pb "github.com/hoang3ber123/proto-golang/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

var connProductClient *grpc.ClientConn
var productClient pb.ProductServiceClient

// Hàm khởi tạo gRPC Client với cơ chế tái kết nối
func InitProductGRPCClient() {
	var err error
	connProductClient, err = grpc.Dial(
		config.Config.ProductServiceHost+":"+config.Config.GRPCProductPort,
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

	productClient = pb.NewProductServiceClient(connProductClient)
	log.Println("Connected gRPC Server!")
}

func GetProductsInfo(productsInfoRequest []*pb.ProductsInfoRequest) ([]*models.Product, *responses.ErrorResponse) {
	// Tạo context với timeout 3 giây
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Gửi request đến product service
	// duyệt cart của người dùng gửi lên thành []*pb.CartItemRequest gửi lên product grpc server
	res, err := productClient.GetProductsInfo(ctx, &pb.GetProductsInfoRequest{
		Products: productsInfoRequest,
	})

	if err != nil {
		log.Printf("Error calling product service: %s", err.Error())
		return nil, responses.NewErrorResponse(fiber.StatusInternalServerError, "Product service error: "+err.Error())
	}

	// Kiểm tra lỗi trả về từ gRPC response
	if res.Error != "" {
		return nil, responses.NewErrorResponse(int(res.StatusCode), res.Error)
	}

	// Tạo danh sách Product từ response gRPC
	products := make([]*models.Product, len(res.Products))
	for index, product := range res.Products {
		products[index] = &models.Product{
			ID:          product.Id,
			Slug:        product.Slug,
			Title:       product.Title,
			Image:       product.Image,
			RelatedType: product.RelatedType,
			Price:       product.Price,
		}
	}
	return products, nil
}

func ClearCartAfterCheckout(productsInfoRequest []*pb.ProductsInfoRequest, user_id string) *responses.ErrorResponse {
	// Tạo context với timeout 3 giây
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Gửi request đến product service
	res, err := productClient.ClearCartAfterCheckout(ctx, &pb.ClearCartAfterCheckoutRequest{
		Products: productsInfoRequest,
		User:     user_id,
	})

	if err != nil {
		log.Printf("Error calling product service: %s", err.Error())
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Product service error: "+err.Error())
	}

	// Kiểm tra lỗi trả về từ gRPC response
	if res.Error != "" {
		return responses.NewErrorResponse(int(res.StatusCode), res.Error)
	}

	return nil
}

// Hàm đóng kết nối khi không cần nữa (nếu cần)
func CloseProductGRPCClient() {
	if connProductClient != nil {
		connProductClient.Close()
	}
}

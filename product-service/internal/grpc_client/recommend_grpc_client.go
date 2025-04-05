package grpcclient

import (
	"context"
	"encoding/json"
	"log"
	"product-service/config"
	"product-service/internal/responses"
	"product-service/internal/services"
	"time"

	"github.com/gofiber/fiber/v2"
	pb "github.com/hoang3ber123/proto-golang/recommend"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

var connRecommendClient *grpc.ClientConn
var recommendClient pb.RecommendServiceClient

// Hàm khởi tạo gRPC Client với cơ chế tái kết nối
func InitRecommendGRPCClient() {
	var err error
	connRecommendClient, err = grpc.Dial(
		config.Config.RecommendServiceHost+":"+config.Config.GRPCRecommendPort,
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

	recommendClient = pb.NewRecommendServiceClient(connRecommendClient)
	log.Println("Connected gRPC Server!")
}

// Hàm gọi gRPC để lấy danh sách category IDs dựa trên query và categories
func GetRecommendCategoryIDs(query string) ([]string, *responses.ErrorResponse) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Tạo slice chứa con trỏ *pb.Category
	var pbCategories []*pb.Category
	// lấy tất cả category từ redis
	categoriesJSON, errRedis := services.GetCategoriesFromRedis()
	if errRedis != nil {
		return nil, errRedis
	}
	// Nếu có data trong Redis, deserialize
	if categoriesJSON != "" {
		if err := json.Unmarshal([]byte(categoriesJSON), &pbCategories); err != nil {
			return nil, responses.NewErrorResponse(
				fiber.StatusInternalServerError,
				"Error when recommend category:"+err.Error())
		}
	}
	// Tạo request cho GetRecommendCategoryIDs
	req := &pb.GetRecommendCategoryIDsRequest{
		Query:      query,
		Categories: pbCategories, // Truyền slice các con trỏ vào
	}

	// Gửi request đến gRPC service
	var res *pb.GetRecommendCategoryIDsResponse
	res, err := recommendClient.GetRecommendCategoryIDs(ctx, req)
	if err != nil {
		log.Printf("Error calling GetRecommendCategoryIDs: %s", err.Error())
		return nil, responses.NewErrorResponse(fiber.StatusInternalServerError, "Recommend service error: "+err.Error())
	}

	// Kiểm tra lỗi trả về từ gRPC response
	if res == nil || len(res.CategoryIds) == 0 {
		return nil, responses.NewErrorResponse(fiber.StatusNotFound, "No recommended categories found")
	}

	// Trả về danh sách Category IDs
	return res.CategoryIds, nil
}

// Hàm gọi gRPC để lấy danh sách category IDs dựa trên query và categories
func GetRecommendProductIDs(recommendProduct []*pb.ProductRecommend, clickProducts []*pb.ClickDetail, viewProducts []*pb.ViewProduct, boughtProducts []string) ([]string, *responses.ErrorResponse) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Tạo request cho GetRecommendCategoryIDs
	req := &pb.GetRecommendProductIDsRequest{
		Products:       recommendProduct,
		ClickDetails:   clickProducts,
		ViewProducts:   viewProducts, // Truyền slice các con trỏ vào
		BoughtProducts: boughtProducts,
	}

	// Gửi request đến gRPC service
	var res *pb.GetRecommendProductIDsResponse
	res, err := recommendClient.GetRecommendProductIDs(ctx, req)
	if err != nil {
		log.Printf("Error calling GetRecommendProductIDs: %s", err.Error())
		return nil, responses.NewErrorResponse(fiber.StatusInternalServerError, "Recommend service error: "+err.Error())
	}

	// Kiểm tra lỗi trả về từ gRPC response
	if res == nil || len(res.ProductIds) == 0 {
		return nil, responses.NewErrorResponse(fiber.StatusNotFound, "No recommended categories found")
	}

	// Trả về danh sách Category IDs
	return res.ProductIds, nil
}

// Hàm đóng kết nối khi không cần nữa (nếu cần)
func CloseRecommendGRPCClient() {
	if connRecommendClient != nil {
		connRecommendClient.Close()
	}
}

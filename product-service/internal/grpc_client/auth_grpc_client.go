package grpcclient

import (
	"context"
	"log"
	"product-service/config"
	"product-service/internal/responses"
	"time"

	"github.com/gofiber/fiber/v2"
	pb "github.com/hoang3ber123/proto-golang/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

var connAuthClient *grpc.ClientConn
var authClient pb.AuthServiceClient

// Hàm khởi tạo gRPC Client với cơ chế tái kết nối
func InitAuthGRPCClient() {
	var err error
	connAuthClient, err = grpc.Dial(
		config.Config.AuthServiceHost+":"+config.Config.GRPCAuthPort,
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

	authClient = pb.NewAuthServiceClient(connAuthClient)
	log.Println("Connected auth gRPC Server!")
}

// Hàm gọi API AuthEmployeeRequest
// trả về nil nếu xác thực và kiểm tra quyền thành công
// AuthEmployeeRequest gọi API xác thực và kiểm tra quyền
// Trả về nil nếu xác thực thành công, ngược lại trả về *responses.ErrorResponse
type EmployeeInfo struct {
	ID          string
	Username    string
	RoleTitle   string
	Email       string
	Identity    string
	Name        string
	PhoneNumber string
	IsActive    bool
}

func AuthEmployeeRequest(token string, allowedRoles []string) (*EmployeeInfo, *responses.ErrorResponse) {
	// Tạo context với timeout 3 giây
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Gửi request đến auth service
	res, err := authClient.AuthenticateEmployee(ctx, &pb.AuthEmployeeRequest{
		Token: token,
		Role:  allowedRoles,
	})

	if err != nil {
		log.Printf("Error calling auth service: %s", err.Error())
		return nil, responses.NewErrorResponse(fiber.StatusInternalServerError, "Authentication service error: "+err.Error())
	}

	// Kiểm tra lỗi trả về từ gRPC response
	if res.Error != "" {
		return nil, responses.NewErrorResponse(int(res.StatusCode), res.Error)
	}

	// Kiểm tra nếu `res.User` là nil
	if res.User == nil {
		return nil, responses.NewErrorResponse(fiber.StatusUnauthorized, "User not found or unauthorized")
	}

	// Tạo `User` từ response gRPC
	user := &EmployeeInfo{
		ID:          res.User.Id,
		Username:    res.User.Username,
		RoleTitle:   res.User.RoleTitle,
		Email:       res.User.Email,
		Identity:    res.User.Identity,
		Name:        res.User.Name,
		PhoneNumber: res.User.PhoneNumber,
		IsActive:    res.User.IsActive,
	}

	return user, nil
}

// Hàm gọi API AuthUserRequest
// trả về nil nếu xác thực và kiểm tra quyền thành công
// AuthUserRequest gọi API xác thực và kiểm tra quyền
// Trả về nil nếu xác thực thành công, ngược lại trả về *responses.ErrorResponse
type UserInfo struct {
	ID       string
	Username string
	Email    string
	Name     string
	IsActive bool
}

func AuthUserRequest(token string) (*UserInfo, *responses.ErrorResponse) {
	// Tạo context với timeout 3 giây
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Gửi request đến auth service
	res, err := authClient.AuthenticateUser(ctx, &pb.AuthUserRequest{
		Token: token,
	})

	if err != nil {
		log.Printf("Error calling auth service: %s", err.Error())
		return nil, responses.NewErrorResponse(fiber.StatusInternalServerError, "Authentication service error: "+err.Error())
	}

	// Kiểm tra lỗi trả về từ gRPC response
	if res.Error != "" {
		return nil, responses.NewErrorResponse(int(res.StatusCode), res.Error)
	}

	// Kiểm tra nếu `res.User` là nil
	if res.User == nil {
		return nil, responses.NewErrorResponse(fiber.StatusUnauthorized, "User not found or unauthorized")
	}

	// Tạo `User` từ response gRPC
	user := &UserInfo{
		ID:       res.User.Id,
		Username: res.User.Username,
		Email:    res.User.Email,
		Name:     res.User.Name,
	}

	return user, nil
}

// Hàm đóng kết nối khi không cần nữa (nếu cần)
func CloseAuthGRPCClient() {
	if connAuthClient != nil {
		connAuthClient.Close()
	}
}

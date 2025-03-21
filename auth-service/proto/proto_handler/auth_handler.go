package protohandler

import (
	"auth-service/internal/db"
	"auth-service/internal/models"
	"auth-service/internal/services"
	"context"
	"log"
	"slices"

	auth_proto "github.com/hoang3ber123/proto-golang/auth"

	"github.com/google/uuid"
)

// AuthServiceServer triển khai interface từ proto
type AuthServiceServer struct {
	auth_proto.UnimplementedAuthServiceServer
}

// Xác thực token từ Product Service
func (s *AuthServiceServer) AuthenticateEmployee(ctx context.Context, req *auth_proto.AuthEmployeeRequest) (*auth_proto.AuthEmployeeResponse, error) {
	tokenString := req.Token // get token
	// Check employee authenticated
	id, err := services.IsEmployeeAuthenticated(tokenString)
	if err != nil {
		return &auth_proto.AuthEmployeeResponse{
			IsAuthenticated: false,
			Error:           err.Message,
			StatusCode:      500,
			User:            nil,
		}, nil
	}
	// Find user
	var user models.Employee
	db.DB.Joins("Role").First(&user, "employees.id = ?", id)
	// check if can find employee
	if user.ID == uuid.Nil {
		log.Println("Warning: Can not find employee")
		return &auth_proto.AuthEmployeeResponse{
			IsAuthenticated: false,
			Error:           "Unauthorized",
			StatusCode:      401,
			User:            nil,
		}, nil
	}
	// check role
	if !slices.Contains(req.Role, user.Role.Title) {
		return &auth_proto.AuthEmployeeResponse{
			IsAuthenticated: false,
			Error:           "Forbiden",
			StatusCode:      403,
			User:            nil,
		}, nil
	}
	// check if user is active
	if !user.IsActive {
		return &auth_proto.AuthEmployeeResponse{
			IsAuthenticated: false,
			Error:           "Blocked/Unactivated",
			StatusCode:      403,
			User:            nil,
		}, nil
	}

	return &auth_proto.AuthEmployeeResponse{
		IsAuthenticated: true,
		StatusCode:      200,
		Error:           "",
		User: &auth_proto.Employee{
			Id:          user.ID.String(),
			Username:    user.Username,
			RoleTitle:   user.Role.Title,
			Email:       user.Email,
			Identity:    user.Identity,
			Name:        user.Name,
			PhoneNumber: user.PhoneNumber,
			IsActive:    user.IsActive,
		},
	}, nil
}

// Xác thực token từ Product Service
func (s *AuthServiceServer) AuthenticateUser(ctx context.Context, req *auth_proto.AuthUserRequest) (*auth_proto.AuthUserResponse, error) {
	tokenString := req.Token // get token
	// Check employee authenticated
	id, err := services.IsUserAuthenticated(tokenString)
	if err != nil {
		return &auth_proto.AuthUserResponse{
			IsAuthenticated: false,
			Error:           err.Message,
			StatusCode:      500,
			User:            nil,
		}, nil
	}
	// Find user
	var user models.User
	db.DB.First(&user, "id = ?", id)
	// check if can find employee
	if user.ID == uuid.Nil {
		return &auth_proto.AuthUserResponse{
			IsAuthenticated: false,
			Error:           "Unauthorized",
			StatusCode:      401,
			User:            nil,
		}, nil
	}
	// check if user is active
	if !user.IsEmailVerify {
		return &auth_proto.AuthUserResponse{
			IsAuthenticated: false,
			Error:           "You still not verify your email",
			StatusCode:      401,
			User:            nil,
		}, nil
	}

	// check if user is active
	if !user.IsActive {
		return &auth_proto.AuthUserResponse{
			IsAuthenticated: false,
			Error:           "Blocked/Unactivated",
			StatusCode:      403,
			User:            nil,
		}, nil
	}
	return &auth_proto.AuthUserResponse{
		IsAuthenticated: true,
		StatusCode:      200,
		Error:           "",
		User: &auth_proto.User{
			Id:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
			Name:     user.Name,
		},
	}, nil
}

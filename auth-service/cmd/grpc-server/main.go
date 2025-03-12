package main

import (
	"auth-service/internal/db"
	protohandler "auth-service/proto/proto_handler"
	"log"
	"net"

	auth_proto "github.com/hoang3ber123/proto-golang/auth"
	"google.golang.org/grpc"
)

func init() {
	db.InitDB()
	// initialize.InitializingDatabase()
	// initialize.ConnectToApiGateway()
}

func main() {
	// Khởi tạo gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Can't not listen: %v", err)
	}

	s := grpc.NewServer()
	auth_proto.RegisterAuthServiceServer(s, &protohandler.AuthServiceServer{})

	log.Println("auth-service is running on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Can not run server server: %v", err)
	}
}

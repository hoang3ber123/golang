package main

import (
	"order-service/config"
	"order-service/initialize"
	"order-service/internal/db"
	grpcclient "order-service/internal/grpc_client"
	"order-service/internal/routes"
	protohandler "order-service/proto/proto_handler"

	order_proto "github.com/hoang3ber123/proto-golang/order"
	"github.com/stripe/stripe-go/v76"

	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"google.golang.org/grpc"
)

func init() {
	db.InitDB()
	if config.Config.SystemStatus == "docker" {
		initialize.ConnectToApiGateway()
	}
	grpcclient.InitAuthGRPCClient()
	grpcclient.InitProductGRPCClient()
}

var grpcServer *grpc.Server // Biến toàn cục để quản lý gRPC server
// Start gRPC Server
func startGRPCServer() {
	grpcPort := config.Config.GRPCPort
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Can't listen on port %s: %v", grpcPort, err)
	}

	grpcServer = grpc.NewServer()
	order_proto.RegisterOrderServiceServer(grpcServer, &protohandler.OrderServiceServer{})

	log.Println("🚀 gRPC order-service is running on port " + grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to run gRPC server: %v", err)
	}
}

// Start Fiber Server
func startFiberServer() *fiber.App {
	app := fiber.New(fiber.Config{
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
		IdleTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})
	// Stripe setting
	stripe.Key = config.Config.StripeSecretKey

	// Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     config.Config.AllowHost,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type, orderorization, api-key",
		AllowCredentials: true,
		ExposeHeaders:    "Set-Cookie",
	}))

	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: config.Config.EncryptCookieKey,
	}))

	// Setup routes
	routes.SetupRoutes(app)

	go func() {
		if err := app.Listen(":" + config.Config.HTTPPort); err != nil {
			log.Fatalf("Failed to start Fiber server: %v", err)
		}
	}()

	return app
}

// Graceful Shutdown Function
func shutdownServers(app *fiber.App) {
	// Shutdown Fiber server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("❌ Error during Fiber shutdown: %v", err)
	} else {
		log.Println("✅ Fiber server shut down successfully")
	}

	// Shutdown gRPC server
	if grpcServer != nil {
		grpcServer.GracefulStop()
		log.Println("✅ gRPC server shut down successfully")
	}
	// Close connected to auth service
	grpcclient.CloseAuthGRPCClient()
	log.Println("gRPC auth client connection closed")
	// Close connected to product service
	grpcclient.CloseProductGRPCClient()
	log.Println("gRPC product client connection closed")
	// Close database connection
	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Printf("❌ Error accessing SQL DB: %v", err)
	} else {
		if err := sqlDB.Close(); err != nil {
			log.Printf("❌ Error closing database: %v", err)
		} else {
			log.Println("✅ Database shut down successfully")
		}
	}

	log.Println("✅ Server gracefully shut down")
}

func main() {
	// Start gRPC server in a goroutine
	go startGRPCServer()

	// Start Fiber server
	app := startFiberServer()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("🛑 Shutdown signal received, initiating graceful shutdown...")

	// Shutdown all servers and database
	shutdownServers(app)
}

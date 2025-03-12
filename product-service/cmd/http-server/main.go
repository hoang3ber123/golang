package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"product-service/config"
	"product-service/initialize"
	"product-service/internal/db"
	grpcclient "product-service/internal/grpc_client"
	"product-service/internal/routes"

	"syscall"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
)

func init() {
	db.InitDB()
	grpcclient.InitGRPCClient()
	initialize.ConnectToApiGateway()
	db.InitRedis()
}

func main() {

	app := fiber.New(fiber.Config{
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
		IdleTimeout:  5 * time.Second, // Max time to wait for in-flight requests
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	// Middleware CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     config.Config.AllowHost, // Chỉ định các domain được phép
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type, Authorization, api-key",
		AllowCredentials: true,         // Cho phép gửi cookie
		ExposeHeaders:    "Set-Cookie", // Để client đọc được cookie từ response
	}))

	// Middleware Encrypt Cookie
	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: config.Config.EncryptCookieKey,
	}))

	// Setup routes
	routes.SetupRoutes(app)

	// Create a channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start Fiber server in a goroutine
	go func() {
		if err := app.Listen(":" + config.Config.HTTPPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, initiating graceful shutdown...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown Fiber server
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	// Close database connection
	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Printf("Error accessing underlying SQL DB: %v", err)
	} else {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database shut down successfully")
		}
	}

	// Close grpc connection
	grpcclient.CloseGRPCClient()
	log.Println("gRPC client connection closed")

	log.Println("Server gracefully shut down")
}

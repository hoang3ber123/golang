package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var Config struct {
	// Base config
	SystemStatus       string
	BasePath           string
	ServiceName        string
	ServiceRoute       string
	ServiceURL         string
	ServicePath        string
	APIKey             string
	APIGatewayHost     string
	AllowHost          string
	HTTPPort           string
	GRPCPort           string
	GRPCAuthPort       string
	AuthServiceHost    string
	GRPCProductPort    string
	ProductServiceHost string

	// Database config
	DatabaseName     string
	DatabaseUsername string
	DatabaseHost     string
	DatabasePort     string
	DatabasePassword string
	DatabaseURL      string

	// Security
	EncryptCookieKey string

	// SMTP config
	SMTPUsername string
	SMTPPassword string
	SMTPHost     string
	SMTPServer   string
	SMTPPort     string

	// Vstorage config
	VstorageProjectID     string
	VstorageAuthURL       string
	VstorageBaseURL       string
	VstorageContainerName string
	VstorageClientSecret  string
	VstorageClientID      string
	VstorageSwiftUsername string
	VstorageSwiftPassword string

	//Redis config
	RedisIndex    string
	RedisProtocol string
	RedisPassword string
	RedisHost     string
	RedisPort     string

	//Stripe
	StripeSecretKey string
	StripePublicKey string
}

// Load từng file .env theo đúng mục đích
func init() {
	log.Println("Loading environment variables...")
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting working directory:", err)
	} else {
		fmt.Println("Current working directory:", dir)
	}
	// Load từng file .env theo đúng mục đích
	Config.SystemStatus = "docker"
	if Config.SystemStatus == "docker" {
		loadEnvFile(dir + "/env/docker/.env")
		loadEnvFile(dir + "/env/docker/.env.database")
		loadEnvFile(dir + "/env/docker/.env.redis")
		loadEnvFile(dir + "/env/docker/.env.vstorage")
		loadEnvFile(dir + "/env/docker/.env.stripe")
	} else if Config.SystemStatus == "local" {
		loadEnvFile(dir + "/env/local/.env")
		loadEnvFile(dir + "/env/local/.env.database")
		loadEnvFile(dir + "/env/local/.env.redis")
		loadEnvFile(dir + "/env/local/.env.vstorage")
		loadEnvFile(dir + "/env/local/.env.stripe")
	}
	// System setting
	Config.BasePath = dir
	Config.AllowHost = os.Getenv("ALLOW_HOST")
	Config.HTTPPort = os.Getenv("HTTP_PORT")
	Config.GRPCAuthPort = os.Getenv("GRPC_AUTH_PORT")
	Config.AuthServiceHost = os.Getenv("AUTH_SERVICE_HOST")
	Config.GRPCProductPort = os.Getenv("GRPC_PRODUCT_PORT")
	Config.ProductServiceHost = os.Getenv("PRODUCT_SERVICE_HOST")
	Config.GRPCPort = os.Getenv("GRPC_PORT")
	// Database setting
	Config.DatabaseURL = os.Getenv("DATABASE_URL")
	Config.DatabaseName = os.Getenv("DATABASE_NAME")
	Config.DatabaseUsername = os.Getenv("DATABASE_USERNAME")
	Config.DatabaseHost = os.Getenv("DATABASE_HOST")
	Config.DatabasePort = os.Getenv("DATABASE_PORT")
	Config.DatabasePassword = os.Getenv("DATABASE_PASSWORD")
	// JWT setting
	Config.EncryptCookieKey = os.Getenv("ENCRYPT_COOKIE_KEY")
	// API gateway setting
	Config.APIGatewayHost = os.Getenv("API_GATEWAY_HOST")
	Config.ServiceName = os.Getenv("SERVICE_NAME")
	Config.ServiceRoute = os.Getenv("SERVICE_ROUTE")
	Config.ServiceURL = os.Getenv("SERVICE_URL")
	Config.ServicePath = os.Getenv("SERVICE_PATH")
	Config.APIKey = os.Getenv("API_KEY")
	// SMPT setting
	Config.SMTPUsername = os.Getenv("SMTP_USERNAME")
	Config.SMTPPassword = os.Getenv("SMTP_PASSWORD")
	Config.SMTPServer = os.Getenv("SMTP_SERVER")
	Config.SMTPPort = os.Getenv("SMTP_PORT")
	// Vstorage setting
	Config.VstorageAuthURL = os.Getenv("AUTH_URL")
	Config.VstorageBaseURL = os.Getenv("BASE_URL")
	Config.VstorageClientID = os.Getenv("CLIENT_ID")
	Config.VstorageClientSecret = os.Getenv("CLIENT_SECRECT")
	Config.VstorageContainerName = os.Getenv("CONTAINER_NAME")
	Config.VstorageProjectID = os.Getenv("PROJECT_ID")
	Config.VstorageSwiftPassword = os.Getenv("SWIFT_PASSWORD")
	Config.VstorageSwiftUsername = os.Getenv("SWIFT_USERNAME")
	// Redis setting
	Config.RedisIndex = os.Getenv("REDIS_INDEX")
	Config.RedisProtocol = os.Getenv("REDIS_PROTOCOL")
	Config.RedisPassword = os.Getenv("REDIS_PASSWORD")
	Config.RedisHost = os.Getenv("REDIS_HOST")
	Config.RedisPort = os.Getenv("REDIS_PORT")
	//Stripe setting
	Config.StripePublicKey = os.Getenv("STRIPE_PUBLIC_KEY")
	Config.StripeSecretKey = os.Getenv("STRIPE_SECRET_KEY")
}

func loadEnvFile(filepath string) {
	err := godotenv.Load(filepath)
	if err != nil {
		log.Printf("Warning: Could not load %s file", filepath)
	}
}

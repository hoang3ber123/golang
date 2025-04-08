package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var Config struct {
	SystemStatus      string // dùng để setting: local, docker
	BasePath          string
	DatabaseName      string
	DatabaseUsername  string
	DatabaseHost      string
	DatabasePort      string
	DatabasePassword  string
	DatabaseURL       string
	EncryptCookieKey  string
	JWTSecret         string
	JWTEmployeeSecret string
	JWTSecretMail     string
	SMTPUsername      string
	SMTPServer        string
	SMTPPort          string
	SMTPPassword      string
	SMTPHost          string
	APIGatewayHost    string
	ServiceName       string
	ServiceRoute      string
	ServiceUrl        string
	ServicePath       string
	APIKey            string
	AllowHost         string
	HTTPPort          string
	GRPCPort          string
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
	// Load từng file .env theo đúng mục đích
	Config.SystemStatus = "local"
	if Config.SystemStatus == "docker" {
		loadEnvFile(dir + "/env/docker/.env")
		loadEnvFile(dir + "/env/docker/.env.database")
		loadEnvFile(dir + "/env/docker/.env.redis")
		loadEnvFile(dir + "/env/docker/.env.vstorage")
	} else if Config.SystemStatus == "local" {
		loadEnvFile(dir + "/env/local/.env")
		loadEnvFile(dir + "/env/local/.env.database")
		loadEnvFile(dir + "/env/local/.env.redis")
		loadEnvFile(dir + "/env/local/.env.vstorage")
	}
	// System setting
	Config.BasePath = dir
	Config.AllowHost = os.Getenv("ALLOW_HOST")
	Config.HTTPPort = os.Getenv("HTTP_PORT")
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
	Config.JWTSecret = os.Getenv("JWT_SECRET")
	Config.JWTSecretMail = os.Getenv("JWT_SECRET_MAIL")
	Config.JWTEmployeeSecret = os.Getenv("JWT_EMPLOYEE_SECRET")
	// API gateway setting
	Config.APIGatewayHost = os.Getenv("API_GATEWAY_HOST")
	Config.ServiceName = os.Getenv("SERVICE_NAME")
	Config.ServiceRoute = os.Getenv("SERVICE_ROUTE")
	Config.ServiceUrl = os.Getenv("SERVICE_URL")
	Config.ServicePath = os.Getenv("SERVICE_PATH")
	Config.APIKey = os.Getenv("API_KEY")
	// SMPT setting
	Config.SMTPUsername = os.Getenv("SMTP_USERNAME")
	Config.SMTPPassword = os.Getenv("SMTP_PASSWORD")
	Config.SMTPServer = os.Getenv("SMTP_SERVER")
	Config.SMTPPort = os.Getenv("SMTP_PORT")
}

func loadEnvFile(filepath string) {
	err := godotenv.Load(filepath)
	if err != nil {
		log.Printf("Warning: Could not load %s file", filepath)
	}
}

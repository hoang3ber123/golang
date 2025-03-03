package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var Config struct {
	DatabaseURL       string
	EncryptCookieKey  string
	JWTSecret         string
	JWTEmployeeSecret string
	JWTSecretMail     string
	SMTPUsername      string
	SMTPPassword      string
	SMTPHost          string
	APIGatewayHost    string
	ServiceName       string
	ServiceRoute      string
	ServiceUrl        string
	ServicePath       string
	APIKey            string
}

// Load từng file .env theo đúng mục đích
func init() {
	log.Println("Loading environment variables...")
	// Load từng file .env theo đúng mục đích
	loadEnvFile("env/.env")
	loadEnvFile("env/.env.database")
	loadEnvFile("env/.env.redis")
	loadEnvFile("env/.env.vstorage")

	// Database setting
	Config.DatabaseURL = os.Getenv("DATABASE_URL")
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
	Config.SMTPHost = os.Getenv("SMTP_HOST")
}

func loadEnvFile(filepath string) {
	err := godotenv.Load(filepath)
	if err != nil {
		log.Printf("Warning: Could not load %s file", filepath)
	}
}

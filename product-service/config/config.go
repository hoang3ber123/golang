package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var Config struct {
	// Base config
	SystemStatus     string
	BasePath         string
	ServiceName      string
	ServiceRoute     string
	ServiceURL       string
	ServicePath      string
	APIKey           string
	APIGatewayHost   string
	AllowHost        string
	HTTPPort         string
	GRPCAuthPort     string
	AuthServiceHost  string
	GRPCPort         string
	OrderServiceHost string
	GRPCOrderPort    string

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

	// Vstorage config
	VstorageProjectID                  string
	VstorageAuthURL                    string
	VstorageBaseURL                    string
	VstorageContainerName              string
	VstorageClientSecret               string
	VstorageClientID                   string
	VstorageSwiftUsername              string
	VstorageSwiftPassword              string
	VstorageURL                        string // url của domain thuê từ vstorage vd: hcm03.vstorage.vngcloud.vn
	VstorageAccount                    string
	VstorageDownloadContainer          string
	VstorageDownloadDirectory          string
	VstorageDownloadExpires            string
	VstorageDownloadContainerSecretKey string

	//Redis config
	RedisIndex    string
	RedisProtocol string
	RedisPassword string
	RedisHost     string
	RedisPort     string
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
	loadEnvFile(dir + fmt.Sprintf("/env/%s/.env", Config.SystemStatus))
	loadEnvFile(dir + fmt.Sprintf("/env/%s/.env.database", Config.SystemStatus))
	loadEnvFile(dir + fmt.Sprintf("/env/%s/.env.redis", Config.SystemStatus))
	loadEnvFile(dir + fmt.Sprintf("/env/%s/.env.vstorage", Config.SystemStatus))
	// System setting
	Config.BasePath = dir
	Config.AllowHost = os.Getenv("ALLOW_HOST")
	Config.HTTPPort = os.Getenv("HTTP_PORT")
	Config.GRPCPort = os.Getenv("GRPC_PORT")
	Config.GRPCAuthPort = os.Getenv("GRPC_AUTH_PORT")
	Config.AuthServiceHost = os.Getenv("AUTH_SERVICE_HOST")
	Config.GRPCOrderPort = os.Getenv("GRPC_ORDER_PORT")
	Config.OrderServiceHost = os.Getenv("ORDER_SERVICE_HOST")
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
	Config.SMTPHost = os.Getenv("SMTP_HOST")
	// Vstorage setting
	Config.VstorageAuthURL = os.Getenv("AUTH_URL")
	Config.VstorageBaseURL = os.Getenv("BASE_URL")
	Config.VstorageClientID = os.Getenv("CLIENT_ID")
	Config.VstorageClientSecret = os.Getenv("CLIENT_SECRECT")
	Config.VstorageContainerName = os.Getenv("CONTAINER_NAME")
	Config.VstorageProjectID = os.Getenv("PROJECT_ID")
	Config.VstorageSwiftPassword = os.Getenv("SWIFT_PASSWORD")
	Config.VstorageSwiftUsername = os.Getenv("SWIFT_USERNAME")
	Config.VstorageURL = os.Getenv("STORAGE_URL")
	Config.VstorageAccount = os.Getenv("ACCOUNT_VSTORAGE")
	Config.VstorageDownloadContainer = os.Getenv("DOWNLOAD_CONTAINER")
	Config.VstorageDownloadDirectory = os.Getenv("DOWNLOAD_DIRECTORY")
	Config.VstorageDownloadExpires = os.Getenv("DOWNLOAD_EXPIRES")
	Config.VstorageDownloadContainerSecretKey = os.Getenv("DOWNLOAD_CONTAINER_SECRET_KEY")

	// Redis setting
	Config.RedisIndex = os.Getenv("REDIS_INDEX")
	Config.RedisProtocol = os.Getenv("REDIS_PROTOCOL")
	Config.RedisPassword = os.Getenv("REDIS_PASSWORD")
	Config.RedisHost = os.Getenv("REDIS_HOST")
	Config.RedisPort = os.Getenv("REDIS_PORT")
}

func loadEnvFile(filepath string) {
	err := godotenv.Load(filepath)
	if err != nil {
		log.Printf("Warning: Could not load %s file", filepath)
	}
}

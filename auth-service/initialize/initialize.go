package initialize

import (
	"auth-service/config"
	"auth-service/internal/db"
	"auth-service/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type KongService struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type KongRoute struct {
	Paths     []string `json:"paths"`
	Name      string   `json:"name"`
	StripPath bool     `json:"strip_path,omitempty"`
}

type PluginConfig struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config,omitempty"`
}

type Consumer struct {
	Username string `json:"username"`
}

type KeyAuth struct {
	Key string `json:"key"`
}

func ConnectToApiGateway() {
	kongAdminURL := config.Config.APIGatewayHost
	client := &http.Client{Timeout: 10 * time.Second}

	fmt.Printf("Starting Kong configuration with Admin URL: %s\n", kongAdminURL)

	// Chờ Kong sẵn sàng
	for {
		fmt.Println("Checking Kong status...")
		resp, err := client.Get(fmt.Sprintf("%s/status", kongAdminURL))
		if err != nil {
			fmt.Printf("🔄 Waiting for Kong to be ready... Error: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()

		fmt.Printf("Received status response with code: %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Status response body: %s\n", string(body))

		var status map[string]interface{}
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&status); err != nil {
			fmt.Printf("Error decoding status response: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if database, ok := status["database"].(map[string]interface{}); ok {
			if reachable, ok := database["reachable"].(bool); ok && reachable {
				fmt.Println("Kong database is reachable!")
				break
			}
			fmt.Println("Database not reachable yet...")
		} else {
			fmt.Println("No database status found in response")
		}
		fmt.Println("🔄 Waiting for Kong to be ready...")
		time.Sleep(5 * time.Second)
	}

	fmt.Println("✅ Kong is ready! Applying configurations...")

	// 1. Tạo Service cho auth-service
	service := KongService{
		Name: config.Config.ServiceName,
		URL:  config.Config.ServiceUrl,
	}
	fmt.Printf("Creating service: %s with URL: %s\n", service.Name, service.URL)
	if err := postJSON(client, fmt.Sprintf("%s/services", kongAdminURL), service); err != nil {
		fmt.Printf("Error creating service: %v\n", err)
	} else {
		fmt.Println("Service created successfully")
	}

	// 2. Tạo Route cho Service
	route := KongRoute{
		Paths:     []string{config.Config.ServicePath},
		Name:      config.Config.ServiceRoute,
		StripPath: true,
	}
	fmt.Printf("Creating route: %s with path: %s\n", route.Name, route.Paths)
	if err := postJSON(client, fmt.Sprintf("%s/services/auth-services/routes", kongAdminURL), route); err != nil {
		fmt.Printf("Error creating route: %v\n", err)
	} else {
		fmt.Println("Route created successfully")
	}

	// 3. Thêm plugin CORS (Allow any host)
	corsPlugin := PluginConfig{
		Name: "cors",
		Config: map[string]interface{}{
			"origins":            []string{"http://localhost:3000", "http://127.0.0.1:3000"},                                         // Thêm cả 127.0.0.1
			"methods":            []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},                                                // Các method được phép
			"headers":            []string{"Accept", "Accept-Version", "Content-Length", "Content-Type", "Authorization", "api-key"}, // Headers được phép
			"exposed_headers":    []string{"X-Auth-token"},
			"credentials":        true,  // Không cho phép gửi credentials
			"max_age":            3600,  // Thời gian cache CORS (giây)
			"preflight_continue": false, // Không tiếp tục xử lý preflight
		},
	}
	fmt.Println("Enabling CORS plugin...")
	if err := postJSON(client, fmt.Sprintf("%s/plugins", kongAdminURL), corsPlugin); err != nil {
		fmt.Printf("Error enabling CORS plugin: %v\n", err)
	} else {
		fmt.Println("CORS plugin enabled successfully")
	}

	// 4. Bật rate-limiting cho toàn bộ Service
	rateLimit := PluginConfig{
		Name: "rate-limiting",
		Config: map[string]interface{}{
			"minute": 50,
			"policy": "local",
		},
	}
	fmt.Println("Enabling rate-limiting plugin...")
	if err := postJSON(client, fmt.Sprintf("%s/plugins", kongAdminURL), rateLimit); err != nil {
		fmt.Printf("Error enabling rate-limiting: %v\n", err)
	} else {
		fmt.Println("Rate-limiting enabled successfully")
	}

	// 5. Tạo Consumer admin
	consumer := Consumer{
		Username: "admin",
	}
	fmt.Println("Creating consumer: admin")
	if err := postJSON(client, fmt.Sprintf("%s/consumers", kongAdminURL), consumer); err != nil {
		fmt.Printf("Error creating consumer: %v\n", err)
	} else {
		fmt.Println("Consumer created successfully")
	}

	// 6. Gán API Key cho admin
	keyAuth := KeyAuth{
		Key: config.Config.APIKey,
	}
	fmt.Printf("Assigning API key: %s to consumer admin\n", keyAuth.Key)
	if err := postJSON(client, fmt.Sprintf("%s/consumers/admin/key-auth", kongAdminURL), keyAuth); err != nil {
		fmt.Printf("Error assigning API key: %v\n", err)
	} else {
		fmt.Println("API key assigned successfully")
	}

	// 7. Enable API Key authentication
	keyAuthPlugin := PluginConfig{
		Name: "key-auth",
		Config: map[string]interface{}{
			"key_names": []string{"api-key"},
		},
	}
	fmt.Println("Enabling key-auth plugin...")
	if err := postJSON(client, fmt.Sprintf("%s/plugins", kongAdminURL), keyAuthPlugin); err != nil {
		fmt.Printf("Error enabling key-auth: %v\n", err)
	} else {
		fmt.Println("Key-auth enabled successfully")
	}

	fmt.Println("✅ Kong setup completed!")
}

// Hàm helper để gửi POST request với JSON
func postJSON(client *http.Client, url string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return err
	}

	fmt.Printf("Sending POST request to %s with data: %s\n", url, string(jsonData))
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending POST request: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Received response with status code: %d\n", resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response body: %s\n", string(body))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func InitializingDatabase() {
	if db.DB == nil {
		log.Fatal("Database connection is nil. Please check db.InitDB()")
	}

	// Bật debug để xem truy vấn SQL
	db.DB = db.DB.Debug()

	// Initialize Role: Admin, Manager, Employee
	roles := []models.Role{
		{BaseSlugUnique: models.BaseSlugUnique{Title: "admin"}},
		{BaseSlugUnique: models.BaseSlugUnique{Title: "manager"}},
		{BaseSlugUnique: models.BaseSlugUnique{Title: "employee"}},
	}

	// Tạo từng role và đảm bảo lưu vào DB
	for i := range roles {
		if err := db.DB.Where("title = ?", roles[i].Title).FirstOrCreate(&roles[i]).Error; err != nil {
			log.Printf("Failed to create or find role %s: %v", roles[i].Title, err)
			return
		}
	}
	fmt.Println("Successfully created roles")

	// Lấy ID của role "admin"
	var adminRoleID *uuid.UUID
	for _, role := range roles {
		if role.Title == "admin" {
			id := role.ID
			adminRoleID = &id // Lấy ID từ role đã được tạo
			break
		}
	}
	if adminRoleID == nil {
		log.Fatal("Admin role ID is nil, something went wrong with role creation")
	}

	// Initial Employee
	dob, err := time.Parse("2006-01-02", "2003-09-16")
	if err != nil {
		log.Printf("Failed to parse DOB: %v", err)
		return
	}

	employees := []models.Employee{
		{
			Username:    "admin",
			Password:    "Admin@123",
			Name:        "Trần Thanh Hoàng",
			Email:       "Hoangila2016@gmail.com",
			Identity:    "000000000",
			Dob:         dob,
			Position:    "Backend",
			PhoneNumber: "090999999",
			Contact:     "https://facebook.com.vn",
			IsActive:    true,
			RoleID:      adminRoleID, // Gán RoleID thay vì Role
		},
	}

	// Tạo từng employee
	for i := range employees {
		if err := db.DB.Where("username = ?", employees[i].Username).FirstOrCreate(&employees[i]).Error; err != nil {
			log.Printf("Failed to create or find employee %s: %v", employees[i].Username, err)
			return
		}
	}
	fmt.Println("Successfully created employees")
}

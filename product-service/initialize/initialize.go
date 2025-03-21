package initialize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"product-service/config"
	"time"
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
		URL:  config.Config.ServiceURL,
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
	if err := postJSON(client, fmt.Sprintf("%s/services/%s/routes", kongAdminURL, service.Name), route); err != nil {
		fmt.Printf("Error creating route: %v\n", err)
	} else {
		fmt.Println("Route created successfully")
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

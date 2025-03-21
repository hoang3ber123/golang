package vstorage

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"product-service/config"
	"product-service/internal/db"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Define struct for auth request body
type AuthRequest struct {
	Auth struct {
		Identity struct {
			Methods  []string `json:"methods"`
			Password struct {
				User struct {
					Domain struct {
						Name string `json:"name"`
					} `json:"domain"`
					Name     string `json:"name"`
					Password string `json:"password"`
				} `json:"user"`
			} `json:"password"`
		} `json:"identity"`
		Scope struct {
			Project struct {
				Domain struct {
					Name string `json:"name"`
				} `json:"domain"`
				ID string `json:"id"`
			} `json:"project"`
		} `json:"scope"`
	} `json:"auth"`
}

// Authorize vstorage and save X-Auth-Token to Redis Cache for 1 hour
func AuthVstorage() (string, error) {
	fmt.Println("Xử lý ở AuthVstorage")
	// Get XAuthToken from redis
	fmt.Printf("Lấy redis")
	xAuthToken, err := db.RedisDB.Get(db.Ctx, "XAuthToken").Result()
	// If don't have key system will process an api to get an new XAuhtToken from vstorage and set it to redis
	if err == redis.Nil {
		fmt.Printf("Chưa lấy được token, tạo token mới")
		requestBody := AuthRequest{}
		requestBody.Auth.Identity.Methods = []string{"password"}
		requestBody.Auth.Identity.Password.User.Domain.Name = "default"
		requestBody.Auth.Identity.Password.User.Name = config.Config.VstorageSwiftUsername
		requestBody.Auth.Identity.Password.User.Password = config.Config.VstorageSwiftPassword
		requestBody.Auth.Scope.Project.Domain.Name = "default"
		requestBody.Auth.Scope.Project.ID = config.Config.VstorageProjectID

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return "", err
		}

		// setting request
		req, err := http.NewRequest("POST", config.Config.VstorageAuthURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}

		// call request
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
		}

		token := resp.Header.Get("X-Subject-Token")
		if token == "" {
			return "", fmt.Errorf("X-Subject-Token not found in response headers")
		}
		// Setting token for an hour
		err = db.RedisDB.Set(db.Ctx, "XAuthToken", token, time.Hour).Err()
		if err != nil {
			return "", fmt.Errorf("error when set key in Redis: %s", err)
		}

		return token, nil
		// If while getting XAuthToken error
	} else if err != nil {
		fmt.Printf("Lỗi khi lấy token")
		return "", fmt.Errorf("error when retrieving data from Redis: %s", err)
	}
	fmt.Printf("Đã lấy được token")
	return xAuthToken, nil
}

// PushFileToVoStorage uploads a file to VStorage
func PushFileToVStorage(XAuthToken string, file *multipart.FileHeader, directory string) (directoryPath string, uploadPath string, err error) {
	// Get binary content from file
	fileOpened, err := file.Open()
	if err != nil {
		return "", "", fmt.Errorf("cannot open file: %v", err)
	}
	defer fileOpened.Close()

	// Get filename and content type
	fileName := file.Filename
	contentType := file.Header.Get("Content-Type")

	// Create url to push file
	directoryPath = fmt.Sprintf("%s/%s", directory, fileName)
	uploadPath = fmt.Sprintf("%s/%s", config.Config.VstorageBaseURL, directoryPath)

	// Create PUT request
	req, err := http.NewRequest("PUT", uploadPath, fileOpened)
	if err != nil {
		return "", "", err
	}

	// Config header
	req.Header.Set("X-Auth-Token", XAuthToken)
	req.Header.Set("Content-Type", contentType)

	// Call request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode == http.StatusCreated { // 201 Created
		return directoryPath, uploadPath, nil
	} else {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}
}

// PushFileToVoStorage uploads a file to VStorage
func PushFileToDownloadVStorage(XAuthToken string, file *multipart.FileHeader, directory string) (directoryPath string, uploadPath string, err error) {
	// Get binary content from file
	fileOpened, err := file.Open()
	if err != nil {
		return "", "", fmt.Errorf("cannot open file: %v", err)
	}
	defer fileOpened.Close()

	// Get filename and content type
	fileName := file.Filename
	contentType := file.Header.Get("Content-Type")

	// Create url to push file
	vstorageUrl := config.Config.VstorageURL
	account := config.Config.VstorageAccount
	container := config.Config.VstorageDownloadContainer
	baseURL := fmt.Sprintf("%s/v1/%s/%s", vstorageUrl, account, container)
	directoryPath = fmt.Sprintf("%s/%s", directory, fileName)
	uploadPath = fmt.Sprintf("%s/%s", baseURL, directoryPath)

	// Create PUT request
	req, err := http.NewRequest("PUT", uploadPath, fileOpened)
	if err != nil {
		return "", "", err
	}

	// Config header
	req.Header.Set("X-Auth-Token", XAuthToken)
	req.Header.Set("Content-Type", contentType)

	// Call request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode == http.StatusCreated { // 201 Created
		return directoryPath, uploadPath, nil
	} else {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}
}

// bulk delete objects from vstorage
func BulkDeleteFileFromVstorage(XAuthToken string, files []string) error {
	// Build API URL
	apiURL := fmt.Sprintf("%s/?bulk-delete", config.Config.VstorageBaseURL)
	fmt.Println(apiURL)
	// Validate files path
	for index, val := range files {
		files[index] = fmt.Sprintf("%s/%s", config.Config.VstorageContainerName, val)
	}
	fmt.Println(files)
	// Convert list of files into newline-separated string
	bodyData := strings.Join(files, "\n")
	bodyReader := bytes.NewBufferString(bodyData)
	fmt.Println(bodyReader)
	// Create POST request
	req, err := http.NewRequest("POST", apiURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("X-Auth-Token", XAuthToken)
	req.Header.Set("Content-Type", "text/plain")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(respBody))
	}
	fmt.Println("Response:", string(respBody))
	return nil
}

func SetSecretKey(XAuthToken string) error {
	// Tạo URL API Metadata
	vstorageUrl := config.Config.VstorageURL
	account := config.Config.VstorageAccount
	container := config.Config.VstorageDownloadContainer
	secretKey := config.Config.VstorageDownloadContainerSecretKey
	// Tạo URL API
	apiURL := fmt.Sprintf("%s/v1/%s/%s", vstorageUrl, account, container)

	// Tạo request PUT để đặt Secret Key
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer([]byte("")))
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", XAuthToken)
	req.Header.Set("X-Container-Meta-Temp-URL-Key", secretKey)

	// Gửi request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Debug Response
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Body:", string(body))

	if resp.StatusCode == 401 {
		return fmt.Errorf("lỖI 401: Token hết hạn hoặc không có quyền truy cập")
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("lỗi đặt Secret Key, mã lỗi: %d", resp.StatusCode)
	}

	fmt.Println("✅ Đã đặt Secret Key thành công!")

	return nil
}

// GenerateTempURL tạo Temp URL cho object trong VStorage
func GenerateTempURL(file string) string {
	// Tạo URL API Metadata
	vstorageUrl := config.Config.VstorageURL
	account := config.Config.VstorageAccount
	container := config.Config.VstorageDownloadContainer
	secretKey := config.Config.VstorageDownloadContainerSecretKey
	expiresIn, _ := strconv.ParseInt(config.Config.VstorageDownloadExpires, 10, 64)
	// Thời gian hết hạn (Unix timestamp)
	expires := time.Now().Unix() + expiresIn

	// Tạo đường dẫn object
	path := fmt.Sprintf("/v1/%s/%s/%s", account, container, file)

	// Tạo chuỗi chữ ký
	signingString := fmt.Sprintf("GET\n%d\n%s", expires, path)

	// Tạo HMAC-SHA1 chữ ký
	h := hmac.New(sha1.New, []byte(secretKey))
	h.Write([]byte(signingString))
	signature := hex.EncodeToString(h.Sum(nil))

	// Encode URL
	tempURL := fmt.Sprintf("%s%s?temp_url_sig=%s&temp_url_expires=%d",
		vstorageUrl, path, url.QueryEscape(signature), expires)

	return tempURL
}

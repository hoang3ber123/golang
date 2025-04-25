package services

import (
	"context"
	"encoding/json"
	"fmt"
	"product-service/config"
	"product-service/internal/db"
	"product-service/internal/models"
	"product-service/internal/responses"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// BuildHierarchy xây dựng cây phân cấp từ danh sách categories
func BuildHierarchy(categories []models.Category, title string, parentID *uuid.UUID) ([]models.Category, error) {
	// Map để tra cứu nhanh category theo ID
	categoryMap := make(map[uuid.UUID]*models.Category)
	for i := range categories {
		categoryMap[categories[i].ID] = &categories[i]
	}

	// Lọc category gốc dựa trên title và parentID
	var root models.Category
	found := false
	for _, cat := range categories {
		if (title == "" || cat.Title == title) &&
			(parentID == nil || (cat.ParentID != nil && *cat.ParentID == *parentID)) {
			root = cat // Gán giá trị, không phải con trỏ
			found = true
			break
		}
	}

	// Nếu không tìm thấy root thỏa mãn điều kiện
	if !found {
		return nil, fmt.Errorf("no matching root category found")
	}

	// Xây dựng cây từ root
	stack := []models.Category{root}
	visited := make(map[uuid.UUID]bool) // Để tránh vòng lặp vô hạn nếu có dữ liệu không hợp lệ
	for len(stack) > 0 {
		// Pop phần tử đầu tiên từ stack
		current := stack[0]
		stack = stack[1:]

		if visited[current.ID] {
			continue
		}
		visited[current.ID] = true

		// Tìm các category con
		for i := range categories {
			cat := categories[i] // Lấy giá trị, không phải con trỏ
			if cat.ParentID != nil && *cat.ParentID == current.ID {
				current.Children = append(current.Children, cat) // Thêm giá trị Category vào []Category
				stack = append(stack, cat)
			}
		}

		// Cập nhật lại categoryMap để phản ánh Children
		categoryMap[current.ID].Children = current.Children
	}

	return []models.Category{root}, nil
}

// GetCategories từ redis
func GetCategoriesFromRedis() (string, *responses.ErrorResponse) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	redisKey := config.Config.RedisCategoriesKey
	categoriesJSON, err := db.RedisDB.Get(ctx, redisKey).Result()
	var allCategories []models.Category

	// Bước 1: Nếu Redis có lỗi khác nil, trả lỗi
	if err != nil && err != redis.Nil {
		return "", responses.NewErrorResponse(fiber.StatusInternalServerError, "Redis error: "+err.Error())
	}

	// Bước 2: Nếu không lỗi nhưng không có dữ liệu, truy vấn từ DB và cache lại
	if categoriesJSON == "" {
		if err := db.DB.Find(&allCategories).Error; err != nil {
			return "", responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+err.Error())
		}
		jsonData, err := json.Marshal(allCategories)
		if err != nil {
			return "", responses.NewErrorResponse(fiber.StatusInternalServerError, "JSON marshal error: "+err.Error())
		}

		// Cache dữ liệu lại Redis (set thời gian 1 giờ hoặc tùy)
		if err := db.RedisDB.Set(db.Ctx, redisKey, jsonData, time.Hour*24).Err(); err != nil {
			return "", responses.NewErrorResponse(fiber.StatusInternalServerError, "Redis set error: "+err.Error())
		}

		// Gán lại categoriesJSON với dữ liệu vừa cache
		categoriesJSON = string(jsonData)
	}
	return categoriesJSON, nil
}

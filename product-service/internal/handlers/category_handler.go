package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"product-service/config"
	"product-service/internal/db"
	grpcclient "product-service/internal/grpc_client"
	"product-service/internal/models"
	"product-service/internal/responses"
	"product-service/internal/serializers"
	"product-service/internal/services"
	"product-service/pagination"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CategoryCreate(c *fiber.Ctx) error {
	// Xử lý tạo category nếu xác thực thành công
	serializer := new(serializers.CategoryCreateSerializer)
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}
	// Serializer to model
	Category := serializer.ToModel()
	if err := db.DB.Create(&Category).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to create category: "+err.Error()).Send(c)
	}
	// xóa categories trong cache khi tạo
	db.RedisDB.Del(db.Ctx, config.Config.RedisCategoriesKey)
	// Response
	return responses.NewSuccessResponse(fiber.StatusCreated, serializers.CategoryDetailResponse(Category)).Send(c)
}

func CategoryUpdate(c *fiber.Ctx) error {
	id := c.Params("id")
	var instance models.Category

	// Kiểm tra xem category có tồn tại không
	if err := db.DB.First(&instance, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return responses.NewErrorResponse(fiber.StatusNotFound, "Category not found").Send(c)
		}
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+err.Error()).Send(c)
	}

	serializer := new(serializers.CategoryUpdateSerializer)

	// Nếu có lỗi validation, return ngay lập tức
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}

	// Nếu validation OK, thực hiện update
	if err := serializer.Update(&instance); err != nil {
		return err.Send(c)
	}
	// xóa categories trong cache khi xóa
	db.RedisDB.Del(db.Ctx, config.Config.RedisCategoriesKey)
	// Trả về response thành công
	return responses.NewSuccessResponse(fiber.StatusOK, serializers.CategoryDetailResponse(&instance)).Send(c)
}

func CategoryList(c *fiber.Ctx) error {
	// gọi hàm lấy danh sách categories trong redis
	categoriesJSON, err := services.GetCategoriesFromRedis()
	if err != nil {
		err.Send(c)
	}

	// Nếu Redis có dữ liệu, parse lại
	var allCategories []models.Category
	fmt.Println("json data:", categoriesJSON)
	if err := json.Unmarshal([]byte(categoriesJSON), &allCategories); err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "JSON unmarshal error: "+err.Error()).Send(c)
	}
	// Lọc dữ liệu theo query string
	title := c.Query("title")
	parentID := c.Query("parent_id")

	filtered := []models.Category{}
	for _, cat := range allCategories {
		if title != "" && !strings.Contains(strings.ToLower(cat.Title), strings.ToLower(title)) {
			continue
		}
		if parentID != "" {
			if cat.ParentID == nil || cat.ParentID.String() != parentID {
				continue
			}
		} else {
			if cat.ParentID != nil {
				continue
			}
		}
		filtered = append(filtered, cat)
	}

	// Pagination
	pageData, paginationInfo, errResp := pagination.PaginationWithSlice(c, filtered)
	if errResp != nil {
		return errResp.Send(c)
	}
	// Serialize kết quả
	result := serializers.CategoryListResponse(&pageData)

	return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
		"pagination": paginationInfo,
		"result":     result,
	}).Send(c)
}

type CategoryTree struct {
	ID       uuid.UUID       `json:"id"`
	Title    string          `json:"title"`
	Slug     string          `json:"slug"`
	Children []*CategoryTree `json:"children"`
}

func CategoryAllList(c *fiber.Ctx) error {

	var categories []struct {
		ID       uuid.UUID  `json:"id"`
		Title    string     `json:"title"`
		Slug     string     `json:"slug"`
		ParentID *uuid.UUID `json:"parent_id"` // Dùng *uuid.UUID vì giá trị có thể NULL
	}

	// Truy vấn tất cả category
	query := db.DB.Table("categories").Select("id, title, slug, parent_id")
	query.Find(&categories)

	// Xây dựng map để nhóm các category theo parent_id
	categoryMap := make(map[uuid.UUID]*CategoryTree)
	var rootCategories []*CategoryTree

	// Tạo các node category
	for _, cat := range categories {
		categoryMap[cat.ID] = &CategoryTree{
			ID:       cat.ID,
			Title:    cat.Title,
			Slug:     cat.Slug,
			Children: []*CategoryTree{},
		}
	}

	// Ghép các category vào cây
	for _, cat := range categories {
		if cat.ParentID != nil {
			parent, exists := categoryMap[*cat.ParentID]
			if exists {
				parent.Children = append(parent.Children, categoryMap[cat.ID])
			}
		} else {
			// Nếu không có ParentID, đây là root category
			rootCategories = append(rootCategories, categoryMap[cat.ID])
		}
	}

	return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
		"result": rootCategories,
	}).Send(c)
}

func CategoryDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var instance models.Category
	if err := db.DB.First(&instance, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return responses.NewErrorResponse(fiber.StatusNotFound, "Category not found").Send(c)
		}
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+err.Error()).Send(c)
	}
	return responses.NewSuccessResponse(fiber.StatusOK, serializers.CategoryDetailResponse(&instance)).Send(c)
}

func CategoryDelete(c *fiber.Ctx) error {
	serializer := new(serializers.CategoryDeleteSerializer)
	//  Nếu có lỗi validation, return ngay lập tức
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}

	//  Nếu validation OK, thực hiện delete
	if err := serializer.Delete(); err != nil {
		return err.Send(c)
	}
	// xóa categories trong cache khi xóa
	db.RedisDB.Del(db.Ctx, config.Config.RedisCategoriesKey)
	//  Trả về response thành công
	return responses.NewSuccessResponse(fiber.StatusOK, "Delete successfully").Send(c)
}

// Gợi ý category
func CategoryRecommend(c *fiber.Ctx) error {
	// Lấy query từ body
	serializer := new(serializers.CategoryRecommendSerializer)
	//  Nếu có lỗi validation, return ngay lập tức
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}
	// lấy query từ body vd: {"query": "i want ...."}
	// Gọi gRPC để lấy danh sách category IDs dựa trên query
	category_ids, err := grpcclient.GetRecommendCategoryIDs(serializer.Query)
	if err != nil {
		// Nếu có lỗi từ gRPC, trả về lỗi với chi tiết lỗi
		return err.Send(c)
	}

	var categories []models.Category
	if err := db.DB.Model(models.Category{}).Where("id IN ?", category_ids).Find(&categories).Error; err != nil {
		// Nếu có lỗi khi truy vấn cơ sở dữ liệu
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error fetching categories from database").Send(c)
	}

	result := serializers.CategoryListResponse(&categories)

	//  Trả về response thành công
	return responses.NewSuccessResponse(fiber.StatusOK, result).Send(c)
}

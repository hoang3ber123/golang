package handlers

import (
	"errors"
	"fmt"
	"product-service/internal/db"
	"product-service/internal/models"
	"product-service/internal/responses"
	"product-service/internal/serializers"
	"product-service/internal/services"
	"product-service/pagination"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Create product
func ProductCreate(c *fiber.Ctx) error {
	// Xử lý tạo category nếu xác thực thành công
	serializer := new(serializers.ProductCreateSerializer)
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}

	// Serializer to model
	product, errCreate := serializer.Create()
	if errCreate != nil {
		return errCreate.Send(c)
	}

	// Starting processing file
	relatedType := product.GetRelatedType()
	directory := product.GetDirectory()
	// get form-data body
	form, err := c.MultipartForm()
	if err != nil {
		return responses.NewSuccessResponse(fiber.StatusBadRequest, "Cannot parse form").Send(c)
	}
	files := form.File["files"]
	if len(files) != 0 {
		if err := services.BulkCreateMedia(files, product.ID, relatedType, directory); err != nil {
			return err.Send(c)
		}
	}

	// Query to get product and role
	var categories []models.Category
	if err := db.DB.Model(&models.Category{}).
		Joins("JOIN product_categories ON product_categories.category_id = categories.id").
		Where("product_categories.product_id = ?", product.ID).
		Find(&categories).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, err.Error()).Send(c)
	}

	product.Categories = categories
	return responses.NewSuccessResponse(fiber.StatusCreated, serializers.ProductDetailResponse(product)).Send(c)
}

// product list
func ProductList(c *fiber.Ctx) error {
	// Prepare search scope (optional)
	titleQueries := c.Query("title")
	priceMin := c.Query("price_min")          // Giá tối thiểu
	priceMax := c.Query("price_max")          // Giá tối đa
	createdAtMin := c.Query("created_at_min") // Format: "YYYY-MM-DD"
	createdAtMax := c.Query("created_at_max") // Format: "YYYY-MM-DD"
	categoryIDsQueries := c.Query("category_ids")
	priceOrderQuery := c.Query("price_order")
	createdAtOrderQuery := c.Query("created_at_order")
	// Initializer query
	query := db.DB.Model(&models.Product{})

	// Tìm kiếm theo query
	if titleQueries != "" {
		query = query.Where("products.title LIKE ?", "%"+titleQueries+"%")
	}

	// Tìm kiếm theo price (min: >=, max: <=, cả hai: BETWEEN)
	if priceMin != "" || priceMax != "" {
		if priceMin != "" && priceMax != "" {
			// Có cả min và max -> BETWEEN
			min, errMin := strconv.ParseFloat(priceMin, 64)
			max, errMax := strconv.ParseFloat(priceMax, 64)
			if errMin != nil || errMax != nil {
				return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid price_min or price_max format").Send(c)
			}
			if min > max {
				return responses.NewErrorResponse(fiber.StatusBadRequest, "price_min must be less than or equal to price_max").Send(c)
			}
			query = query.Where("price BETWEEN ? AND ?", min, max)
		} else if priceMin != "" {
			// Chỉ có min -> >=
			min, err := strconv.ParseFloat(priceMin, 64)
			if err != nil {
				return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid price_min format").Send(c)
			}
			query = query.Where("price >= ?", min)
		} else if priceMax != "" {
			// Chỉ có max -> <=
			max, err := strconv.ParseFloat(priceMax, 64)
			if err != nil {
				return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid price_max format").Send(c)
			}
			query = query.Where("price <= ?", max)
		}
	}

	// Tìm kiếm theo created_at
	if createdAtMin != "" || createdAtMax != "" {
		if createdAtMin != "" && createdAtMax != "" {
			// Parse min và max thành time.Time
			min, errMin := time.Parse("2006-01-02", createdAtMin)
			max, errMax := time.Parse("2006-01-02", createdAtMax)
			if errMin != nil || errMax != nil {
				return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid created_at_min or created_at_max format (use YYYY-MM-DD)").Send(c)
			}
			if min.After(max) {
				return responses.NewErrorResponse(fiber.StatusBadRequest, "created_at_min must be earlier than or equal to created_at_max").Send(c)
			}
			query = query.Where("created_at BETWEEN ? AND ?", min, max)
		} else if createdAtMin != "" {
			min, err := time.Parse("2006-01-02", createdAtMin)
			if err != nil {
				return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid created_at_min format (use YYYY-MM-DD)").Send(c)
			}
			query = query.Where("created_at >= ?", min)
		} else if createdAtMax != "" {
			max, err := time.Parse("2006-01-02", createdAtMax)
			if err != nil {
				return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid created_at_max format (use YYYY-MM-DD)").Send(c)
			}
			query = query.Where("created_at <= ?", max)
		}
	}

	// Tìm kiếm theo category_ids với Joins
	if categoryIDsQueries != "" {
		categoryIDs := strings.Split(categoryIDsQueries, ",")
		if len(categoryIDs) > 0 {
			fmt.Println("validIDs:", categoryIDs) // Debug
			query = query.Joins("LEFT JOIN product_categories ON product_categories.product_id = products.id").
				Where("product_categories.category_id IN ?", categoryIDs) // Gán lại query
		}
	}

	// Sắp xếp theo price (nếu có)
	if priceOrderQuery != "" {
		order := strings.ToLower(priceOrderQuery)
		if order == "asc" || order == "desc" {
			query = query.Order("price " + order)
		} else {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid price_order, use 'asc' or 'desc'").Send(c)
		}
	}

	// Sắp xếp theo price (nếu có)
	if createdAtOrderQuery != "" {
		order := strings.ToLower(createdAtOrderQuery)
		if order == "asc" || order == "desc" {
			query = query.Order("products.created_at " + order)
		} else {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid created_at_order, use 'asc' or 'desc'").Send(c)
		}
	}

	// Sử dụng hàm phân trang
	var products []models.Product
	paginator, err := pagination.PaginateWithGORM(c, query, &products)
	if err != nil {
		return err.Send(c)
	}

	return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
		"pagination": paginator,
		"result":     serializers.ProductListResponse(&products),
	}).Send(c)
}

func ProductUpdate(c *fiber.Ctx) error {
	id := c.Params("id")
	var instance models.Product

	// Kiểm tra xem Product có tồn tại không
	if err := db.DB.First(&instance, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return responses.NewErrorResponse(fiber.StatusNotFound, "Product not found").Send(c)
		}
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+err.Error()).Send(c)
	}

	serializer := new(serializers.ProductUpdateSerializer)

	// Nếu có lỗi validation, return ngay lập tức
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}

	// Nếu validation OK, thực hiện update
	if err := serializer.Update(&instance); err != nil {
		return err.Send(c)
	}

	// Starting processing file
	relatedType := instance.GetRelatedType()
	directory := instance.GetDirectory()
	// get form-data body
	form, err := c.MultipartForm()
	if err != nil {
		return responses.NewSuccessResponse(fiber.StatusBadRequest, "Cannot parse form").Send(c)
	}
	files := form.File["files"]
	if len(files) != 0 {
		if err := services.BulkCreateMedia(files, instance.ID, relatedType, directory); err != nil {
			return err.Send(c)
		}
	}

	// Query to get product and role
	var categories []models.Category
	if err := db.DB.Model(&models.Category{}).
		Joins("JOIN product_categories ON product_categories.category_id = categories.id").
		Where("product_categories.product_id = ?", instance.ID).
		Find(&categories).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, err.Error()).Send(c)
	}

	instance.Categories = categories
	return responses.NewSuccessResponse(fiber.StatusCreated, serializers.ProductDetailResponse(&instance)).Send(c)
}

func ProductDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var instance models.Product
	if err := db.DB.First(&instance, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return responses.NewErrorResponse(fiber.StatusNotFound, "Product not found").Send(c)
		}
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+err.Error()).Send(c)
	}
	// Query to get product and role
	var categories []models.Category
	if err := db.DB.Model(&models.Category{}).
		Joins("JOIN product_categories ON product_categories.category_id = categories.id").
		Where("product_categories.product_id = ?", instance.ID).
		Find(&categories).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, err.Error()).Send(c)
	}
	instance.Categories = categories
	return responses.NewSuccessResponse(fiber.StatusOK, serializers.ProductDetailResponse(&instance)).Send(c)
}

func ProductDelete(c *fiber.Ctx) error {
	serializer := new(serializers.ProductDeleteSerializer)
	//  Nếu có lỗi validation, return ngay lập tức
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}

	//  Nếu validation OK, thực hiện delete
	if err := serializer.Delete(); err != nil {
		return err.Send(c)
	}

	//  Trả về response thành công
	return responses.NewSuccessResponse(fiber.StatusOK, "Delete successfully").Send(c)
}

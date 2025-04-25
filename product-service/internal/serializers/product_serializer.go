package serializers

import (
	"fmt"
	"path/filepath"
	"product-service/config"
	"product-service/internal/db"
	"product-service/internal/models"
	"product-service/internal/responses"
	"product-service/internal/services"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type ProductCreateSerializer struct {
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description" validate:"required"`
	Link        *string   `json:"link"`
	Price       *float64  `json:"price"`
	Categories  []string  `json:"categories" validate:"required,dive,uuid_rfc4122"`
	UserID      uuid.UUID `json:"user_id"`
}

func (s *ProductCreateSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}

	// Lấy thông tin user từ context
	user, ok := c.Locals("employee").(*models.Employee) // Giả sử user đã được middleware xác thực và lưu vào context
	if !ok {
		return responses.NewErrorResponse(fiber.StatusUnauthorized, "Unauthorized")
	}
	s.UserID = user.ID // Gán UserID từ context vào struct

	// Basic validation với go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	// Custom validation: Kiểm tra danh sách category có empty không
	if len(s.Categories) == 0 {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Categories list cannot be empty")
	}
	var count int
	if err := db.DB.Raw("SELECT COUNT(*) FROM categories WHERE id IN (?)", s.Categories).Scan(&count).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error checking categories: "+err.Error())
	}
	if count != len(s.Categories) {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Some categories do not exist")
	}

	// Kiểm tra file
	form, err := c.MultipartForm()
	if err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Cannot parse form")
	}
	files := form.File["files"]
	// Kiểm tra chỉ có đúng một file có phần mở rộng .zip hoặc .rar
	archiveCount := 0
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext == ".zip" || ext == ".rar" {
			archiveCount++
		}
	}
	if archiveCount == 0 {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "At least one file must have a .zip or .rar extension")
	} else if archiveCount > 1 {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Only one .zip or .rar file is allowed")
	}

	// Nếu không có lỗi, trả về nil
	return nil
}

// ToModel converts the serializer to a model
func (s *ProductCreateSerializer) Create() (*models.Product, *responses.ErrorResponse) {
	product := models.Product{
		BaseSlug: models.BaseSlug{
			Title: s.Title,
		},
		Description: s.Description,
		Link:        s.Link,
		Price:       s.Price,
		UserID:      s.UserID,
	}

	// Insert product
	if err := db.DB.Create(&product).Error; err != nil {
		return nil, responses.NewErrorResponse(fiber.StatusInternalServerError, "Error when creating product")
	}

	// Bulk insert product_categories nếu có category
	lenCategories := len(s.Categories)
	if lenCategories > 0 {
		productCategories := make([]models.ProductCategory, lenCategories)
		for i, categoryID := range s.Categories {
			parsedID, _ := uuid.Parse(categoryID)
			productCategories[i] = models.ProductCategory{
				ProductID:  product.ID,
				CategoryID: parsedID,
			}
		}
		// Batch insert
		db.DB.Create(&productCategories)
	}

	return &product, nil
}

type ProductDetailResponseSerializer struct {
	BaseResponseSerializer
	Title       string                           `json:"title"`
	Slug        string                           `json:"slug"`
	Description string                           `json:"description"`
	Link        string                           `json:"link"`
	Price       float64                          `json:"price"`
	Categories  []CategoryListResponseSerializer `json:"categories"`
	Medias      []MediaListResponseSerializer    `json:"medias"`
}

func ProductDetailResponse(instance *models.Product) *ProductDetailResponseSerializer {
	// Xử lý danh sách category
	categories := CategoryListResponse(&instance.Categories)
	// Xử lý danh sách media
	var mediaList []models.Media
	db.DB.Where("related_id = ? AND related_type = ? AND file_type ='image' AND status = 'using'", instance.ID, instance.GetRelatedType()).Find(&mediaList)
	medias := MediaListResponse(&mediaList)
	// Xử lý trường nullable (Link, Price)
	var link string
	if instance.Link != nil {
		link = *instance.Link
	}

	var price float64
	if instance.Price != nil {
		price = *instance.Price
	}

	return &ProductDetailResponseSerializer{
		BaseResponseSerializer: BaseResponseSerializer{
			ID:        instance.ID,
			CreatedAt: instance.CreatedAt,
			UpdatedAt: instance.UpdatedAt,
		},
		Slug:        instance.Slug,
		Title:       instance.Title,
		Description: instance.Description,
		Link:        link,
		Price:       price,
		Categories:  categories,
		Medias:      medias,
	}
}

func ProductDetailEmployeeResponse(instance *models.Product) *ProductDetailResponseSerializer {
	// Xử lý danh sách category
	categories := CategoryListResponse(&instance.Categories)
	// Xử lý danh sách media
	var mediaList []models.Media
	db.DB.Where("related_id = ? AND related_type = ? AND status = 'using'", instance.ID, instance.GetRelatedType()).Find(&mediaList)
	medias := MediaListResponse(&mediaList)
	// Xử lý trường nullable (Link, Price)
	var link string
	if instance.Link != nil {
		link = *instance.Link
	}

	var price float64
	if instance.Price != nil {
		price = *instance.Price
	}

	return &ProductDetailResponseSerializer{
		BaseResponseSerializer: BaseResponseSerializer{
			ID:        instance.ID,
			CreatedAt: instance.CreatedAt,
			UpdatedAt: instance.UpdatedAt,
		},
		Slug:        instance.Slug,
		Title:       instance.Title,
		Description: instance.Description,
		Link:        link,
		Price:       price,
		Categories:  categories,
		Medias:      medias,
	}
}

// ProductListResponseSerializer struct để serialize danh sách Product
type ProductListResponseSerializer struct {
	BaseSlugResponseSerializer
	Description string                           `json:"description"`
	Link        string                           `json:"link"`
	Price       float64                          `json:"price"`
	Categories  []CategoryListResponseSerializer `json:"categories"`
	Medias      []MediaListResponseSerializer    `json:"medias"`
}

// ProductListResponse serialize danh sáchProduct thành slice ProductListResponseSerializer
func ProductListResponse(instance *[]models.Product) []ProductListResponseSerializer {
	lenProdList := len(*instance)
	results := make([]ProductListResponseSerializer, lenProdList) // Preallocate slice

	// Tạo mảng id product
	productIDs := make([]uuid.UUID, lenProdList)
	for prod_index, prod := range *instance {
		productIDs[prod_index] = prod.ID
	}

	// Hash table để chứa dữ liệu
	categoryMap := make(map[uuid.UUID][]CategoryListResponseSerializer)
	mediaMap := make(map[uuid.UUID][]MediaListResponseSerializer)

	// Xử lý lọc danh sách category thành một hashtable
	query := `
	WITH ranked_categories AS (
		SELECT 
			c.id, c.title, c.slug, pc.product_id,
			ROW_NUMBER() OVER (PARTITION BY pc.product_id ORDER BY c.id) AS row_num
		FROM categories c
		JOIN product_categories pc ON pc.category_id = c.id
		WHERE pc.product_id IN (?)
	)
	SELECT id, title, slug, product_id FROM ranked_categories WHERE row_num <= 3;
	`
	// truy vấn bỏ vào danh sách category
	var categoryResults []struct {
		ID        uuid.UUID
		Title     string
		Slug      string
		ProductID uuid.UUID
	}
	db.DB.Raw(query, productIDs).Scan(&categoryResults)
	// duyệt một vòng lặp tạo hash table
	// 🔥 Lưu category vào hash table
	for _, cat := range categoryResults {
		categoryMap[cat.ProductID] = append(categoryMap[cat.ProductID], CategoryListResponseSerializer{
			ID:    cat.ID,
			Title: cat.Title,
			Slug:  cat.Slug,
		})
	}

	// 🔥 Truy vấn media (Lấy tối đa 3 media cho mỗi product)
	relatedType := "products" // bảng mà media chứa ảnh
	status := "using"         // media đang sử dụng
	maxRowNum := 3            // Số lượng media tối đa cho mỗi product
	file_type := "image"      // kiểu file
	queryMedia := fmt.Sprintf(`
WITH ranked_media AS (
    SELECT 
        m.id, m.file, m.file_type, m.related_id AS product_id,
        ROW_NUMBER() OVER (PARTITION BY m.related_id ORDER BY m.id) AS row_num
    FROM media m
    WHERE m.related_id IN (?) AND m.related_type = '%s' AND m.status = '%s' AND m.file_type = '%s'
)
SELECT id, file, file_type, product_id FROM ranked_media WHERE row_num <= %d;
`, relatedType, status, file_type, maxRowNum)

	var mediaResults []struct {
		ID        uint
		FileType  string
		File      string
		ProductID uuid.UUID
	}
	db.DB.Raw(queryMedia, productIDs).Scan(&mediaResults)

	// Lưu media vào hash table
	baseURL := config.Config.VstorageBaseURL
	for _, media := range mediaResults {
		mediaMap[media.ProductID] = append(mediaMap[media.ProductID], MediaListResponseSerializer{
			ID:       media.ID,
			FileType: media.FileType,
			File:     fmt.Sprintf("%s/%s", baseURL, media.File),
		})
	}

	for prod_index, product := range *instance {
		var link string
		if product.Link != nil {
			link = *product.Link
		}

		var price float64
		if product.Price != nil {
			price = *product.Price
		}
		results[prod_index] = ProductListResponseSerializer{
			BaseSlugResponseSerializer: BaseSlugResponseSerializer{
				BaseResponseSerializer: BaseResponseSerializer{
					ID:        product.ID,
					CreatedAt: product.CreatedAt,
				},
				Title: product.Title,
				Slug:  product.Slug,
			},
			Description: product.Description,
			Price:       price,
			Link:        link,
			Categories:  categoryMap[product.ID], // Ghép category
			Medias:      mediaMap[product.ID],    // Ghép media
		}
	}

	return results
}

type ProductDeleteSerializer struct {
	IDs []string `json:"ids" validate:"required,dive,uuid_rfc4122"`
}

func (s *ProductDeleteSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}
	// Basic validation with go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	//
	return nil
}

func (s *ProductDeleteSerializer) Delete() *responses.ErrorResponse {
	// Thực hiện xóa các category có ID trong s.IDs
	result := db.DB.Where("id IN (?)", s.IDs).Delete(&models.Product{})
	if result.Error != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to delete product: "+result.Error.Error())
	}

	// Kiểm tra nếu không có bản ghi nào bị xóa
	if result.RowsAffected == 0 {
		return responses.NewErrorResponse(fiber.StatusNotFound, "No matching categories found to delete")
	}
	// Xóa media của product
	relatedType := "product"
	result = db.DB.Model(&models.Media{}).
		Where("related_id IN ? AND related_type = ?", s.IDs, relatedType).
		Update("status", models.MediaStatusDeleteCascade)

	// Kiểm tra lỗi
	if result.Error != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to update media status: "+result.Error.Error())
	}

	// Trả về nil nếu thành công
	return nil
}

type ProductUpdateSerializer struct {
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Link             *string  `json:"link"`
	Price            *float64 `json:"price"`
	Categories       []string `json:"categories" validate:"dive,uuid_rfc4122"`
	CategoriesRemove []string `json:"categoriesremove" validate:"dive,uuid_rfc4122"`
	FilesRemove      []uint   `json:"filesremove" validate:"dive,number"`
}

func (s *ProductUpdateSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}
	// Basic validation với go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	// lấy dữ liệu trước để kiểm tra dữ liệu trước khi update
	// Kiểm tra sự tồn tại của product
	productID := c.Params("id")
	if productID == "" {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Product ID is required in path (/:id)")
	}
	var productExists bool
	if err := db.DB.Raw("SELECT EXISTS (SELECT 1 FROM products WHERE id = ?)", productID).Scan(&productExists).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error checking product ID: "+err.Error())
	}
	if !productExists {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Product ID does not exist: "+productID)
	}
	// Custom validation: Kiểm tra danh sách category và dạnh sách media remove
	if len(s.Categories) > 0 || len(s.CategoriesRemove) > 0 {
		var categoryIDs []string
		err := db.DB.Model(&models.Product{}).
			Joins("LEFT JOIN product_categories ON product_categories.product_id = products.id").
			Where("products.id = ?", productID).
			Select("product_categories.category_id").
			Find(&categoryIDs).Error
		if err != nil {
			return responses.NewErrorResponse(fiber.StatusInternalServerError, "Find category failed: "+err.Error())
		}

		// Tạo map để kiểm tra nhanh hơn
		categoryMap := make(map[string]bool)
		for _, id := range categoryIDs {
			categoryMap[id] = true
		}

		// Kiểm tra nếu s.Categories có category đã tồn tại
		var existingCategories []string
		for _, cat := range s.Categories {
			if categoryMap[cat] {
				existingCategories = append(existingCategories, cat)
			}
		}
		if len(existingCategories) > 0 {
			return responses.NewErrorResponse(fiber.StatusBadRequest, fmt.Sprintf("Categories %v already exist in product", existingCategories))
		}

		// Kiểm tra nếu s.CategoriesRemove có category không tồn tại
		var notFoundCategories []string
		for _, cat := range s.CategoriesRemove {
			if !categoryMap[cat] {
				notFoundCategories = append(notFoundCategories, cat)
			}
		}
		if len(notFoundCategories) > 0 {
			return responses.NewErrorResponse(fiber.StatusBadRequest, fmt.Sprintf("Some categories %v do not exist in product, cannot remove", notFoundCategories))
		}

		// Kiểm tra nếu tổng số category sau khi cập nhật < 1
		newCategoryCount := len(categoryIDs) + len(s.Categories) - len(s.CategoriesRemove)
		if newCategoryCount < 1 {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "A product must have at least one category")
		}
	}

	// Custom validation: Kiểm tra danh sách media remove

	form, err := c.MultipartForm()
	if err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Cannot parse form")
	}
	files := form.File["files"]

	// Kiểm tra chỉ có ít hơn 1 file có phần mở rộng .zip hoặc .rar
	archiveCount := 0
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext == ".zip" || ext == ".rar" {
			archiveCount++
		}
	}

	if archiveCount > 1 {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Only one .zip or .rar file is allowed")
	}

	// kiểm tra xem thử là có một file_type = 'download_file' bị xóa
	var removeDownloadFileExists bool
	if err := db.DB.Raw("SELECT EXISTS (SELECT 1 FROM media WHERE id IN ? AND related_id = ? AND related_type = 'products' AND status = ? AND file_type = 'download_file')",
		s.FilesRemove, productID, models.MediaStatusUsing).Scan(&removeDownloadFileExists).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error checking product ID: "+err.Error())
	}

	// Nếu có file .zip hoặc .rar được upload, thì bắt buộc phải có một file download_file bị xóa
	// Ngược lại, nếu có một file download_file bị xóa, thì bắt buộc phải có một file .zip hoặc .rar được upload
	if (archiveCount == 1 && !removeDownloadFileExists) || (archiveCount == 0 && removeDownloadFileExists) {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Product must have at least 1 download_file")
	}

	if len(s.FilesRemove) > 0 {
		var mediaIDs []uint // Chứa danh sách ID của media đã liên kết với sản phẩm

		// Truy vấn để lấy danh sách media ID status = 'using' liên kết với sản phẩm
		err := db.DB.Model(&models.Media{}).
			Where("related_id = ? AND status = ? AND related_type = 'products'", productID, models.MediaStatusUsing).
			Select("id").
			Find(&mediaIDs).Error

		if err != nil {
			return responses.NewErrorResponse(fiber.StatusInternalServerError, "Find media failed: "+err.Error())
		}

		// Tạo một map để kiểm tra nhanh hơn
		mediaMap := make(map[uint]bool)
		for _, id := range mediaIDs {
			mediaMap[id] = true
		}

		// Kiểm tra xem có media nào trong FilesRemove không tồn tại trong mediaMap
		var notFoundMedia []uint
		for _, fileID := range s.FilesRemove {
			if !mediaMap[fileID] {
				notFoundMedia = append(notFoundMedia, fileID)
			}
		}

		// Nếu có file không tồn tại, trả về lỗi
		if len(notFoundMedia) > 0 {
			return responses.NewErrorResponse(fiber.StatusBadRequest, fmt.Sprintf("Some media %v do not exist in product, cannot remove", notFoundMedia))
		}
	}

	// Nếu không có lỗi, trả về nil
	return nil
}

// Change validate data to instance
func (s *ProductUpdateSerializer) Update(instance *models.Product) *responses.ErrorResponse {
	// Sao chép dữ liệu dạng PATCH
	if err := copier.CopyWithOption(instance, s, copier.Option{IgnoreEmpty: true}); err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to copy data: "+err.Error())
	}

	// Lưu thay đổi vào database
	// dùng omit để loại bỏ 'layer magic tự tạo category của gorm' :))
	if err := db.DB.Omit("Categories").Save(instance).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to update: "+err.Error())
	}

	// Bulk insert product_categories nếu có category
	lenCategories := len(s.Categories)
	if lenCategories > 0 {
		productCategories := make([]models.ProductCategory, lenCategories)
		for i, categoryID := range s.Categories {
			parsedID, _ := uuid.Parse(categoryID)
			productCategories[i] = models.ProductCategory{
				ProductID:  instance.ID,
				CategoryID: parsedID,
			}
		}
		// Batch insert
		db.DB.Create(&productCategories)
	}

	// Delete category trong category remove
	lenCategoriesRemove := len(s.CategoriesRemove)
	if lenCategoriesRemove > 0 {
		db.DB.Delete(&models.ProductCategory{}, "product_id = ? AND category_id IN ?", instance.ID, s.CategoriesRemove)
	}

	// Delete media
	services.BulkUpdateMedia(models.MediaStatusUpdated, s.FilesRemove)

	// Trả về nil nếu thành công
	return nil
}

// ProductQuerySerializer định nghĩa các tham số truy vấn để lọc và sắp xếp product
type ProductQuerySerializer struct {
	OrderID         string  `query:"order_id" json:"order_id,omitempty"`
	UserID          string  `query:"user_id" json:"user_id,omitempty"`
	RelatedType     string  `query:"related_type" json:"related_type,omitempty"`
	PaymentMethod   string  `query:"payment_method" json:"payment_method,omitempty"`
	PaymentStatus   string  `query:"payment_status" json:"payment_status,omitempty" default:"success"`
	Page            int32   `query:"page" json:"page" default:"1"`
	PageSize        int32   `query:"page_size" json:"page_size" default:"10"`
	MaxPrice        float64 `query:"max_price" json:"max_price,omitempty"`
	MinPrice        float64 `query:"min_price" json:"min_price,omitempty"`
	EndPaymentDay   string  `query:"end_payment_day" json:"end_payment_day,omitempty"`     // YYYY-MM-DD
	StartPaymentDay string  `query:"start_payment_day" json:"start_payment_day,omitempty"` // YYYY-MM-DD
	PaymentDayOrder string  `query:"payment_day_order" json:"payment_day_order,omitempty"` // asc hoặc desc
	PriceOrder      string  `query:"price_order" json:"price_order,omitempty"`             // asc hoặc desc
}

func (s *ProductQuerySerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse body to struct
	if err := c.QueryParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid pagination parameters: "+err.Error())
	}
	// Basic validation với go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}
	// validation cho order
	if s.PaymentDayOrder != "" && s.PaymentDayOrder != "asc" && s.PaymentDayOrder != "desc" {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "payment_day_order must be 'asc' or 'desc'")
	}
	if s.PriceOrder != "" && s.PriceOrder != "asc" && s.PriceOrder != "desc" {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "price_order must be 'asc' or 'desc'")
	}
	// Validation tùy chỉnh cho ngày tháng
	if s.StartPaymentDay != "" {
		startTime, err := time.Parse("2006-01-02", s.StartPaymentDay)
		if err != nil {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid start_payment_day format: must be YYYY-MM-DD")
		}
		if s.EndPaymentDay != "" {
			endTime, err := time.Parse("2006-01-02", s.EndPaymentDay)
			if err != nil {
				return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid end_payment_day format: must be YYYY-MM-DD")
			}
			if startTime.After(endTime) {
				return responses.NewErrorResponse(fiber.StatusBadRequest, "start_payment_day must not be after end_payment_day")
			}
		}
	} else if s.EndPaymentDay != "" {
		if _, err := time.Parse("2006-01-02", s.EndPaymentDay); err != nil {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid end_payment_day format: must be YYYY-MM-DD")
		}
	}

	// Validation tùy chỉnh cho MinPrice và MaxPrice
	if s.MinPrice < 0 {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "min_price must be greater than 0")
	}
	if s.MaxPrice < 0 {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "max_price must be greater than 0")
	}
	if s.MinPrice > 0 && s.MaxPrice > 0 && s.MinPrice > s.MaxPrice {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "min_price must not be greater than max_price")
	}
	return nil
}

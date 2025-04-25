package handlers

import (
	"errors"
	"fmt"
	"product-service/internal/db"
	grpcclient "product-service/internal/grpc_client"
	"product-service/internal/models"
	"product-service/internal/responses"
	"product-service/internal/serializers"
	"product-service/internal/services"
	utils_system "product-service/internal/utils"
	"sync"

	"product-service/pagination"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	pb_order "github.com/hoang3ber123/proto-golang/recommend"
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
	priceMin := c.Query("price_min")              // Giá tối thiểu
	priceMax := c.Query("price_max")              // Giá tối đa
	createdAtMin := c.Query("created_at_min")     // Format: "YYYY-MM-DD"
	createdAtMax := c.Query("created_at_max")     // Format: "YYYY-MM-DD"
	categoryIDsQueries := c.Query("category_ids") //Format: "id1,id2"
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
	var categoryIDs []string
	if categoryIDsQueries != "" {
		categoryIDs = strings.Split(categoryIDsQueries, ",")
		if len(categoryIDs) > 0 {
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
	var result interface{}
	if products != nil {
		result = serializers.ProductListResponse(&products)
	}

	// Lưu lại lịch sử tìm kiếm theo categories
	if len(categoryIDs) > 0 {
		userInterface := c.Locals("user")
		if userInterface != nil {
			user, ok := userInterface.(*models.User)
			if ok {
				go services.SaveSearchCategoryProduct(categoryIDs, user)
			}
		}
	}

	return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
		"pagination": paginator,
		"result":     result,
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
	// Lưu lượt click
	userInterface := c.Locals("user")
	if userInterface != nil {
		user, ok := userInterface.(*models.User)
		if ok {
			go services.SaveClickProduct(instance.ID, user)
		}
	}
	return responses.NewSuccessResponse(fiber.StatusOK, serializers.ProductDetailResponse(&instance)).Send(c)
}

// Lấy detail product employee
func ProductDetailEmployee(c *fiber.Ctx) error {
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
	return responses.NewSuccessResponse(fiber.StatusOK, serializers.ProductDetailEmployeeResponse(&instance)).Send(c)
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

func ProductDownload(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Product ID is required").Send(c)
	}
	type ProductWithFile struct {
		Product models.Product `gorm:"embedded"`
		File    string         `gorm:"column:file"`
	}
	// Struct tạm để lưu kết quả
	var result ProductWithFile

	// Xử lý lỗi
	if err := db.DB.Model(&models.Product{}).
		Select("products.*, media.file").
		Joins("LEFT JOIN media ON media.related_id = products.id AND media.related_type = ? AND media.status = 'using' AND media.file_type = 'download_file'", "products").
		Where("products.id = ?", id).
		First(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return responses.NewErrorResponse(fiber.StatusNotFound, "Product not found").Send(c)
		}
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+err.Error()).Send(c)
	}
	fmt.Println("object:", result)
	// Kiểm tra file_directory
	if result.File == "" {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: file is empty").Send(c)
	}

	// Kiểm tra grpc xem khách hàng mua sản phẩm này chưa? (giữ nguyên logic của bạn)
	// Check employee authenticated
	user := c.Locals("user").(*models.User)
	userID := user.ID.String()
	relatedID := id
	relatedType := "products"
	isBought, err := grpcclient.CheckBoughtRequest(userID, relatedID, relatedType)
	if err != nil {
		return err.Send(c)
	}
	// nếu chưa tải
	if !isBought {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "You did not buy this product.").Send(c)
	}
	// Tạo link tải tạm
	temp_url := utils_system.GenerateTempURL(result.File)

	// Trả về response thành công
	return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
		"download_link": temp_url,
	}).Send(c)
}

// danh sách product đã mua
func ProductFromOrder(c *fiber.Ctx) error {
	serializer := new(serializers.ProductQuerySerializer)
	//  Nếu có lỗi validation, return ngay lập tức
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}
	// lấy user
	user := c.Locals("user").(*models.User)

	serializer.UserID = user.ID.String()
	serializer.RelatedType = "products"

	// Gọi gRPC để lấy danh sách product ID từ order service
	productIDs, pagination, errResp := grpcclient.GetProductIDs(serializer)
	if errResp != nil {
		return c.Status(errResp.StatusCode).JSON(fiber.Map{"error": errResp.Message})
	}

	// Nếu không có product ID nào, trả về kết quả rỗng
	if len(productIDs) == 0 {
		return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
			"pagination": pagination,
			"result":     []models.Product{},
		}).Send(c)
	}

	// Truy vấn danh sách sản phẩm có ID nằm trong productIDs
	var products []models.Product
	if err := db.DB.Where("id IN (?)", productIDs).Find(&products).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+err.Error()).Send(c)
	}

	// Nếu không tìm thấy sản phẩm nào
	if len(products) == 0 {
		return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
			"pagination": pagination,
			"result":     []models.Product{},
		}).Send(c)
	}

	// Chuyển đổi danh sách sản phẩm thành response
	result := serializers.ProductListResponse(&products)

	// Trả về response thành công
	return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
		"pagination": pagination,
		"result":     result,
	}).Send(c)
}

// lấy danh sách product gợi ý
func ProductRecommend(c *fiber.Ctx) error {
	// sử dụng wait group gọi đồng thời các hàm
	var wg sync.WaitGroup
	var clickedProducts []services.ClickedProduct
	var viewProducts []services.ProductSearchCount
	var allProducts []services.ProductResponse
	var boughtProducts []string
	var errClickedProducts, errViewProducts, errGetAllProductInDB error
	var errBoughtProducts *responses.ErrorResponse
	// số lượng gorountine trong group
	wg.Add(4)
	// lấy danh sách lịch sử xem sản phẩm
	go func() {
		// đảm bảo rời thoát group khi thực hiện xong
		defer wg.Done()
		clickedProducts, errClickedProducts = services.GetClickedProductIDs(c)
	}()
	// lấy danh sách lịch sử tìm kiếm sản phẩm
	go func() {
		defer wg.Done()
		viewProducts, errViewProducts = services.GetProductSearchCounts(c)
	}()
	// lấy danh sách product đã mua
	go func() {
		defer wg.Done()
		boughtProducts, errBoughtProducts = grpcclient.GetAllProductIDs(c, "products")
	}()
	// Lấy tất cả product trong csdl
	go func() {
		defer wg.Done()
		allProducts, errGetAllProductInDB = services.GetAllProductInDatabase()
	}()
	// đợi các hàm thực hiện xong
	wg.Wait()

	if errViewProducts != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "error when errViewProducts:"+errViewProducts.Error()).Send(c)
	}
	if errClickedProducts != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "error when errClickedProducts: "+errClickedProducts.Error()).Send(c)
	}
	if errGetAllProductInDB != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "error when errGetAllProductInDB: "+errGetAllProductInDB.Error()).Send(c)
	}
	if errBoughtProducts != nil {
		return errBoughtProducts.Send(c)
	}
	// duyệt mảng maping dữ liệu
	clickDetails := make([]*pb_order.ClickDetail, len(clickedProducts))
	for i, cp := range clickedProducts {
		clickDetails[i] = &pb_order.ClickDetail{
			ProductId: cp.ProductID,
			ClickTime: int32(cp.ClickTime), // Chuyển đổi kiểu dữ liệu nếu cần
		}
	}
	viewProductDetails := make([]*pb_order.ViewProduct, len(viewProducts))
	for i, vp := range viewProducts {
		viewProductDetails[i] = &pb_order.ViewProduct{
			ProductId: vp.ProductID.String(), // Chuyển đổi UUID thành string
			ViewTime:  int32(vp.ViewTime),    // Chuyển đổi kiểu dữ liệu nếu cần
		}
	}
	// Chuyển Categories từ string sang []string
	allProductProto := make([]*pb_order.ProductRecommend, len(allProducts))
	for i, p := range allProducts {
		allProductProto[i] = &pb_order.ProductRecommend{}
		allProductProto[i].Id = p.ID
		allProductProto[i].Title = p.Title
		allProductProto[i].CreatedAt = p.CreatedAt.Format("2006-01-06")
		allProductProto[i].Pricing = p.Price

		if p.Categories != "" {
			allProductProto[i].Categories = strings.Split(p.Categories, ",")
		} else {
			allProductProto[i].Categories = []string{}
		}
	}
	// Gọi hàm GetRecommendProductIDs để lấy danh sách sản phẩm gợi ý
	recommendProductIDs, err := grpcclient.GetRecommendProductIDs(allProductProto, clickDetails, viewProductDetails, boughtProducts)
	if err != nil {
		return err.Send(c)
	}

	// lấy danh sách product
	var recommendProduct []models.Product
	if len(recommendProductIDs) > 0 {
		if err := db.DB.Model(models.Product{}).
			Where("id IN ?", recommendProductIDs).
			Find(&recommendProduct).Error; err != nil {
			return responses.NewErrorResponse(fiber.StatusInternalServerError, err.Error()).Send(c)
		}
	}
	// serializer mảng trả về
	var result interface{}
	if recommendProduct != nil {
		result = serializers.ProductListResponse(&recommendProduct)
	}
	//  Trả về response thành công
	return responses.NewSuccessResponse(fiber.StatusOK, result).Send(c)
}

package handlers

import (
	"product-service/internal/db"
	grpcclient "product-service/internal/grpc_client"
	"product-service/internal/models"
	"product-service/internal/responses"
	"product-service/internal/serializers"

	"github.com/gofiber/fiber/v2"
)

func CartProductAdd(c *fiber.Ctx) error {
	// Xử lý tạo category nếu xác thực thành công
	serializer := new(serializers.CartProductAddSerializer)
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}

	// Tìm hoặc tạo giỏ hàng
	user := c.Locals("user").(*models.User)
	var cart models.Cart
	if err := db.DB.Where("user_id = ?", user.ID).FirstOrCreate(&cart, models.Cart{UserID: user.ID}).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to get or create cart"+err.Error()).Send(c)
	}

	// Kiểm tra product tồn tại
	var productExists bool
	if err := db.DB.Raw("SELECT EXISTS (SELECT 1 FROM products WHERE id = ?)", serializer.ID).Scan(&productExists).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error checking product ID: "+err.Error()).Send(c)
	}
	if !productExists {
		return responses.NewErrorResponse(fiber.StatusNotFound, "Product does not exist").Send(c)
	}

	// kiểm tra xem đã mua sản phẩm chưa ?
	isBought, err := grpcclient.CheckBoughtRequest(user.ID.String(), serializer.ID.String(), "products")
	if err != nil {
		return err.Send(c)
	}
	// Nếu đã tải
	if isBought {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "You've already bought this product.").Send(c)
	}

	// Kiểm tra xem product đã có trong giỏ hàng chưa
	var productInCartExists bool
	if err := db.DB.Raw(
		"SELECT EXISTS (SELECT 1 FROM cart_items WHERE cart_id = ? AND related_id = ? AND related_type = 'products')",
		cart.ID, serializer.ID,
	).Scan(&productInCartExists).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error checking cart item: "+err.Error()).Send(c)
	}
	if productInCartExists {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Product is already in cart").Send(c)
	}

	// Thêm item vào cart
	cartItem := models.CartItem{
		CartID:      cart.ID.String(),
		RelatedID:   serializer.ID,
		RelatedType: "products",
	}
	if err := db.DB.Create(&cartItem).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to add item to cart").Send(c)
	}

	return responses.NewSuccessResponse(fiber.StatusOK, "Adding product to cart successfully").Send(c)
}

func CartProductRemove(c *fiber.Ctx) error {
	// Tạo serializer và kiểm tra tính hợp lệ
	serializer := new(serializers.CartProductRemoveSerializer)
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}

	// Lấy user từ context (giả định đã có middleware auth)
	user := c.Locals("user").(*models.User)

	// Tìm hoặc tạo giỏ hàng cho người dùng
	var cart models.Cart
	if err := db.DB.Where("user_id = ?", user.ID).FirstOrCreate(&cart, models.Cart{UserID: user.ID}).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to get or create cart: "+err.Error()).Send(c)
	}

	// Kiểm tra xem sản phẩm đã có trong giỏ hàng chưa
	var productInCartExists bool
	if err := db.DB.Raw(
		"SELECT EXISTS (SELECT 1 FROM cart_items WHERE cart_id = ? AND related_id IN ? AND related_type = 'products')",
		cart.ID, serializer.IDs,
	).Scan(&productInCartExists).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error checking cart item: "+err.Error()).Send(c)
	}

	// Nếu sản phẩm chưa có trong giỏ hàng, trả về lỗi
	if !productInCartExists {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Product is not in cart").Send(c)
	}

	// Xóa item khỏi giỏ hàng
	if err := db.DB.Where("cart_id = ? AND related_id IN ? AND related_type = 'products'", cart.ID, serializer.IDs).Delete(&models.CartItem{}).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to remove item from cart: "+err.Error()).Send(c)
	}

	// Trả về kết quả thành công
	return responses.NewSuccessResponse(fiber.StatusOK, "Product removed from cart successfully").Send(c)
}

func CartProductDetail(c *fiber.Ctx) error {
	// Lấy user từ context (giả định đã có middleware auth)
	user := c.Locals("user").(*models.User)

	// Lấy danh sách các product trong cart của user
	var products []models.Product
	if err := db.DB.
		Table("cart_items").
		Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Joins("JOIN products ON products.id = cart_items.related_id").
		Where("carts.user_id = ? AND cart_items.related_type = 'products'", user.ID).
		Select("products.*").
		Find(&products).Error; err != nil {
		// Trả về lỗi nếu có lỗi trong quá trình truy vấn
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to get products from cart: "+err.Error()).Send(c)
	}

	// Trả về kết quả thành công với danh sách sản phẩm
	var result interface{}
	if products != nil {
		result = serializers.ProductListResponse(&products)
	}
	return responses.NewSuccessResponse(fiber.StatusOK, result).Send(c)
}

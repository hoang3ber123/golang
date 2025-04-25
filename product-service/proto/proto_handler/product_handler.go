package protohandler

import (
	"context"
	"product-service/config"
	"product-service/internal/db"
	"product-service/internal/models"

	product_proto "github.com/hoang3ber123/proto-golang/product"
	"gorm.io/gorm"
)

// ProductServiceServer triển khai interface từ proto
type ProductServiceServer struct {
	product_proto.UnimplementedProductServiceServer
}

// Xác thực token từ Product Service
func (s *ProductServiceServer) GetProductsInfo(ctx context.Context, req *product_proto.GetProductsInfoRequest) (*product_proto.GetProductsInfoResponse, error) {
	var response product_proto.GetProductsInfoResponse
	var products []*product_proto.Product
	// Tách dữ liệu ra các mảng id khác nhau dựa vào relatedType
	var productIDs []string
	for _, p := range req.Products {
		if p.RelatedType == "products" {
			productIDs = append(productIDs, p.RelatedId)
		}
	}
	baseUrl := config.Config.VstorageBaseURL + "/"
	err := db.DB.Raw(`
            SELECT 
                p.id, 
                p.title, 
                p.slug, 
                'products' as related_type, 
                p.price,
                (SELECT m.file 
                 FROM media m 
                 WHERE m.related_id = p.id 
                 AND m.related_type = 'products' 
                 AND m.file_type = 'image' 
                 AND m.status = 'using' 
                 ORDER BY m.created_at ASC 
                 LIMIT 1) as image
            FROM products p
            WHERE p.id IN ?
        `, productIDs).Scan(&products).Error

	if err != nil {
		response.StatusCode = 500
		response.Error = "Failed to query products: " + err.Error()
		return &response, err
	}
	// duyệt mảng product thêm base url
	for _, p := range products {
		if p.Image != "" {
			p.Image = baseUrl + p.Image
		}
	}
	// Trả về response
	response.Products = products
	response.StatusCode = 200
	response.Error = ""
	return &response, nil
}

func (s *ProductServiceServer) ClearCartAfterCheckout(ctx context.Context, req *product_proto.ClearCartAfterCheckoutRequest) (*product_proto.ClearCartAfterCheckoutResponse, error) {
	var response product_proto.ClearCartAfterCheckoutResponse
	// Tách dữ liệu ra các mảng id khác nhau dựa vào relatedType
	var productIDs []string
	for _, p := range req.Products {
		if p.RelatedType == "products" {
			productIDs = append(productIDs, p.RelatedId)
		}
	}

	// Nếu không có productIDs, vẫn trả về thành công
	if len(productIDs) == 0 {
		response.StatusCode = 200
		response.Error = ""
		return &response, nil
	}

	// Tìm Cart của user
	var cart models.Cart
	err := db.DB.Model(&models.Cart{}).
		Where("user_id = ?", req.User).
		First(&cart).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Không tìm thấy Cart, vẫn trả về thành công (vì không có gì để xóa)
			response.StatusCode = 200
			response.Error = ""
			return &response, nil
		}
		// Lỗi khác (ví dụ: database error)
		response.StatusCode = 500
		response.Error = "Failed to find cart: " + err.Error()
		return &response, err
	}

	// Thực hiện truy vấn xóa CartItem
	err = db.DB.
		Where("cart_id = ? AND related_id IN ? AND related_type = 'products'", cart.ID, productIDs).
		Delete(&models.CartItem{}).Error
	if err != nil {
		response.StatusCode = 500
		response.Error = "Failed to clear cart: " + err.Error()
		return &response, err
	}

	// Trả về response
	response.StatusCode = 200
	response.Error = ""
	return &response, nil
}

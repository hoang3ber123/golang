package protohandler

import (
	"context"
	"fmt"
	"log"
	"product-service/config"
	"product-service/internal/db"
	"strings"

	product_proto "github.com/hoang3ber123/proto-golang/product"
)

// ProductServiceServer triển khai interface từ proto
type ProductServiceServer struct {
	product_proto.UnimplementedProductServiceServer
}

// Xác thực token từ Product Service
func (s *ProductServiceServer) GetProductsInCart(ctx context.Context, req *product_proto.GetProductsInCartRequest) (*product_proto.GetProductsInCartResponse, error) {
	var response product_proto.GetProductsInCartResponse
	allowRelatedType := map[string]bool{
		"products": true,
		"services": true,
	}
	// Kiểm tra có trong relatetype không
	for _, item := range req.Cart {
		if !allowRelatedType[item.Key] {
			{
				log.Println("Warning: Error allow realated type")
				return &product_proto.GetProductsInCartResponse{
					Products:   nil,
					Error:      "Invalid related type: " + item.Key,
					StatusCode: 400,
				}, nil
			}
		}
	}
	// Tạo danh sách câu truy vấn có thêm cột related_type
	queries := make([]string, len(req.Cart))
	for inx, item := range req.Cart {
		// Nối các ID thành chuỗi '123', '1234', '1235'
		idStr := "'" + strings.Join(item.Values, "', '") + "'"
		queries[inx] = fmt.Sprintf(
			"SELECT id, title, price, '%s' AS related_type FROM %s WHERE id IN (%s)",
			item.Key, item.Key, idStr,
		)
	}
	// Hợp nhất truy vấn bằng UNION ALL
	finalQuery := strings.Join(queries, " UNION ALL ")
	// Đổ product lấy từ các bảng vào một bảng tạm
	cteQuery := `
	WITH temp_products AS (
		%s
	),
	ranked_media AS (
		SELECT m.*, 
		       ROW_NUMBER() OVER (PARTITION BY m.related_id, m.related_type ORDER BY m.created_at ASC) AS rnk
		FROM media m
		WHERE m.status = 'using' AND m.file_type = 'image'
	)
	SELECT tp.id as id, tp.title as title, CONCAT('%s', rm.file) as image, tp.related_type as related_type, tp.price as price
	FROM temp_products tp
	LEFT JOIN ranked_media rm 
	    ON tp.id = rm.related_id AND tp.related_type = rm.related_type
	WHERE rm.rnk = 1 OR rm.rnk IS NULL;
`

	// Chạy truy vấn với GORM
	baseURL := fmt.Sprintf("%s/", config.Config.VstorageBaseURL)
	products := []*product_proto.Product{}
	db.DB.Raw(fmt.Sprintf(cteQuery, finalQuery, baseURL)).Scan(&products)
	// Trả về response
	response.Products = products
	response.StatusCode = 200
	response.Error = ""
	return &response, nil
}

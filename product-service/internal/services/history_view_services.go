package services

import (
	"errors"
	"fmt"
	"product-service/internal/db"
	"product-service/internal/models"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// lưu lượt xem detail product
func SaveClickProduct(productID uuid.UUID, ctx *fiber.Ctx) {
	userInterface := ctx.Locals("user")
	if userInterface == nil {
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return
	}

	var history models.HistoryView

	// Kiểm tra xem bản ghi có tồn tại không
	err := db.DB.Where("user_id = ? AND related_id = ? AND realated_type = ?", user.ID, productID, "products").First(&history).Error

	if err != nil {
		// Nếu không tìm thấy (bản ghi chưa tồn tại), tạo mới với ClickTime = 1
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newHistory := models.HistoryView{
				UserID:       user.ID,
				RelatedID:    productID,
				RealatedType: "products",
				ClickTime:    1,
			}
			db.DB.Create(&newHistory)
		} else {
			// Nếu có lỗi khác, có thể log hoặc xử lý theo nhu cầu
			fmt.Println("Error when query HistoryView:", err)
		}
	} else {
		// Nếu đã tồn tại, tăng ClickTime lên 1
		db.DB.Model(&history).UpdateColumn("click_time", gorm.Expr("click_time + ?", 1))
	}
}

// lưu lượt search sản phẩm bằng product
func SaveSearchCategoryProduct(categoriesID []string, ctx *fiber.Ctx) {
	userInterface := ctx.Locals("user")
	if userInterface == nil {
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return
	}
	// truy vấn lấy title của danh sách category
	var categoriesTitle []string
	db.DB.Model(models.Category{}).
		Where("id IN ?", categoriesID).
		Select("title").
		Find(&categoriesTitle)
	// join lại với '|' ngăn cách giữa các cate
	categories := strings.Join(categoriesTitle, "|")
	if categories != "" {
		action := models.HistorySearch{UserID: user.ID, Categories: categories}
		db.DB.Create(&action)
	}
}

type ClickedProduct struct {
	ProductID string `json:"product_id"`
	ClickTime int    `json:"click_time"`
}

// lấy danh sách click sản phẩm của user
func GetClickedProductIDs(ctx *fiber.Ctx) ([]ClickedProduct, error) {
	// Lấy thông tin user từ context
	userInterface := ctx.Locals("user")
	if userInterface == nil {
		return nil, errors.New("user not found")
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return nil, errors.New("invalid user data")
	}

	// Tạo slice chứa kết quả
	var clickedProducts []ClickedProduct

	// Truy vấn dữ liệu vào struct
	err := db.DB.Table("history_views").
		Where("user_id = ? AND realated_type = ?", user.ID, "products").
		Select("related_id as product_id, click_time").
		Find(&clickedProducts).Error

	if err != nil {
		return nil, err
	}

	return clickedProducts, nil
}

type ProductSearchCount struct {
	ProductID uuid.UUID `json:"product_id"`
	ViewTime  int       `json:"view_time"`
}

func GetProductSearchCounts(ctx *fiber.Ctx) ([]ProductSearchCount, error) {
	userInterface := ctx.Locals("user")
	if userInterface == nil {
		return nil, nil
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return nil, nil
	}

	var results []ProductSearchCount

	tx := db.DB.Begin() // Bắt đầu transaction
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Bước 1: Tạo bảng tạm xem lịch sử search category
	err := tx.Exec(`
		CREATE TEMPORARY TABLE temp_category_search_counts AS 
		SELECT 
			categories.id AS category_id, 
			COUNT(history_searches.id) AS search_count 
		FROM categories
		LEFT JOIN history_searches ON FIND_IN_SET(categories.title, history_searches.categories) > 0
		WHERE history_searches.user_id = ?
		GROUP BY categories.id;
	`, user.ID).Error

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Bước 2: Join bảng tạm với product_categories để đếm tổng lượt tìm kiếm sản phẩm dựa trên lượt xem category
	err = tx.Table("product_categories").
		Select("product_categories.product_id, COALESCE(SUM(temp_category_search_counts.search_count), 0) AS view_time").
		Joins("LEFT JOIN temp_category_search_counts ON product_categories.category_id = temp_category_search_counts.category_id").
		Group("product_categories.product_id").
		Scan(&results).Error

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Xóa bảng tạm thủ công
	_ = tx.Exec("DROP TEMPORARY TABLE IF EXISTS temp_category_search_counts").Error

	tx.Commit() // Hoàn thành transaction

	return results, nil
}

type ProductResponse struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	CreatedAt  time.Time `json:"created_at"`
	Price      float64   `json:"price"`
	Categories string    `json:"categories"`
}

func GetAllProductInDatabase() ([]ProductResponse, error) {
	var products []ProductResponse

	// Preload Categories để lấy thông tin category liên quan
	if err := db.DB.Table("products AS p").
		Joins("LEFT JOIN product_categories AS pc ON p.id = pc.product_id").
		Joins("LEFT JOIN categories AS c ON c.id = pc.category_id").
		Group("p.id").
		Select("p.id AS id, p.title AS title, p.created_at AS created_at, p.price AS price, GROUP_CONCAT(c.title) as categories").
		Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}
